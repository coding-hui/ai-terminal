// Package coders provides command execution capabilities for the AI terminal.
// It handles commands like adding/removing files, asking questions, coding with AI,
// committing changes, and more.
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
	"github.com/coding-hui/ai-terminal/internal/ui/chat"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/rest"
)

// supportCommands maps command names to their handler functions
var supportCommands = map[string]func(context.Context, ...string) error{}

// getSupportedCommands returns a list of all supported command names
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
	supportCommands["help"] = c.help
	supportCommands["diff"] = c.diff
}

// isCommand checks if the input string is a command (starts with ! or /)
func (c *CommandExecutor) isCommand(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, "!") || strings.HasPrefix(input, "/")
}

// Executor handles command execution. It parses the input, validates the command,
// and executes the corresponding handler function.
func (c *CommandExecutor) Executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Handle non-command input
	if !c.isCommand(input) {
		console.Render("Please use a command to interact with the system. Type / to see all available commands.")
		return
	}

	// Handle command execution
	cmd, args := extractCmdArgs(input)
	fn, ok := supportCommands[cmd]
	if !ok {
		console.RenderError(
			errbook.ErrInvalidArgument,
			"Unknown command: %s. Supported commands: %s. Type / to see all recommended commands.",
			cmd, strings.Join(getSupportedCommands(), ", "),
		)
		return
	}

	// Execute the recognized command
	if err := fn(context.Background(), args...); err != nil {
		console.RenderError(err, "Failed to execute command %s", cmd)
	}
}

// askFiles Ask GPT to edit the files in the chat
func (c *CommandExecutor) askFiles(ctx context.Context, args ...string) error {
	messages, err := c.prepareAskCompletionMessages(strings.Join(args, " "))
	if err != nil {
		return errbook.Wrap("Failed to prepare ask completion messages", err)
	}

	chatModel := chat.NewChat(c.coder.cfg,
		chat.WithContext(ctx),
		chat.WithMessages(messages),
		chat.WithEngine(c.coder.engine),
	)

	return chatModel.Run()
}

// addFiles Add files to the chat so GPT can edit them or review them in detail
func (c *CommandExecutor) loadFileContent(path string) (string, error) {
	// Handle remote URLs
	if rest.IsRemoteFile(path) {
		return rest.FetchRemoteContent(path)
	}

	// Handle local files
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errbook.Wrap("Failed to read local file", err)
	}
	return string(content), nil
}

func (c *CommandExecutor) addFiles(_ context.Context, files ...string) error {
	if len(files) <= 0 {
		return errbook.New("Please provide at least one file or URL")
	}

	var matchedFiles = make([]string, 0, len(files))

	for _, pattern := range files {
		// Handle remote URLs directly
		if rest.IsRemoteFile(pattern) {
			matchedFiles = append(matchedFiles, pattern)
			continue
		}

		// Handle local file patterns
		matches, err := filepath.Glob(filepath.Join(c.coder.codeBasePath, pattern))
		if err != nil {
			return errbook.Wrap("Failed to glob files", err)
		}

		if len(matches) <= 0 {
			console.Render("No files matched pattern [%s]", pattern)
			continue
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
	if len(c.coder.absFileNames) == 0 {
		return errbook.New("No files added. Please use /add to add files first")
	}

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

	// Auto-commit if enabled in config
	if c.coder.cfg.AutoCoder.AutoCommit {
		if err := c.commit(ctx); err != nil {
			return errbook.Wrap("Failed to auto-commit changes", err)
		}
	}

	return nil
}

func (c *CommandExecutor) undo(ctx context.Context, _ ...string) error {
	// First check if there are any files in the chat
	if len(c.coder.absFileNames) == 0 {
		return errbook.New("No files added. Please use /add to add files first")
	}

	modifiedFiles, err := c.editor.GetModifiedFiles(ctx)
	if err != nil {
		return err
	}
	if len(modifiedFiles) == 0 {
		return errbook.New("There are no file modifications to undo")
	}

	// Confirm with user before undoing
	if !console.WaitForUserConfirm(console.No, "Are you sure you want to undo the last changes?") {
		console.Render("Undo canceled")
		return nil
	}

	if err := c.coder.repo.RollbackLastCommit(); err != nil {
		return errbook.Wrap("Failed to undo changes", err)
	}

	console.Render("Successfully undone last changes")
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

func (c *CommandExecutor) help(_ context.Context, _ ...string) error {
	console.Render("Available commands:")

	commands := []struct {
		cmd  string
		desc string
	}{
		{"/add <file patterns>", "Add files to the chat"},
		{"/list", "List all added files"},
		{"/remove <file patterns>", "Remove files from chat"},
		{"/ask <question>", "Ask about the code"},
		{"/drop", "Remove all files from chat"},
		{"/coding <instructions>", "Code with AI assistance"},
		{"/commit", "Commit changes"},
		{"/undo", "Undo last changes"},
		{"/diff", "Show changes in added files"},
		{"/exit", "Exit the terminal"},
		{"/help", "Show this help message"},
	}

	for _, cmd := range commands {
		formatted := fmt.Sprintf("  %-44s", cmd.cmd)
		console.Render(
			"%s%s",
			console.StdoutStyles().Flag.Render(formatted),
			console.StdoutStyles().FlagDesc.Render(cmd.desc),
		)
	}
	return nil
}

func (c *CommandExecutor) diff(_ context.Context, _ ...string) error {
	// Stage all added files
	filesToStage := make([]string, 0, len(c.coder.absFileNames))
	for file := range c.coder.absFileNames {
		filesToStage = append(filesToStage, file)
	}

	// Add files to git staging area
	if err := c.coder.repo.AddFiles(filesToStage); err != nil {
		return errbook.Wrap("Failed to stage files", err)
	}

	// Get the diff
	diffOutput, err := c.coder.repo.DiffFiles()
	if err != nil {
		return errbook.Wrap("Failed to get diff", err)
	}

	// Process and format the diff
	fmt.Println(c.coder.repo.FormatDiff(diffOutput))

	return nil
}

func (c *CommandExecutor) exit(_ context.Context, _ ...string) error {
	fmt.Println("Bye!")
	os.Exit(0)

	return nil
}

func (c *CommandExecutor) prepareAskCompletionMessages(userInput string) ([]llms.ChatMessage, error) {
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

	return messages, nil
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
	content, err := c.loadFileContent(filePath)
	if err != nil {
		return "", err
	}

	// For remote URLs, use the full URL as the identifier
	if rest.IsRemoteFile(filePath) {
		return fmt.Sprintf("\n%s%s", filePath, wrapFenceWithType(content, filePath)), nil
	}

	// For local files, use relative path
	relPath, err := filepath.Rel(c.coder.codeBasePath, filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\n%s%s", relPath, wrapFenceWithType(content, filePath)), nil
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
