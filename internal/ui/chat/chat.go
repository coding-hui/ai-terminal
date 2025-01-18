package chat

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

type state int

const (
	startState state = iota
	configLoadedState
	requestState
	responseState
	doneState
	errorState
)

type Chat struct {
	Error      *errbook.AiError
	Output     string
	GlamOutput string

	state  state
	opts   *Options
	config *options.Config
	engine *llm.Engine

	anim         tea.Model
	renderer     *lipgloss.Renderer
	glam         *glamour.TermRenderer
	glamViewport viewport.Model
	styles       console.Styles
	glamHeight   int
	width        int
	height       int

	content      []string
	contentMutex *sync.Mutex
}

func NewChat(cfg *options.Config, opts ...Option) *Chat {
	o := NewOptions(opts...)

	gr, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
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

// Run starts the chat.
func (c *Chat) Run() error {
	if _, err := tea.NewProgram(c).Run(); err != nil {
		return errbook.Wrap("Couldn't start Bubble Tea program.", err)
	}

	if c.Error != nil {
		return *c.Error
	}

	if term.IsOutputTTY() {
		if c.config.Raw && c.Output != "" {
			fmt.Print(c.Output)
		} else {
			switch {
			case c.GlamOutput != "":
				fmt.Print(c.GlamOutput)
			case c.Output != "":
				fmt.Print(c.Output)
			}
		}
	}

	if c.config.Show != "" || c.config.ShowLast {
		return nil
	}

	if c.config.CacheWriteToID != "" {
		return saveConversation(c)
	}

	return nil
}

func (c *Chat) Init() tea.Cmd {
	return c.startCliCmd()
}

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

	case llm.CompletionInput:
		if len(msg.Messages) == 0 {
			return c, c.quit
		}
		c.state = requestState
		cmds = append(cmds, c.startCompletionCmd(msg.Messages), c.awaitChatCompletedCmd())

	case llm.StreamCompletionOutput:
		if msg.GetContent() != "" {
			c.appendToOutput(msg.GetContent())
			c.state = responseState
		}
		if msg.IsLast() {
			c.state = doneState
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

func (c *Chat) viewportNeeded() bool {
	return c.glamHeight > c.height
}

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
			return c.GlamOutput
		}

		if term.IsOutputTTY() && !c.config.Raw {
			return c.Output
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

func (c *Chat) appendToOutput(s string) {
	c.Output += s
	if !term.IsOutputTTY() || c.config.Raw {
		c.contentMutex.Lock()
		c.content = append(c.content, s)
		c.contentMutex.Unlock()
		return
	}

	wasAtBottom := c.glamViewport.ScrollPercent() == 1.0
	oldHeight := c.glamHeight
	c.GlamOutput, _ = c.glam.Render(c.Output)
	c.GlamOutput = strings.TrimRightFunc(c.GlamOutput, unicode.IsSpace)
	c.GlamOutput = strings.ReplaceAll(c.GlamOutput, "\t", strings.Repeat(" ", tabWidth))
	c.glamHeight = lipgloss.Height(c.GlamOutput)
	c.GlamOutput += "\n"
	truncatedGlamOutput := c.renderer.NewStyle().Padding(2).Width(c.width).Render(c.GlamOutput)
	c.glamViewport.SetContent(truncatedGlamOutput)
	if oldHeight < c.glamHeight && wasAtBottom {
		// If the viewport's at the bottom and we've received a new
		// line of content, follow the output by auto scrolling to
		// the bottom.
		c.glamViewport.GotoBottom()
	}
}

func (c *Chat) quit() tea.Msg {
	return tea.Quit()
}

func (c *Chat) startCompletionCmd(messages []llms.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		output, err := c.engine.CreateStreamCompletion(context.Background(), messages)
		if err != nil {
			return err
		}
		return output
	}
}

func (c *Chat) awaitChatCompletedCmd() tea.Cmd {
	return func() tea.Msg {
		return <-c.engine.GetChannel()
	}
}

func (c *Chat) startCliCmd() tea.Cmd {
	return func() tea.Msg {
		details, err := convo.GetCurrentConversationID(context.Background(), c.config, c.engine.GetConvoStore())
		if err != nil {
			return err
		}
		return details
	}
}

// readStdinCmd reads from stdin and returns a CompletionInput message.
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
	return llm.CompletionInput{
		Messages: messages,
	}
}

func saveConversation(chat *Chat) error {
	if chat.config.NoCache {
		if !chat.config.Quiet {
			fmt.Fprintf(
				os.Stderr,
				"\nConversation was not saved because %s or %s is set.\n",
				console.StderrStyles().InlineCode.Render("--no-cache"),
				console.StderrStyles().InlineCode.Render("NO_CACHE"),
			)
		}
		return nil
	}

	ctx := context.Background()
	convoStore := chat.engine.GetConvoStore()
	writeToID := chat.config.CacheWriteToID
	writeToTitle := strings.TrimSpace(chat.config.CacheWriteToTitle)

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
			chat.config.CacheWriteToID,
			console.StderrStyles().InlineCode.Render("--no-cache"),
			console.StderrStyles().InlineCode.Render("NO_CACHE"),
		), err)
	}

	if err := convoStore.SaveConversation(ctx, writeToID, writeToTitle, chat.config.Model); err != nil {
		return errbook.Wrap(fmt.Sprintf(
			"There was a problem writing %s to the cache. Use %s / %s to disable it.",
			chat.config.CacheWriteToID,
			console.StderrStyles().InlineCode.Render("--no-cache"),
			console.StderrStyles().InlineCode.Render("NO_CACHE"),
		), err)
	}

	if !chat.config.Quiet {
		fmt.Fprintln(
			os.Stderr,
			"\nConversation saved:",
			console.StderrStyles().InlineCode.Render(chat.config.CacheWriteToID[:convo.Sha1short]),
			console.StderrStyles().Comment.Render(writeToTitle),
		)
	}

	return nil
}

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

func firstLine(s string) string {
	first, _, _ := strings.Cut(s, "\n")
	return first
}
