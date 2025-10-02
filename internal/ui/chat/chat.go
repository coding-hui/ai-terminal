// Package chat implements the terminal-based chat interface for AI interactions.
// It handles the UI rendering, state management, and communication with the AI engine.
package chat

import (
	"context"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

// Define constants for chat history functionality
const (
	chatHistoryFilename = ".ai.chat.history.md"
)

// state represents the current state of the chat application
type state int

const (
	startState        state = iota // Initial state when chat starts
	configLoadedState              // State after configuration is loaded
	requestState                   // State when making a request to the AI
	responseState                  // State when receiving response from the AI
	doneState                      // State when chat is completed
	errorState                     // State when an error occurs
)

// Chat represents the main chat application structure
type Chat struct {
	Error      *errbook.AiError // Error encountered during chat
	TokenUsage llms.Usage       // Token usage statistics from the AI

	output     string // Raw output from the AI
	glamOutput string // Formatted output with markdown rendering

	state  state           // Current state of the chat
	opts   *Options        // Chat options
	config *options.Config // Application configuration
	engine *ai.Engine      // AI engine for processing requests

	anim         tea.Model             // Animation model for loading states
	renderer     *lipgloss.Renderer    // Text renderer for styling
	glam         *glamour.TermRenderer // Markdown renderer
	glamViewport viewport.Model        // Viewport for scrolling content
	styles       console.Styles        // Predefined styles for console output
	glamHeight   int                   // Height of the rendered markdown content
	width        int                   // Terminal width
	height       int                   // Terminal height

	content      []string    // Buffered content for non-TTY output
	contentMutex *sync.Mutex // Mutex for thread-safe content access
}

// NewChat creates and initializes a new Chat instance
// cfg: Application configuration
// opts: Optional parameters for customizing chat behavior
func NewChat(cfg *options.Config, opts ...Option) *Chat {
	o := NewOptions(opts...)

	gr, _ := glamour.NewTermRenderer(
		// detect configured terminal color support
		glamour.WithEnvironmentConfig(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(cfg.WordWrap),
	)
	vp := viewport.New(0, 0)
	vp.GotoBottom()

	return &Chat{
		engine:       o.engine,
		config:       cfg,
		glam:         gr,
		opts:         o,
		glamViewport: vp,
		contentMutex: &sync.Mutex{},
		renderer:     o.renderer,
		state:        startState,
		styles:       console.MakeStyles(o.renderer),
	}
}

// writeChatHistory writes chat interactions to the history file using HistoryWriter
func (c *Chat) writeChatHistory(input, response string) error {
	writer := NewHistoryWriter()
	
	// Build history content
	var historyContent strings.Builder

	// Write input and response
	if input != "" {
		// Format commands with '>' prefix and add prompt prefix
		if strings.HasPrefix(input, "/") {
			historyContent.WriteString(fmt.Sprintf("#### %s\n", input))
		} else {
			// For other inputs, use the '>' prefix with prompt prefix
			promptPrefix := c.getPromptPrefix()
			historyContent.WriteString(fmt.Sprintf("> %s %s\n", promptPrefix, input))
		}
		// Add a newline after input
		historyContent.WriteString("\n")
	}

	if response != "" {
		// Format responses that indicate file operations
		if strings.Contains(response, "Added") && strings.Contains(response, "to the chat") {
			historyContent.WriteString(fmt.Sprintf("> %s\n", response))
		} else if strings.Contains(response, "Error:") || strings.Contains(response, "Failed to") {
			// Format error messages with special markdown
			historyContent.WriteString(fmt.Sprintf("**Error:** %s\n", response))
		} else {
			// For other responses, just write them directly
			historyContent.WriteString(response)
			if !strings.HasSuffix(response, "\n") {
				historyContent.WriteString("\n")
			}
		}
		// Add a newline after response
		historyContent.WriteString("\n")
	}

	return writer.WriteToHistory(historyContent.String())
}

// Run starts the chat application and handles the main execution loop
// Returns error if the program fails to start or encounters an error during execution
func (c *Chat) Run() error {
	var opts []tea.ProgramOption
	if c.config.Raw || c.config.Quiet {
		opts = append(opts, tea.WithoutRenderer())
	}

	if _, err := tea.NewProgram(c, opts...).Run(); err != nil {
		return errbook.Wrap("Couldn't start Bubble Tea program.", err)
	}

	if c.Error != nil {
		return *c.Error
	}

	if term.IsOutputTTY() && !c.config.Raw {
		switch {
		case c.glamOutput != "":
			fmt.Print(c.glamOutput)
		case c.output != "":
			fmt.Print(c.output)
		}
	}

	// Add clipboard support
	if c.opts.copyToClipboard {
		if c.output != "" {
			if err := clipboard.WriteAll(c.output); err != nil {
				console.RenderError(err, "Failed to copy to clipboard")
			}
		}
	}

	if c.config.Show != "" || c.config.ShowLast {
		return nil
	}

	if c.config.CacheWriteToID != "" {
		return c.saveConversation()
	}

	return nil
}

// Init initializes the chat application and returns the initial command to execute
// This is part of the Bubble Tea framework's initialization process
func (c *Chat) Init() tea.Cmd {
	return c.startCliCmd()
}

// Update handles state updates and message processing for the chat application
// msg: The incoming message to process
// Returns the updated model and commands to execute
// This is part of the Bubble Tea framework's update loop
//
//nolint:golint,gocyclo
func (c *Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case convo.CacheDetailsMsg:
		c.config.CacheWriteToID = msg.WriteID
		c.config.CacheWriteToTitle = msg.Title
		c.config.CacheReadFromID = msg.ReadID
		c.config.Model = msg.Model

		if !c.config.Quiet {
			c.anim = console.NewAnim(c.config.Fanciness, c.config.LoadingText, c.renderer, c.styles)
			cmds = append(cmds, c.anim.Init())
		}
		c.state = configLoadedState
		cmds = append(cmds, c.readStdinCmd)

	case ai.CompletionInput:
		if len(msg.Messages) == 0 {
			return c, c.quit
		}
		c.state = requestState
		if c.config.Show != "" || c.config.ShowLast {
			cmds = append(cmds, c.readFromCacheCmd())
		} else {
			cmds = append(cmds, c.startCompletionCmd(msg.Messages), c.awaitChatCompletedCmd())
		}

	case ai.StreamCompletionOutput:
		if msg.GetContent() != "" {
			c.appendToOutput(msg.GetContent())
			c.state = responseState
		}
		if msg.IsLast() {
			c.state = doneState
			c.TokenUsage = msg.GetUsage()
			// Write chat history when conversation is completed
			if err := c.writeChatHistory("", c.GetOutput()); err != nil {
				console.RenderError(err, "Failed to write chat history")
			}
			return c, c.quit
		}
		cmds = append(cmds, c.awaitChatCompletedCmd())

	case errbook.AiError:
		c.Error = &msg
		c.state = errorState
		return c, c.quit

	case tea.WindowSizeMsg:
		c.width, c.height = msg.Width, msg.Height
		c.glamViewport.Width = c.width
		c.glamViewport.Height = c.height
		return c, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			c.state = doneState
			return c, c.quit
		}
	}

	if !c.config.Quiet && (c.state == configLoadedState || c.state == requestState) {
		var cmd tea.Cmd
		c.anim, cmd = c.anim.Update(msg)
		cmds = append(cmds, cmd)
	}

	if c.viewportNeeded() {
		// Only respond to keypresses when the viewport (i.e. the content) is
		// taller than the window.
		var cmd tea.Cmd
		c.glamViewport, cmd = c.glamViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

// viewportNeeded checks if a viewport is required based on content height
// Returns true if the content exceeds the terminal height
func (c *Chat) viewportNeeded() bool {
	return c.glamHeight > c.height
}

// View renders the current state of the chat application
// Returns the string representation of the current view
// This is part of the Bubble Tea framework's rendering process
func (c *Chat) View() string {
	switch c.state {
	case errorState:
		return ""
	case requestState:
		if !c.config.Quiet {
			return c.anim.View()
		}
	case responseState:
		if !c.config.Raw && term.IsOutputTTY() {
			if c.viewportNeeded() {
				return c.glamViewport.View()
			}
			// We don't need the viewport yet.
			return c.glamOutput
		}

		if term.IsOutputTTY() && !c.config.Raw {
			return c.output
		}

		c.contentMutex.Lock()
		for _, c := range c.content {
			fmt.Print(c)
		}
		c.content = []string{}
		c.contentMutex.Unlock()
	case doneState:
		if !term.IsOutputTTY() {
			fmt.Printf("\n")
		}
		return ""
	}
	return ""
}

const tabWidth = 4

// appendToOutput adds new content to the chat output and updates the view
// s: The string content to append to the output
func (c *Chat) appendToOutput(s string) {
	c.output += s
	if !term.IsOutputTTY() || c.config.Raw {
		c.contentMutex.Lock()
		c.content = append(c.content, s)
		c.contentMutex.Unlock()
		return
	}

	wasAtBottom := c.glamViewport.ScrollPercent() == 1.0
	oldHeight := c.glamHeight
	c.glamOutput, _ = c.glam.Render(html.UnescapeString(c.output))
	c.glamOutput = strings.TrimRightFunc(c.glamOutput, unicode.IsSpace)
	c.glamOutput = strings.ReplaceAll(c.glamOutput, "\t", strings.Repeat(" ", tabWidth))
	c.glamHeight = lipgloss.Height(c.glamOutput)
	c.glamOutput += "\n"
	truncatedGlamOutput := c.renderer.NewStyle().Padding(2).Width(c.width).Render(c.glamOutput)
	c.glamViewport.SetContent(truncatedGlamOutput)
	if oldHeight < c.glamHeight && wasAtBottom {
		// If the viewport's at the bottom and we've received a new
		// line of content, follow the output by auto scrolling to
		// the bottom.
		c.glamViewport.GotoBottom()
	}
}

// quit returns a quit message to terminate the chat application
// This is part of the Bubble Tea framework's shutdown process
func (c *Chat) quit() tea.Msg {
	return tea.Quit()
}

// startCompletionCmd creates a command to start an AI completion request
// messages: The chat messages to send to the AI
// Returns a command that will initiate the completion request
func (c *Chat) startCompletionCmd(messages []llms.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		output, err := c.engine.CreateStreamCompletion(context.Background(), messages)
		if err != nil {
			return err
		}
		return output
	}
}

// awaitChatCompletedCmd creates a command to wait for chat completion
// Returns a command that will wait for the AI response
func (c *Chat) awaitChatCompletedCmd() tea.Cmd {
	return func() tea.Msg {
		return <-c.engine.GetChannel()
	}
}

// startCliCmd creates a command to initialize the CLI interface
// Returns a command that will fetch the current conversation details
func (c *Chat) startCliCmd() tea.Cmd {
	return func() tea.Msg {
		details, err := convo.GetCurrentConversationID(context.Background(), c.config, c.engine.GetConvoStore())
		if err != nil {
			return err
		}
		return details
	}
}

// readFromCacheCmd creates a command to read messages from the conversation cache
// Returns a command that will fetch cached messages
func (c *Chat) readFromCacheCmd() tea.Cmd {
	return func() tea.Msg {
		convoStore := c.engine.GetConvoStore()
		messages, err := convoStore.Messages(context.Background(), c.config.CacheReadFromID)
		if err != nil {
			return err
		}
		lastContent := ""
		if len(messages) > 0 {
			lastContent = messages[len(messages)-1].GetContent()
		}
		return ai.StreamCompletionOutput{
			Content: lastContent,
			Last:    true,
		}
	}
}

// readStdinCmd reads input from stdin and creates a completion input message
// Returns a CompletionInput message containing the read messages
func (c *Chat) readStdinCmd() tea.Msg {
	var messages []llms.ChatMessage
	if len(c.opts.messages) > 0 {
		messages = append(messages, c.opts.messages...)
	}
	if c.opts.content != "" {
		messages = append(messages, llms.HumanChatMessage{
			Content: c.opts.content,
		})
	}
	return ai.CompletionInput{
		Messages: messages,
	}
}

// renderMarkdown renders markdown content for terminal display
// raw: The raw markdown string to render
// Returns the rendered string or the original if rendering fails
func (c *Chat) renderMarkdown(raw string) string {
	if c.config.Raw {
		return raw
	}
	rendered, err := c.glam.Render(raw)
	if err != nil {
		return raw
	}
	return rendered
}

// GetOutput returns the unescaped chat output
// Returns the processed chat output string
func (c *Chat) GetOutput() string {
	return html.UnescapeString(c.output)
}

// GetGlamOutput returns the formatted markdown output
// Returns the rendered markdown output string
func (c *Chat) GetGlamOutput() string {
	return c.glamOutput
}

// saveConversation saves the current conversation to persistent storage
// Returns error if the save operation fails
func (c *Chat) saveConversation() error {
	if c.config.NoCache {
		return nil
	}

	ctx := context.Background()
	convoStore := c.engine.GetConvoStore()
	writeToID := c.config.CacheWriteToID
	writeToTitle := strings.TrimSpace(c.config.CacheWriteToTitle)

	if convo.MatchSha1(writeToTitle) || writeToTitle == "" {
		messages, err := convoStore.Messages(ctx, writeToID)
		if err != nil {
			return err
		}
		writeToTitle = firstLine(lastPrompt(messages))
	}

	if writeToTitle == "" {
		writeToTitle = writeToID[:convo.Sha1short]
	}

	if err := convoStore.PersistentMessages(ctx, writeToID); err != nil {
		return errbook.Wrap(fmt.Sprintf(
			"There was a problem writing %s to the cache. Use %s / %s to disable it.",
			c.config.CacheWriteToID,
			console.StderrStyles().InlineCode.Render("--no-cache"),
			console.StderrStyles().InlineCode.Render("NO_CACHE"),
		), err)
	}

	if err := convoStore.SaveConversation(ctx, writeToID, writeToTitle, c.config.Model); err != nil {
		return errbook.Wrap(fmt.Sprintf(
			"There was a problem writing %s to the cache. Use %s / %s to disable it.",
			c.config.CacheWriteToID,
			console.StderrStyles().InlineCode.Render("--no-cache"),
			console.StderrStyles().InlineCode.Render("NO_CACHE"),
		), err)
	}

	// Write save confirmation to history file using HistoryWriter
	writer := NewHistoryWriter()
	
	// Build save content
	saveContent := fmt.Sprintf("\n**Conversation successfully saved:** `%s` `%s`\n", c.config.CacheWriteToID[:convo.Sha1short], writeToTitle)
	if c.config.ShowTokenUsages {
		saveContent += fmt.Sprintf("\nFirst Token: `%.3fs` | Avg: `%.3f/s` | Total: `%.3fs` | Tokens: `%d`",
			c.TokenUsage.FirstTokenTime.Seconds(),
			c.TokenUsage.AverageTokensPerSecond,
			c.TokenUsage.TotalTime.Seconds(),
			c.TokenUsage.TotalTokens,
		)
	}

	// Write to history file using HistoryWriter
	if err := writer.WriteToHistory(saveContent); err != nil {
		console.RenderError(err, "Failed to write save confirmation to chat history")
	}

	return nil
}

// lastPrompt finds the last human prompt in a list of chat messages
// messages: The list of chat messages to search
// Returns the content of the last human message
func lastPrompt(messages []llms.ChatMessage) string {
	var result string
	for _, msg := range messages {
		if msg.GetType() != llms.ChatMessageTypeHuman {
			continue
		}
		result = msg.GetContent()
	}
	return result
}

// getPromptPrefix returns the appropriate prompt prefix based on the chat mode
func (c *Chat) getPromptPrefix() string {
	// Determine the prompt mode from chat options
	// Since Chat doesn't have an explicit PromptMode field, we'll use a default
	// You may want to add a PromptMode field to Chat if needed
	promptPrefix := "chat" // Default prefix for chat mode

	// If the chat has access to configuration for prompt prefixes, use them
	if c.config != nil && c.config.AutoCoder.PromptPrefixChat != "" {
		promptPrefix = c.config.AutoCoder.PromptPrefixChat
	}

	return promptPrefix
}

// firstLine extracts the first line from a multi-line string
// s: The input string to process
// Returns the first line of the string
func firstLine(s string) string {
	first, _, _ := strings.Cut(s, "\n")
	return first
}
