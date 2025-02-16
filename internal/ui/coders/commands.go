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
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/prompt"
	"github.com/coding-hui/ai-terminal/internal/ui/chat"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/rest"
)

// supportCommands maps command names to their handler functions
var supportCommands = map[string]func(context.Context, string) error{}

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
	flags  map[string]bool
}

func NewCommandExecutor(coder *AutoCoder) *CommandExecutor {
	editor := NewEditBlockCoder(coder, fences[0])
	cmds := &CommandExecutor{coder: coder, editor: editor}
	cmds.registryCmds()
	return cmds
}

func (c *CommandExecutor) registryCmds() {
	supportCommands["add"] = c.add
	supportCommands["list"] = c.list
	supportCommands["remove"] = c.remove
	supportCommands["ask"] = c.ask
	supportCommands["design"] = c.design
	supportCommands["drop"] = c.drop
	supportCommands["coding"] = c.coding
	supportCommands["commit"] = c.commit
	supportCommands["undo"] = c.undo
	supportCommands["exit"] = c.exit
	supportCommands["diff"] = c.diff
	supportCommands["apply"] = c.apply
	supportCommands["chat-model"] = c.switchNewChatModel
	supportCommands["help"] = c.help
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

	// Parse command arguments
	filteredInput := c.parseFlags(args)
	userInput := strings.Join(filteredInput, " ")

	// Execute the recognized command
	if err := fn(context.Background(), userInput); err != nil {
		console.RenderError(err, "Failed to execute command %s", cmd)
	}
}

// ask Ask GPT to edit the files in the chat
func (c *CommandExecutor) ask(ctx context.Context, input string) error {
	messages, err := c.prepareAskCompletionMessages(input)
	if err != nil {
		return errbook.Wrap("Failed to prepare ask completion messages", err)
	}

	if c.flags[FlagVerbose] {
		return console.RenderChatMessages(messages)
	}

	chatModel := chat.NewChat(c.coder.cfg,
		chat.WithContext(ctx),
		chat.WithMessages(messages),
		chat.WithEngine(c.coder.engine),
	)

	return chatModel.Run()
}

// add Add files to the chat so GPT can edit them or review them in detail
func (c *CommandExecutor) loadFileContent(path string) (string, error) {
	// Handle remote URLs
	if rest.IsValidURL(path) {
		console.Render("Loading remote content [%s]", path)
		return rest.FetchURLContent(path)
	}

	// Handle local files
	console.Render("Loading local file [%s]", path)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errbook.Wrap("Failed to read local file", err)
	}

	return string(content), nil
}

func (c *CommandExecutor) add(_ context.Context, input string) (err error) {
	files := strings.Fields(input)
	if len(files) == 0 {
		return errbook.New("Please provide at least one file or URL")
	}

	var matchedFiles = make([]string, 0, len(files))

	for _, pattern := range files {
		// Handle remote URLs directly
		if rest.IsValidURL(pattern) {
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
		absPath := filePath
		if !rest.IsValidURL(filePath) {
			absPath, err = absFilePath(c.coder.codeBasePath, filePath)
			if err != nil {
				return errbook.Wrap("Failed to get abs path", err)
			}
		}

		// Check if file already loaded
		for _, lc := range c.coder.loadedContexts {
			if lc.FilePath == absPath || lc.URL == absPath {
				return errbook.New("File [%s] already exists", absPath)
			}
		}

		// Create new LoadContext
		lc := &convo.LoadContext{
			Type:    convo.ContentTypeFile,
			URL:     absPath,
			Content: "", // Will be loaded on demand
			Name:    filepath.Base(absPath),
		}
		if rest.IsValidURL(absPath) {
			lc.Type = convo.ContentTypeURL
		} else {
			lc.FilePath = absPath
		}

		c.coder.loadedContexts = append(c.coder.loadedContexts, lc)

		// Persist the new context
		if err := c.coder.saveContext(context.Background(), lc); err != nil {
			return errbook.Wrap("Failed to persist file context", err)
		}

		console.Render("Added [%s]", absPath)
	}

	return nil
}

// list List all files that have been added to the chat
func (c *CommandExecutor) list(_ context.Context, _ string) error {
	if len(c.coder.loadedContexts) <= 0 {
		console.Render("No files added in chat currently")
		return nil
	}
	no := 1
	for _, lc := range c.coder.loadedContexts {
		path := lc.FilePath
		if lc.Type == convo.ContentTypeURL {
			path = lc.URL
		}
		relPath, err := filepath.Rel(c.coder.codeBasePath, path)
		if err != nil {
			return errbook.Wrap("Failed to get relative path", err)
		}
		console.Render("%d.%s (%s)", no, relPath, lc.Type)
		no++
	}

	return nil
}

// remove files from the chat so GPT won't edit them or review them in detail
func (c *CommandExecutor) remove(_ context.Context, input string) error {
	files := strings.Fields(input)
	if len(files) == 0 {
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

			// Remove matching contexts
			newContexts := make([]*convo.LoadContext, 0, len(c.coder.loadedContexts))
			for _, lc := range c.coder.loadedContexts {
				if lc.FilePath == abs || lc.URL == abs {
					deleteCount++
					// Remove from persistence store
					if err := c.coder.deleteContext(context.Background(), lc.ID); err != nil {
						return errbook.Wrap("Failed to delete file context", err)
					}
					console.Render("Removed file [%s]", abs)
					continue
				}
				newContexts = append(newContexts, lc)
			}
			c.coder.loadedContexts = newContexts
		}
	}

	return nil
}

func (c *CommandExecutor) drop(_ context.Context, _ string) error {
	dropCount := len(c.coder.loadedContexts)

	// Clean all contexts from persistence store
	if _, err := c.coder.store.CleanContexts(context.Background(), c.coder.cfg.CacheWriteToID); err != nil {
		return errbook.Wrap("Failed to clean file contexts", err)
	}

	c.coder.loadedContexts = []*convo.LoadContext{}
	console.Render("Dropped %d files", dropCount)

	return nil
}

func (c *CommandExecutor) coding(ctx context.Context, input string) error {
	addedFiles, err := c.getAddedFileContent()
	if err != nil {
		return err
	}

	openFence, closeFence := c.editor.UpdateCodeFences(ctx, addedFiles)

	console.RenderStep("Selected coder block fences %s %s", openFence, closeFence)
	messages, err := c.editor.FormatMessages(map[string]any{
		userQuestionKey: input,
		addedFilesKey:   addedFiles,
		openFenceKey:    openFence,
		closeFenceKey:   closeFence,
		lazyPromptKey:   lazyPrompt,
	})
	if err != nil {
		return err
	}

	if c.flags[FlagVerbose] {
		return console.RenderChatMessages(messages)
	}

	err = c.editor.Execute(ctx, messages)
	if err != nil {
		return err
	}

	// Auto-commit if enabled in config
	if c.coder.cfg.AutoCoder.AutoCommit {
		if err := c.commit(ctx, ""); err != nil {
			return errbook.Wrap("Failed to auto-commit changes", err)
		}
	}

	// Check for modified/new files and prompt to save to context
	modifiedFiles, err := c.editor.GetModifiedFiles(ctx)
	if err != nil {
		return err
	}

	for _, file := range modifiedFiles {
		// Check if file is already in loaded contexts
		found := false
		for _, lc := range c.coder.loadedContexts {
			if lc.FilePath == file {
				found = true
				break
			}
		}

		if !found {
			if console.WaitForUserConfirm(console.Yes, "Do you want to add modified file %s to context?", file) {
				if err := c.add(ctx, file); err != nil {
					console.RenderError(err, "Failed to add file %s to context", file)
				}
			}
		}
	}

	return nil
}

func (c *CommandExecutor) undo(ctx context.Context, _ string) error {
	// First check if there are any files in the chat
	if len(c.coder.loadedContexts) == 0 {
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

func (c *CommandExecutor) commit(ctx context.Context, _ string) error {
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

func (c *CommandExecutor) design(ctx context.Context, input string) error {
	messages, err := c.prepareDesignCompletionMessages(input)
	if err != nil {
		return errbook.Wrap("Failed to prepare design completion messages", err)
	}

	if c.flags[FlagVerbose] {
		return console.RenderChatMessages(messages)
	}

	chatModel := chat.NewChat(c.coder.cfg,
		chat.WithContext(ctx),
		chat.WithMessages(messages),
		chat.WithEngine(c.coder.engine),
	)

	return chatModel.Run()
}

func (c *CommandExecutor) help(_ context.Context, _ string) error {
	console.Render("Available commands:")

	commands := []struct {
		cmd  string
		desc string
	}{
		{"/add <file patterns/URLs>", "Add local files or URLs to chat context"},
		{"/list", "List files in chat context"},
		{"/remove <patterns>", "Remove files from context"},
		{"/ask <question>", "Ask about code in context"},
		{"/drop", "Clear all files from context"},
		{"/design <requirements>", "Design system architecture and components"},
		{"/coding <instructions>", "Code with AI (use for details)"},
		{"/commit", "Commit changes to version control"},
		{"/undo", "Revert last code changes"},
		{"/diff", "Show diffs of context files"},
		{"/apply <edit blocks>", "Apply AI-generated code edits"},
		{"/chat-model <model> <api>", "Switch to a new chat mode"},
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

func (c *CommandExecutor) diff(_ context.Context, _ string) error {
	// Stage all added files
	filesToStage := make([]string, 0, len(c.coder.loadedContexts))
	for _, lc := range c.coder.loadedContexts {
		if lc.Type == convo.ContentTypeFile {
			filesToStage = append(filesToStage, lc.FilePath)
		}
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

func (c *CommandExecutor) apply(ctx context.Context, codes string) error {
	if codes == "" {
		return errbook.New("Please provide edit blocks to apply")
	}
	openFence, closeFence := chooseExistingFence(codes)

	console.Render("Selected coder block fences %s %s", openFence, closeFence)

	// Parse edit blocks from input
	edits, err := c.editor.GetEdits(ctx, codes, []string{openFence, closeFence})
	if err != nil {
		return errbook.Wrap("Failed to get edits blocks", err)
	}

	// Check if any edits were found
	if len(edits) == 0 {
		console.Render("No changes were applied - no valid edit blocks found")
		return nil
	}

	// Apply the edits
	if err := c.editor.ApplyEdits(ctx, edits); err != nil {
		return errbook.Wrap("Failed to apply edits", err)
	}

	console.Render("Successfully applied %d edit blocks", len(edits))

	return nil
}

func (c *CommandExecutor) exit(_ context.Context, _ string) error {
	fmt.Println("Bye!")
	os.Exit(0)

	return nil
}

func (c *CommandExecutor) prepareDesignCompletionMessages(userInput string) ([]llms.ChatMessage, error) {
	addedFileMessages, err := c.getAddedFileContent()
	if err != nil {
		return nil, err
	}

	messages, err := promptDesign.FormatMessages(map[string]any{
		addedFilesKey:   addedFileMessages,
		userQuestionKey: userInput,
	})
	if err != nil {
		return nil, err
	}

	return messages, nil
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
	if len(c.coder.loadedContexts) > 0 {
		for _, lc := range c.coder.loadedContexts {
			filePath := lc.FilePath
			if lc.Type == convo.ContentTypeURL {
				filePath = lc.URL
			}
			content, err := c.formatFileContent(filePath)
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
	if rest.IsValidURL(filePath) {
		name := rest.SanitizeURL(filePath)
		// show the first 20 characters, then ellipsis then the last 20 characters of 'name'
		if len(name) > 40 {
			name = name[:20] + "⋯" + name[len(name)-20:]
		}
		return fmt.Sprintf("\n%s%s", filePath, wrapFenceWithType(content, name)), nil
	}

	// For local files, use relative path
	relPath, err := filepath.Rel(c.coder.codeBasePath, filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\n%s%s", relPath, wrapFenceWithType(content, filePath)), nil
}

func (c *CommandExecutor) switchNewChatModel(_ context.Context, input string) error {
	args := strings.Fields(input)
	if len(args) < 2 {
		return errbook.New("Please provide both model and api parameters")
	}

	model := args[0]
	api := args[1]

	// Validate the model exists
	if _, err := c.coder.cfg.GetModel(model); err != nil {
		return errbook.Wrap("Invalid model", err)
	}

	// Validate the API exists
	if _, err := c.coder.cfg.GetAPI(api); err != nil {
		return errbook.Wrap("Invalid API", err)
	}

	// Update config
	c.coder.cfg.AutoCoder.CodingModel = model
	c.coder.cfg.API = api

	console.Render("Updated coding model to %s using API %s", model, api)

	return nil
}

func (c *CommandExecutor) parseFlags(args []string) (filteredArgs []string) {
	c.flags = make(map[string]bool)

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			c.flags[arg[2:]] = true
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	return
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
