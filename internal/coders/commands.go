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

	"github.com/coding-hui/ai-terminal/internal/cli/commit"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

var supportCommands = map[string]func(context.Context, ...string) tea.Msg{}

func getSupportedCommands() []string {
	commands := make([]string, 0, len(supportCommands))
	for cmd := range supportCommands {
		commands = append(commands, cmd)
	}
	return commands
}

type command struct {
	coder  *AutoCoder
	editor *EditBlockCoder
}

func newCommand(coder *AutoCoder) *command {
	editor := NewEditBlockCoder(coder, fences[0])
	cmds := &command{coder: coder, editor: editor}
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
	supportCommands["/commit"] = c.commit
	supportCommands["/undo"] = c.undo
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

	_, err = c.coder.llmEngine.ChatStream(context.Background(), messages)
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

	var matchedFiles = make([]string, 0, len(files))

	for _, pattern := range files {
		matches, err := filepath.Glob(filepath.Join(c.coder.codeBasePath, pattern))
		if err != nil {
			return c.coder.Error(err)
		}

		for _, filePath := range matches {
			fileExists, err := fileutil.FileExists(filePath)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return c.coder.Error(err)
				}
			}

			if fileExists {
				matchedFiles = append(matchedFiles, filePath)
			} else {
				return c.coder.Errorf("File [%s] does not exist", filePath)
			}
		}
	}

	for _, filePath := range matchedFiles {
		absPath, err := absFilePath(c.coder.codeBasePath, filePath)
		if err != nil {
			return c.coder.Error(err)
		}

		if _, ok := c.coder.absFileNames[absPath]; ok {
			return c.coder.Errorf("File [%s] already exists", filePath)
		}

		c.coder.absFileNames[absPath] = struct{}{}

		c.coder.Successf("Added file [%s] to chat context", filePath)
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
		c.coder.Tracef("%d. %s", no, relPath)
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
	for _, pattern := range files {
		matches, err := filepath.Glob(filepath.Join(c.coder.codeBasePath, pattern))
		if err != nil {
			return c.coder.Error(err)
		}

		for _, filePath := range matches {
			abs, err := absFilePath(c.coder.codeBasePath, filePath)
			if err != nil {
				return c.coder.Error(err)
			}
			if _, ok := c.coder.absFileNames[abs]; ok {
				deleteCount++
				delete(c.coder.absFileNames, abs)
				c.coder.Successf("Deleted file [%s]", filePath)
			}
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

	openFence, closeFence := c.editor.UpdateCodeFences(ctx, addedFiles)

	c.coder.Infof("Selected coder block fences %s %s", openFence, closeFence)

	userInput := strings.Join(args, " ")
	messages, err := c.editor.FormatMessages(map[string]any{
		userQuestionKey: userInput,
		addedFilesKey:   addedFiles,
		openFenceKey:    openFence,
		closeFenceKey:   closeFence,
		lazyPromptKey:   lazyPrompt,
	})
	if err != nil {
		return c.coder.Error(err)
	}

	err = c.editor.Execute(ctx, messages)
	if err != nil {
		return c.coder.Error(err)
	}

	c.coder.Done()

	return nil
}

func (c *command) undo(_ context.Context, _ ...string) tea.Msg {
	if err := c.coder.gitRepo.RollbackLastCommit(); err != nil {
		return c.coder.Errorf("Failed to rollback last commit: %v", err)
	}
	c.coder.Successf("Successfully rolled back the last commit")
	defer c.coder.Done()

	return nil
}

func (c *command) commit(ctx context.Context, _ ...string) tea.Msg {
	// Get the list of files that were modified by the coding command
	modifiedFiles, err := c.editor.GetModifiedFiles(ctx)
	if err != nil {
		return c.coder.Error(err)
	}

	// Add the modified files to the Git staging area
	if err := c.coder.gitRepo.AddFiles(modifiedFiles); err != nil {
		return c.coder.Errorf("Failed to add files to Git: %v", err)
	}

	c.coder.Loading("Committing code changes")

	// Execute the commit command
	ioStreams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	commitCmd := commit.NewOptions(true, modifiedFiles, ioStreams)
	if err := commitCmd.AutoCommit(nil, nil); err != nil {
		return c.coder.Errorf("Failed to execute commit command: %v", err)
	}

	c.coder.Successf("Code submitted successfully")

	defer c.coder.Done()

	return nil
}

func (c *command) awaitChatCompleted() tea.Cmd {
	return func() tea.Msg {
		output := <-c.coder.llmEngine.GetChannel()
		c.coder.state.buffer += output.GetContent()
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
