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

var supportCommands = map[string]func(...string) tea.Msg{}

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
}

func (c *command) isCommand(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, "!") || strings.HasPrefix(input, "/")
}

func (c *command) run(input string) tea.Cmd {
	return func() tea.Msg {
		cmd, args := extractCmdArgs(input)

		cmdFunc, ok := supportCommands[cmd]
		if !ok {
			return c.coder.Errorf("Invalid command %s", cmd)
		}

		// do run
		return cmdFunc(args...)
	}
}

func (c *command) askFiles(args ...string) tea.Msg {
	input := strings.Join(args, " ")

	c.coder.Loading(components.spinner.randMsg)

	c.coder.state.buffer = ""
	c.coder.state.command = ""

	messages, err := c.prepareCompletionMessages(input)
	if err != nil {
		return c.coder.Error(err)
	}

	err = c.coder.llmEngine.ChatStream(context.Background(), messages)
	if err != nil {
		return c.coder.Error(err)
	}

	return nil
}

// addFiles Add files to the chat so GPT can edit them or review them in detail
func (c *command) addFiles(files ...string) tea.Msg {
	if len(files) <= 0 {
		return c.coder.Error("Please provide at least one file")
	}

	var (
		err          error
		matchedFiles = make([]string, len(files))
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
		absPath, err := c.absFilePath(f)
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

func (c *command) listFiles(_ ...string) tea.Msg {
	fileCount := len(c.coder.absFileNames)
	c.coder.Infof("Loaded %d files", fileCount)

	for file := range c.coder.absFileNames {
		relPath, err := filepath.Rel(c.coder.codeBasePath, file)
		if err != nil {
			return c.coder.Error(err)
		}
		c.coder.Info(relPath)
	}

	defer c.coder.Done()

	return nil
}

func (c *command) removeFiles(files ...string) tea.Msg {
	if len(files) <= 0 {
		return c.coder.Error("Please provide at least one file")
	}

	deleteCount := 0
	for _, file := range files {
		abs, err := c.absFilePath(file)
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
func (c *command) absFilePath(matchedFile string) (abs string, err error) {
	abs = filepath.Join(c.coder.codeBasePath, matchedFile)
	if filepath.IsAbs(matchedFile) {
		abs, err = filepath.Abs(matchedFile)
		if err != nil {
			return "", err
		}
	}
	return
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

func (c *command) prepareCompletionMessages(userInput string) (messages []llms.MessageContent, err error) {
	addedFileMessages, err := c.getFilesMessages()
	if err != nil {
		return nil, c.coder.Error(err)
	}

	messages = append(messages, addedFileMessages...)
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart(userInput)},
	})

	return
}

func (c *command) getFilesMessages() ([]llms.MessageContent, error) {
	var messagesParts []llms.ContentPart

	messagesParts = append(messagesParts, llms.TextPart(FileContentPrefix))
	if len(c.coder.absFileNames) > 0 {
		for file := range c.coder.absFileNames {
			content, err := c.readFileContent(file)
			if err != nil {
				return nil, c.coder.Error(err)
			}
			messagesParts = append(messagesParts, llms.TextPart(content))
		}
	}

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: messagesParts,
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Ok, any changes I propose will be to those files.")},
		},
	}

	return messages, nil
}

func (c *command) readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	relPath, err := filepath.Rel(c.coder.codeBasePath, filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\n%s%s", relPath, wrapFence(string(content))), nil
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
