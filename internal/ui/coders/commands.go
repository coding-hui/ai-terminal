// Package coders implements command execution for the AI terminal.
// Handles commands for file management, AI coding, version control, and chat operations.
package coders

import (
	"context"
	"errors"
	"fmt"
	"html"
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
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/ui"
	"github.com/coding-hui/ai-terminal/internal/ui/chat"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/rest"
)

// supportCommands maps command names to their handler implementations
var supportCommands = map[string]func(context.Context, string) error{}

// getSupportedCommands returns all registered command names
func getSupportedCommands() []string {
	commands := make([]string, 0, len(supportCommands))
	for cmd := range supportCommands {
		commands = append(commands, cmd)
	}
	return commands
}

type CommandExecutor struct {
	coder         *AutoCoder
	editor        *EditBlockCoder
	flags         map[string]bool
	historyWriter *chat.HistoryWriter
}

func NewCommandExecutor(coder *AutoCoder, historyWriter *chat.HistoryWriter) *CommandExecutor {
	editor := NewEditBlockCoder(coder, fences[0])
	cmds := &CommandExecutor{
		coder:         coder,
		editor:        editor,
		historyWriter: historyWriter,
	}
	cmds.registryCmds()
	return cmds
}

func (c *CommandExecutor) registryCmds() {
	supportCommands["/add"] = c.add
	supportCommands["/list"] = c.list
	supportCommands["/remove"] = c.remove
	supportCommands["/ask"] = c.ask
	supportCommands["/design"] = c.design
	supportCommands["/drop"] = c.drop
	supportCommands["/coding"] = c.coding
	supportCommands["/exec"] = c.exec
	supportCommands["/commit"] = c.commit
	supportCommands["/undo"] = c.undo
	supportCommands["/exit"] = c.exit
	supportCommands["/diff"] = c.diff
	supportCommands["/apply"] = c.apply
	supportCommands["/chat-model"] = c.switchNewChatModel
	supportCommands["/help"] = c.help
	supportCommands["/clear"] = c.clear
}

// isCommand detects if input is a command (prefixed with ! or /)
func (c *CommandExecutor) isCommand(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, "!") || strings.HasPrefix(input, "/")
}

// Executor processes command input - parsing, validation and execution
func (c *CommandExecutor) Executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Handle non-command input
	if !c.isCommand(input) {
		// Map plain input to the active mode's default action
		switch c.coder.promptMode {
		case ui.ChatPromptMode:
			input = "/ask " + input
		case ui.ExecPromptMode:
			input = "/exec " + input
		default:
			input = "/coding " + input
		}
	}

	// Handle command execution
	cmd, args := extractCmdArgs(input)
	fn, ok := supportCommands[cmd]
	if !ok {
		c.historyWriter.RenderError(
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
		c.historyWriter.RenderError(err, "Failed to execute command %s", cmd)
	}
}

// ask queries GPT to analyze or edit files in context
func (c *CommandExecutor) ask(ctx context.Context, input string) error {
	if input == "" {
		c.historyWriter.RenderComment("Switched /ask mode")
		return nil
	}

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
		chat.WithPromptMode(c.coder.promptMode),
		chat.WithCopyToClipboard(true),
	)

	return chatModel.Run()
}

// add registers files/URLs for GPT to analyze or edit
func (c *CommandExecutor) loadFileContent(path string) (string, error) {
	// Handle remote URLs
	if rest.IsValidURL(path) {
		c.historyWriter.Render("Loading remote content [%s]", path)
		return rest.FetchURLContent(path)
	}

	// Handle local files
	c.historyWriter.Render("Loading local file [%s]", path)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errbook.Wrap("Failed to read local file", err)
	}

	return string(content), nil
}

func (c *CommandExecutor) add(_ context.Context, input string) (err error) {
	files := strings.Fields(input)
	if len(files) == 0 {
		e := errbook.New("Please provide at least one file or URL")
		c.historyWriter.RenderError(e, "")
		return e
	}

	var matchedFiles = make([]string, 0, len(files))

	for _, pattern := range files {
		// Handle remote URLs directly
		if rest.IsValidURL(pattern) {
			matchedFiles = append(matchedFiles, pattern)
			continue
		}

		// Construct full path for the pattern
		fullPath := filepath.Join(c.coder.codeBasePath, pattern)

		// Check if the path is a directory
		fileInfo, err := os.Stat(fullPath)
		if err == nil && fileInfo.IsDir() {
			// Handle directory recursively
			c.historyWriter.Render("Processing directory [%s]", fullPath)
			err := filepath.Walk(fullPath, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				// Skip directories
				if info.IsDir() {
					return nil
				}
				matchedFiles = append(matchedFiles, path)
				return nil
			})
			if err != nil {
				return errbook.Wrap("Failed to walk directory", err)
			}
			continue
		}

		// Handle local file patterns with glob
		matches, err := filepath.Glob(filepath.Join(c.coder.codeBasePath, pattern))
		if err != nil {
			return errbook.Wrap("Failed to glob files", err)
		}

		if len(matches) <= 0 {
			c.historyWriter.Render("No files matched pattern [%s]", pattern)
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
				e := errbook.New("File [%s] not found", filePath)
				c.historyWriter.RenderError(e, "")
				return e
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
				e := errbook.New("File [%s] already exists", absPath)
				c.historyWriter.RenderError(e, "")
				return e
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

		c.historyWriter.Render("Added [%s]", absPath)
	}

	return nil
}

// list displays all files currently in context
func (c *CommandExecutor) list(_ context.Context, _ string) error {
	if len(c.coder.loadedContexts) <= 0 {
		c.historyWriter.Render("No files added in chat currently")
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
		c.historyWriter.Render("%d.%s (%s)", no, relPath, lc.Type)
		no++
	}

	return nil
}

// remove deletes files from context to exclude from GPT analysis
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
					c.historyWriter.Render("Removed file [%s]", abs)
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
	c.historyWriter.Render("Dropped %d files", dropCount)

	return nil
}

func (c *CommandExecutor) coding(ctx context.Context, input string) error {
	addedFiles, err := c.getAddedFileContent()
	if err != nil {
		return err
	}

	if input == "" {
		c.historyWriter.RenderComment("Switched /coding mode")
		return nil
	}

	if len(addedFiles) == 0 {
		return errbook.New("No files added in chat currently. Use /add to add files first")
	}

	openFence, closeFence := c.editor.UpdateCodeFences(ctx, addedFiles)

	c.historyWriter.Render("Selected coder block fences %s %s", openFence, closeFence)
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
					c.historyWriter.Render("Failed to add file %s to context", file)
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
		c.historyWriter.Render("Undo canceled")
		return nil
	}

	if err := c.coder.repo.RollbackLastCommit(); err != nil {
		return errbook.Wrap("Failed to undo changes", err)
	}

	c.historyWriter.Render("Successfully undone last changes")
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
		commit.WithIsAutoCoder(true), // Mark as auto coder commit
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
		chat.WithPromptMode(c.coder.promptMode),
		chat.WithCopyToClipboard(true),
	)

	return chatModel.Run()
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

	c.historyWriter.Render("Selected coder block fences %s %s", openFence, closeFence)

	// Parse edit blocks from input
	edits, err := c.editor.GetEdits(ctx, codes, []string{openFence, closeFence})
	if err != nil {
		return errbook.Wrap("Failed to get edits blocks", err)
	}

	// Check if any edits were found
	if len(edits) == 0 {
		c.historyWriter.Render("No changes were applied - no valid edit blocks found")
		return nil
	}

	// Apply the edits
	if err := c.editor.ApplyEdits(ctx, edits); err != nil {
		return errbook.Wrap("Failed to apply edits", err)
	}

	return nil
}

// exec infers a POSIX shell command via AI and executes it
func (c *CommandExecutor) exec(_ context.Context, input string) error {
	if input == "" {
		c.historyWriter.RenderComment("Switched /exec mode")
		return nil
	}

	if strings.TrimSpace(input) == "" {
		return errbook.New("Please provide an instruction to execute")
	}

	// Get current user and OS info
	currentUser := c.coder.cfg.System.GetUsername()
	osInfo := c.coder.cfg.System.GetOperatingSystem()
	archInfo := c.coder.cfg.System.GetDistribution()

	system := llms.SystemChatMessage{Content: strings.Join([]string{
		"You are a helpful terminal assistant.",
		"Convert the user's request into a single-line POSIX shell command.",
		fmt.Sprintf("Current user: %s", currentUser),
		fmt.Sprintf("Current operating system: %s (%s)", osInfo, archInfo),
		"Return ONLY the command without explanations, quotes, code fences, or newlines.",
		"If the request is unclear or unsafe (like destructive operations), respond with [noexec] and a short reason.",
	}, " ")}
	human := llms.HumanChatMessage{Content: input}

	ch := chat.NewChat(c.coder.cfg,
		chat.WithEngine(c.coder.engine),
		chat.WithPromptMode(c.coder.promptMode),
		chat.WithMessages([]llms.ChatMessage{system, human}),
	)
	if err := ch.Run(); err != nil {
		return err
	}

	cmd := strings.TrimSpace(ch.GetOutput())
	cmd = strings.Trim(cmd, "`")
	if cmd == "" || strings.HasPrefix(strings.ToLower(cmd), "[noexec]") {
		c.historyWriter.Render("No executable command was produced")
		return nil
	}
	if idx := strings.IndexByte(cmd, '\n'); idx >= 0 {
		cmd = strings.TrimSpace(cmd[:idx])
	}

	// Confirm before execution unless --yes flag is provided
	if !c.flags[FlagYes] {
		if !console.WaitForUserConfirm(console.Yes, "\nRun inferred command? %s", console.StderrStyles().InlineCode.Render(cmd)) {
			return nil
		}
	}

	shellCmd := runner.PrepareInteractiveCommand(cmd)
	shellCmd.Stdin = os.Stdin
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr
	return shellCmd.Run()
}

func (c *CommandExecutor) help(_ context.Context, _ string) error {
	c.historyWriter.Render("Available commands (grouped by functionality):")

	// Group commands by functionality
	fileCommands := []ui.Command{
		{Name: "/add <file/folder patterns/URLs>", Desc: "Add local files or URLs to chat context"},
		{Name: "/list", Desc: "List files currently in chat context"},
		{Name: "/remove <patterns>", Desc: "Remove files from context"},
		{Name: "/drop", Desc: "Clear all files from context"},
	}

	aiCommands := []ui.Command{
		{Name: "/ask <question>", Desc: "Ask questions about code in context"},
		{Name: "/design <requirements>", Desc: "Design system architecture and components"},
		{Name: "/coding <instructions>", Desc: "Generate and modify code with AI"},
	}

	codeManagementCommands := []ui.Command{
		{Name: "/commit", Desc: "Commit changes to version control"},
		{Name: "/undo", Desc: "Revert last code changes"},
		{Name: "/diff", Desc: "Show diffs of context files"},
		{Name: "/apply <edit blocks>", Desc: "Apply AI-generated code edits directly"},
	}

	systemCommands := []ui.Command{
		{Name: "/exec <instruction>", Desc: "Infer and execute a shell command"},
		{Name: "/chat-model <model> <api>", Desc: "Switch to a different chat model and API"},
		{Name: "/clear", Desc: "Clear current conversation"},
		{Name: "/exit", Desc: "Exit the terminal"},
		{Name: "/help", Desc: "Show this help message"},
	}

	// Render each group with headers
	c.historyWriter.Render("\n📁 File Management:")
	c.historyWriter.RenderHelps(fileCommands)

	c.historyWriter.Render("\n🤖 AI Interactions:")
	c.historyWriter.RenderHelps(aiCommands)

	c.historyWriter.Render("\n🔧 Code Management:")
	c.historyWriter.RenderHelps(codeManagementCommands)

	c.historyWriter.Render("\n⚙️  System Operations:")
	c.historyWriter.RenderHelps(systemCommands)

	c.historyWriter.Render("\n💡 Tips:")
	c.historyWriter.Render("  • Type commands directly or use ! prefix")
	c.historyWriter.Render("  • In default mode, input is treated as /coding command")
	c.historyWriter.Render("  • Use /ask mode for questions, /exec for shell commands")
	c.historyWriter.Render("  • Add --verbose flag to see raw AI messages")
	c.historyWriter.Render("  • Add -y flag to skip confirmations")

	return nil
}

func (c *CommandExecutor) clear(ctx context.Context, _ string) error {
	// Clear all loaded contexts first
	if _, err := c.coder.store.CleanContexts(ctx, c.coder.cfg.CacheWriteToID); err != nil {
		return errbook.Wrap("Failed to clean file contexts", err)
	}
	c.coder.loadedContexts = []*convo.LoadContext{}

	// Clear conversation messages
	if err := c.coder.store.InvalidateMessages(ctx, c.coder.cfg.CacheWriteToID); err != nil {
		return errbook.Wrap("Failed to clear conversation messages", err)
	}

	c.historyWriter.Render("Cleared current conversation")
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

	// Use different prompt templates based on whether files are present
	if len(addedFileMessages) > 0 {
		messages, err := promptAskWithFiles.FormatMessages(map[string]any{
			addedFilesKey:   addedFileMessages,
			userQuestionKey: userInput,
		})
		if err != nil {
			return nil, err
		}
		return messages, nil
	} else {
		// No files added - use general assistant prompt
		messages, err := promptAskGeneral.FormatMessages(map[string]any{
			userQuestionKey: userInput,
		})
		if err != nil {
			return nil, err
		}
		return messages, nil
	}
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
		return fmt.Sprintf("\n%s%s", filePath, html.UnescapeString(wrapFenceWithType(content, name, c.coder.cfg.AutoCoder.GetDefaultFences()))), nil
	}

	// For local files, use relative path
	relPath, err := filepath.Rel(c.coder.codeBasePath, filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\n%s%s", relPath, html.UnescapeString(wrapFenceWithType(content, filePath, c.coder.cfg.AutoCoder.GetDefaultFences()))), nil
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

	c.historyWriter.Render("Updated coding model to %s using API %s", model, api)

	return nil
}

func (c *CommandExecutor) parseFlags(args []string) (filteredArgs []string) {
	c.flags = make(map[string]bool)

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			c.flags[arg[2:]] = true
			continue
		}
		// support common short flags
		if strings.HasPrefix(arg, "-") && len(arg) == 2 {
			switch arg {
			case "-y":
				c.flags[FlagYes] = true
				continue
			}
		}
		filteredArgs = append(filteredArgs, arg)
	}

	return
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
