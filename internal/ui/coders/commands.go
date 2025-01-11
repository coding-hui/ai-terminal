package coders

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/coding-hui/common/util/fileutil"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/cli/commit"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/prompt"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

var supportCommands = map[string]func(context.Context, ...string) error{}

func getSupportedCommands() []string {
	commands := make([]string, 0, len(supportCommands))
	for cmd := range supportCommands {
		commands = append(commands, cmd)
	}
	return commands
}

type CommandExecutor struct {
	coder  *AutoCoder
	editor *EditBlockCoder
}

func NewCommandExecutor(coder *AutoCoder) *CommandExecutor {
	editor := NewEditBlockCoder(coder, fences[0])
	cmds := &CommandExecutor{coder: coder, editor: editor}
	cmds.registryCmds()
	return cmds
}

func (c *CommandExecutor) registryCmds() {
	supportCommands["add"] = c.addFiles
	supportCommands["list"] = c.listFiles
	supportCommands["remove"] = c.removeFiles
	supportCommands["ask"] = c.askFiles
	supportCommands["drop"] = c.dropFiles
	supportCommands["coding"] = c.coding
	supportCommands["commit"] = c.commit
	supportCommands["undo"] = c.undo
	supportCommands["exit"] = c.exit
}

func (c *CommandExecutor) isCommand(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, "!") || strings.HasPrefix(input, "/")
}

// Executor Execute the command
func (c *CommandExecutor) Executor(input string) {
	if input == "" {
		return
	}

	cmd, args := "ask", []string{input}
	if c.isCommand(input) {
		cmd, args = extractCmdArgs(input)
	}

	fn, ok := supportCommands[cmd]
	if !ok {
		console.RenderError(
			errbook.ErrInvalidArgument,
			"Unknown command: %s. Only support commands: %s. Type / to see all recommended commands.", cmd, strings.Join(getSupportedCommands(), ", "),
		)

		return
	}

	// do executor
	if err := fn(context.Background(), args...); err != nil {
		fmt.Println("ddd")
		console.RenderError(err, "Failed to execute command %s", cmd)
	}
}

// askFiles Ask GPT to edit the files in the chat
func (c *CommandExecutor) askFiles(ctx context.Context, args ...string) error {
	messages, err := c.prepareAskCompletionMessages(strings.Join(args, " "))
	if err != nil {
		return errbook.Wrap("Failed to prepare ask completion messages", err)
	}

	go func() {
		for msg := range c.coder.engine.GetChannel() {
			fmt.Print(msg.GetContent())
		}
	}()

	_, err = c.coder.engine.ChatStream(ctx, messages)
	if err != nil {
		return err
	}

	fmt.Println()

	return nil
}

// addFiles Add files to the chat so GPT can edit them or review them in detail
func (c *CommandExecutor) addFiles(_ context.Context, files ...string) error {
	if len(files) <= 0 {
		return errbook.New("Please provide at least one file")
	}

	var matchedFiles = make([]string, 0, len(files))

	for _, pattern := range files {
		matches, err := filepath.Glob(filepath.Join(c.coder.codeBasePath, pattern))
		if err != nil {
			return errbook.Wrap("Failed to glob files", err)
		}

		if len(matches) <= 0 {
			console.Render("No files matched pattern [%s]", pattern)
			break
		}

		for _, filePath := range matches {
			fileExists, err := fileutil.FileExists(filePath)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return errbook.Wrap("Failed to check if file exists", err)
				}
			}

			if fileExists {
				matchedFiles = append(matchedFiles, filePath)
			} else {
				return errbook.New("File [%s] not found", filePath)
			}
		}
	}

	for _, filePath := range matchedFiles {
		absPath, err := absFilePath(c.coder.codeBasePath, filePath)
		if err != nil {
			return errbook.Wrap("Failed to get abs path", err)
		}

		if _, ok := c.coder.absFileNames[absPath]; ok {
			return errbook.New("File [%s] already exists", absPath)
		}

		c.coder.absFileNames[absPath] = struct{}{}

		console.Render("Added file [%s]", absPath)
	}

	return nil
}

// listFiles List all files that have been added to the chat
func (c *CommandExecutor) listFiles(_ context.Context, _ ...string) error {
	if len(c.coder.absFileNames) <= 0 {
		console.Render("No files added in chat currently")
		return nil
	}
	no := 1
	for file := range c.coder.absFileNames {
		relPath, err := filepath.Rel(c.coder.codeBasePath, file)
		if err != nil {
			return errbook.Wrap("Failed to get relative path", err)
		}
		console.Render("%d.%s", no, relPath)
		no++
	}

	return nil
}

// removeFiles Remove files from the chat so GPT won't edit them or review them in detail
func (c *CommandExecutor) removeFiles(_ context.Context, files ...string) error {
	if len(files) <= 0 {
		return errbook.New("Please provide at least one file")
	}

	deleteCount := 0
	for _, pattern := range files {
		matches, err := filepath.Glob(filepath.Join(c.coder.codeBasePath, pattern))
		if err != nil {
			return errbook.Wrap("Failed to glob files", err)
		}

		for _, filePath := range matches {
			abs, err := absFilePath(c.coder.codeBasePath, filePath)
			if err != nil {
				return errbook.Wrap("Failed to get abs path", err)
			}
			if _, ok := c.coder.absFileNames[abs]; ok {
				deleteCount++
				delete(c.coder.absFileNames, abs)
				console.Render("Removed file [%s]", abs)
			}
		}
	}

	return nil
}

func (c *CommandExecutor) dropFiles(_ context.Context, _ ...string) error {
	dropCount := len(c.coder.absFileNames)
	c.coder.absFileNames = map[string]struct{}{}
	console.Render("Dropped %d files", dropCount)

	return nil
}

func (c *CommandExecutor) coding(ctx context.Context, args ...string) error {
	addedFiles, err := c.getAddedFileContent()
	if err != nil {
		return err
	}

	openFence, closeFence := c.editor.UpdateCodeFences(ctx, addedFiles)

	console.Render("Selected coder block fences %s %s", openFence, closeFence)

	userInput := strings.Join(args, " ")
	messages, err := c.editor.FormatMessages(map[string]any{
		userQuestionKey: userInput,
		addedFilesKey:   addedFiles,
		openFenceKey:    openFence,
		closeFenceKey:   closeFence,
		lazyPromptKey:   lazyPrompt,
	})
	if err != nil {
		return err
	}

	err = c.editor.Execute(ctx, messages)
	if err != nil {
		return err
	}

	return nil
}

func (c *CommandExecutor) undo(ctx context.Context, _ ...string) error {
	modifiedFiles, err := c.editor.GetModifiedFiles(ctx)
	if err != nil {
		return err
	}
	if len(modifiedFiles) == 0 {
		return errbook.New("There are no file modifications")
	}
	if err := c.coder.repo.RollbackLastCommit(); err != nil {
		return errbook.Wrap("Failed to get modified files", err)
	}

	return nil
}

func (c *CommandExecutor) commit(ctx context.Context, _ ...string) error {
	// Get the list of files that were modified by the coding CommandExecutor
	modifiedFiles, err := c.editor.GetModifiedFiles(ctx)
	if err != nil {
		return err
	}

	// Add the modified files to the Git staging area
	if err := c.coder.repo.AddFiles(modifiedFiles); err != nil {
		return errbook.Wrap("Failed to add files to Git", err)
	}

	// Execute the commit CommandExecutor
	ioStreams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	commitCmd := commit.New(
		commit.WithNoConfirm(true),
		commit.WithFilesToAdd(modifiedFiles),
		commit.WithIOStreams(ioStreams),
		commit.WithConfig(c.coder.cfg),
		commit.WithCommitPrefix(c.coder.cfg.AutoCoder.CommitPrefix), // Use configured commit prefix
		commit.WithCommitLang(prompt.DefaultLanguage),
	)
	if err := commitCmd.AutoCommit(nil, nil); err != nil {
		return errbook.Wrap("Failed to commit changes", err)
	}

	return nil
}

func (c *CommandExecutor) exit(_ context.Context, _ ...string) error {
	fmt.Println("Bye!")
	os.Exit(0)

	return nil
}

func (c *CommandExecutor) prepareAskCompletionMessages(userInput string) ([]llms.MessageContent, error) {
	addedFileMessages, err := c.getAddedFileContent()
	if err != nil {
		return nil, err
	}

	messages, err := promptAsk.FormatMessages(map[string]any{
		addedFilesKey:   addedFileMessages,
		userQuestionKey: userInput,
	})
	if err != nil {
		return nil, err
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

func (c *CommandExecutor) getAddedFileContent() (string, error) {
	addedFiles := ""
	if len(c.coder.absFileNames) > 0 {
		for file := range c.coder.absFileNames {
			content, err := c.formatFileContent(file)
			if err != nil {
				return "", err
			}
			addedFiles += content
		}
	}

	return addedFiles, nil
}

func (c *CommandExecutor) formatFileContent(filePath string) (string, error) {
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
	input = strings.TrimPrefix(input, "!")
	input = strings.TrimPrefix(input, "/")
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
