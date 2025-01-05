package chat

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui"
	"github.com/coding-hui/ai-terminal/internal/ui/display"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

const (
	defaultWidth = 120
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

	state   state
	config  *options.Config
	input   *ui.Input
	engine  *llm.Engine
	history *ui.History

	anim         tea.Model
	renderer     *lipgloss.Renderer
	glam         *glamour.TermRenderer
	glamViewport viewport.Model
	styles       display.Styles
	glamHeight   int
	width        int
	height       int

	content      []string
	contentMutex *sync.Mutex
}

func NewChat(input *ui.Input, r *lipgloss.Renderer, cfg *options.Config) (*Chat, error) {
	gr, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(cfg.WordWrap),
	)
	vp := viewport.New(0, 0)
	vp.GotoBottom()

	engine, err := llm.NewLLMEngine(llm.ChatEngineMode, cfg)
	if err != nil {
		return nil, err
	}

	return &Chat{
		engine:       engine,
		config:       cfg,
		glam:         gr,
		input:        input,
		glamViewport: vp,
		contentMutex: &sync.Mutex{},
		renderer:     r,
		state:        startState,
		styles:       display.MakeStyles(r),
		history:      ui.NewHistory(),
	}, nil
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
	case ui.CacheDetailsMsg:
		if !c.config.Quiet {
			c.anim = ui.NewAnim(c.config.Fanciness, c.config.LoadingText, c.renderer, c.styles)
			cmds = append(cmds, c.anim.Init())
		}
		c.state = configLoadedState
		cmds = append(cmds, c.readStdinCmd)
	case llm.CompletionInput:
		if removeWhitespace(msg.Content) == "" {
			return c, c.quit
		}
		c.state = requestState
		cmds = append(cmds, c.startCompletionCmd(msg.Content), c.awaitChatCompletedCmd())
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

func (c Chat) viewportNeeded() bool {
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

func (c *Chat) startCompletionCmd(content string) tea.Cmd {
	return func() tea.Msg {
		return c.engine.CreateStreamCompletion(content)
	}
}

func (c *Chat) awaitChatCompletedCmd() tea.Cmd {
	return func() tea.Msg {
		return <-c.engine.GetChannel()
	}
}

func (c *Chat) startCliCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CacheDetailsMsg{
			WriteID: "",
			Title:   "",
			ReadID:  "",
			Model:   "",
		}
	}
}

// readStdinCmd reads from stdin and returns a CompletionInput message.
func (c *Chat) readStdinCmd() tea.Msg {
	return llm.CompletionInput{Content: c.input.GetPipe() + "\n\n" + c.input.GetArgs()}
}

// if the input is whitespace only, make it empty.
func removeWhitespace(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return s
}
