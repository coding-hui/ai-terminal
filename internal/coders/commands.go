package coders

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coding-hui/common/util/fileutil"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

var supportCommands = map[string]func(context.Context, ...string) tea.Msg{}

type command struct {
	coder *AutoCoder
}

func newCommand(coder *AutoCoder) *command {
	cmds := &command{coder: coder}
	cmds.registryCmds()
	return cmds
}

func (c *command) registryCmds() {
	supportCommands["/add"] = c.addFiles
	supportCommands["/list"] = c.listFiles
	supportCommands["/remove"] = c.removeFiles
	supportCommands["/ask"] = c.askFiles
	supportCommands["/drop"] = c.dropFiles
	supportCommands["/coding"] = c.coding
}

func (c *command) isCommand(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, "!") || strings.HasPrefix(input, "/")
}

func (c *command) run(input string) tea.Cmd {
	return func() tea.Msg {
		cmd, args := "/ask", []string{input}
		if c.isCommand(input) {
			cmd, args = extractCmdArgs(input)
		}

		cmdFunc, ok := supportCommands[cmd]
		if !ok {
			return c.coder.Errorf("Invalid command %s", cmd)
		}

		// do run
		return cmdFunc(context.Background(), args...)
	}
}

func (c *command) askFiles(_ context.Context, args ...string) tea.Msg {
	c.coder.Loading(components.spinner.randMsg)

	c.coder.state.buffer = ""

	messages, err := c.prepareAskCompletionMessages(strings.Join(args, " "))
	if err != nil {
		return c.coder.Errorf("Failed to prepare ask completion messages: %v", err)
	}

	err = c.coder.llmEngine.ChatStream(context.Background(), messages)
	if err != nil {
		return c.coder.Errorf("Failed to chat stream: %v", err)
	}

	return nil
}

// addFiles Add files to the chat so GPT can edit them or review them in detail
func (c *command) addFiles(_ context.Context, files ...string) tea.Msg {
	if len(files) <= 0 {
		return c.coder.Error("Please provide at least one file")
	}

	var (
		err          error
		matchedFiles = []string{}
	)

	for _, file := range files {
		filePath := filepath.Join(c.coder.codeBasePath, file)
		if filepath.IsAbs(file) {
			filePath, err = filepath.Abs(file)
			if err != nil {
				return c.coder.Error(err)
			}
		}

		fileExists, err := fileutil.FileExists(filePath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return c.coder.Error(err)
			}
		}

		if fileExists {
			matchedFiles = append(matchedFiles, file)
		} else {
			return c.coder.Errorf("File [%s] does not exist", filePath)
		}
	}

	for _, f := range matchedFiles {
		absPath, err := absFilePath(c.coder.codeBasePath, f)
		if err != nil {
			return c.coder.Error(err)
		}

		if _, ok := c.coder.absFileNames[absPath]; ok {
			return c.coder.Errorf("File [%s] already exists", f)
		}

		c.coder.absFileNames[absPath] = struct{}{}

		c.coder.Successf("Added file [%s] to chat context", f)
	}

	defer c.coder.Done()

	return nil
}

func (c *command) listFiles(_ context.Context, _ ...string) tea.Msg {
	fileCount := len(c.coder.absFileNames)
	c.coder.Infof("Loaded %d files", fileCount)

	no := 1
	for file := range c.coder.absFileNames {
		relPath, err := filepath.Rel(c.coder.codeBasePath, file)
		if err != nil {
			return c.coder.Error(err)
		}
		c.coder.Infof("%d. %s", no, relPath)
		no++
	}

	defer c.coder.Done()

	return nil
}

func (c *command) removeFiles(_ context.Context, files ...string) tea.Msg {
	if len(files) <= 0 {
		return c.coder.Error("Please provide at least one file")
	}

	deleteCount := 0
	for _, file := range files {
		abs, err := absFilePath(c.coder.codeBasePath, file)
		if err != nil {
			return c.coder.Error(err)
		}
		if _, ok := c.coder.absFileNames[abs]; !ok {
			c.coder.Infof("File [%s] does not exist", file)
		} else {
			deleteCount++
			delete(c.coder.absFileNames, abs)
			c.coder.Successf("Deleted file [%s]", file)
		}
	}
	c.coder.Infof("Deleted %d files", deleteCount)

	defer c.coder.Done()

	return nil
}

func (c *command) dropFiles(_ context.Context, _ ...string) tea.Msg {
	dropCount := len(c.coder.absFileNames)
	c.coder.absFileNames = map[string]struct{}{}
	c.coder.Infof("Dropped %d files", dropCount)

	defer c.coder.Done()

	return nil
}

func (c *command) coding(ctx context.Context, args ...string) tea.Msg {
	addedFiles, err := c.getAddedFileContent()
	if err != nil {
		return c.coder.Error(err)
	}

	openFence, closeFence := chooseBestFence(addedFiles)
	c.coder.Infof("Selected coder block fences %s %s", openFence, closeFence)

	editor := NewEditBlockCoder(c.coder, []string{openFence, closeFence})

	userInput := strings.Join(args, " ")
	messages, err := editor.FormatMessages(map[string]any{
		userQuestionKey: userInput,
		addedFilesKey:   addedFiles,
		openFenceKey:    openFence,
		closeFenceKey:   closeFence,
		lazyPromptKey:   lazyPrompt,
	})
	if err != nil {
		return c.coder.Error(err)
	}

	err = editor.Execute(ctx, messages)
	if err != nil {
		return c.coder.Error(err)
	}

	edits, err := editor.GetEdits(ctx)
	if err != nil {
		return c.coder.Error(err)
	}

	if len(edits) <= 0 {
		return c.coder.Errorf("No edits were made")
	}

	c.coder.Infof("Applying %d edits...", len(edits))

	err = editor.ApplyEdits(ctx, edits, true)
	if err != nil {
		return c.coder.Error(err)
	}

	c.coder.Successf("Code editing completed")
	c.coder.Done()

	return nil
}

func (c *command) awaitChatCompleted() tea.Cmd {
	return func() tea.Msg {
		output := <-c.coder.llmEngine.GetChannel()
		c.coder.state.buffer += output.GetContent()
		c.coder.state.querying = !output.IsLast()
		if output.IsLast() {
			c.coder.Done()
		}
		return output
	}
}

func (c *command) prepareAskCompletionMessages(userInput string) ([]llms.MessageContent, error) {
	addedFileMessages, err := c.getAddedFileContent()
	if err != nil {
		return nil, c.coder.Error(err)
	}

	messages, err := promptAsk.FormatMessages(map[string]any{
		addedFilesKey:   addedFileMessages,
		userQuestionKey: userInput,
	})
	if err != nil {
		return nil, c.coder.Error(err)
	}

	ret := make([]llms.MessageContent, 0, len(messages))
	for _, msg := range messages {
		switch msg.GetType() {
		case llms.ChatMessageTypeAI:
			ret = append(ret, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
			})
		case llms.ChatMessageTypeHuman:
			ret = append(ret, llms.MessageContent{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
			})
		case llms.ChatMessageTypeSystem:
			ret = append(ret, llms.MessageContent{
				Role:  llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
			})
		}
	}

	return ret, nil
}

func (c *command) getAddedFileContent() (string, error) {
	addedFiles := ""
	if len(c.coder.absFileNames) > 0 {
		for file := range c.coder.absFileNames {
			content, err := c.formatFileContent(file)
			if err != nil {
				return "", c.coder.Error(err)
			}
			addedFiles += content
		}
	}

	return addedFiles, nil
}

func (c *command) formatFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	relPath, err := filepath.Rel(c.coder.codeBasePath, filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\n%s%s", relPath, wrapFenceWithType(string(content), filePath)), nil
}

func extractCmdArgs(input string) (string, []string) {
	words := strings.Split(strings.TrimSpace(input), " ")
	args := []string{}
	for _, word := range words[1:] {
		if len(word) > 0 {
			args = append(args, word)
		}
	}
	return words[0], args
}

func absFilePath(basePath, matchedFile string) (abs string, err error) {
	abs = filepath.Join(basePath, matchedFile)
	if filepath.IsAbs(matchedFile) {
		abs, err = filepath.Abs(matchedFile)
		if err != nil {
			return "", err
		}
	}
	return
}
