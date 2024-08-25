package coders

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/ui"
	"github.com/coding-hui/ai-terminal/internal/util/display"
)

var program *tea.Program

type State struct {
	error      error
	buffer     string
	command    string
	confirming bool
}

// AutoCoder is a auto generate coders user interface.
type AutoCoder struct {
	state State

	command      *command
	gitRepo      *git.Command
	codeBasePath string
	absFileNames map[string]struct{}

	history        *ui.History
	checkpointChan chan Checkpoint
	checkpoints    []Checkpoint

	cfg       *options.Config
	llmEngine *llm.Engine
}

func StartAutCoder() error {
	coder := NewAutoCoder()
	program = tea.NewProgram(coder)
	if _, err := program.Run(); err != nil {
		fmt.Println("Error running auto chat program:", err)
		os.Exit(1)
	}
	return nil
}

func NewAutoCoder() *AutoCoder {
	return &AutoCoder{
		state: State{
			error:   nil,
			buffer:  "",
			command: "",
		},
		gitRepo:        git.New(),
		cfg:            options.NewConfig(),
		checkpoints:    []Checkpoint{},
		history:        ui.NewHistory(),
		absFileNames:   map[string]struct{}{},
		checkpointChan: make(chan Checkpoint),
	}
}

func (a *AutoCoder) Init() tea.Cmd {
	var err error

	a.cfg = options.NewConfig()
	a.llmEngine, err = llm.NewLLMEngine(llm.ChatEngineMode, a.cfg)
	if err != nil {
		display.FatalErr(err)
		return tea.Quit
	}

	root, err := a.gitRepo.GitDir()
	if err != nil {
		display.FatalErr(err)
		return tea.Quit
	}

	a.codeBasePath = filepath.Dir(root)
	a.command = newCommand(a)

	return tea.Sequence(
		tea.ClearScreen,
		tea.Println(components.renderer.RenderContent(components.renderer.RenderWelcomeMessage(a.cfg.System.GetUsername()))),
		textinput.Blink,
		a.statusTickCmd(),
	)
}

func (a *AutoCoder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds       []tea.Cmd
		promptCmd  tea.Cmd
		spinnerCmd tea.Cmd
		confirmCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if a.isQuerying() && len(a.state.buffer) <= 0 {
			components.spinner, spinnerCmd = components.spinner.Update(msg)
			cmds = append(
				cmds,
				spinnerCmd,
			)
		}
	case tea.WindowSizeMsg:
		components.renderer = NewRenderer(
			glamour.WithEmoji(),
			glamour.WithAutoStyle(),
			glamour.WithPreservedNewLines(),
			glamour.WithWordWrap(msg.Width),
		)

	case Checkpoint:
		if len(a.checkpoints) <= 0 || a.checkpoints[len(a.checkpoints)-1].Desc != msg.Desc {
			a.checkpoints = append(a.checkpoints, msg)
		}
		cmds = append(
			cmds,
			a.statusTickCmd(),
			components.spinner.Tick,
		)

	case WaitFormUserConfirm:
		components.confirm, confirmCmd = components.confirm.Update(msg)
		return a, tea.Sequence(
			confirmCmd,
			textinput.Blink,
		)

	case tea.KeyMsg:
		switch msg.Type {
		// quit
		case tea.KeyCtrlC:
			return a, tea.Quit
		// help
		case tea.KeyCtrlH:
			if !a.isQuerying() {
				components.prompt, promptCmd = components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.Println(components.renderer.RenderContent(components.renderer.RenderHelpMessage())),
					textinput.Blink,
				)
			}

		// history
		case tea.KeyUp, tea.KeyDown:
			if !a.isQuerying() {
				var input *string
				if msg.Type == tea.KeyUp {
					input = a.history.GetPrevious()
				} else {
					input = a.history.GetNext()
				}
				if input != nil {
					components.prompt.SetValue(*input)
					components.prompt, promptCmd = components.prompt.Update(msg)
					cmds = append(
						cmds,
						promptCmd,
					)
				}
			}

		// handle user input
		case tea.KeyEnter:
			input := components.prompt.GetValue()
			if input != "" {
				a.state.buffer = ""
				a.checkpoints = make([]Checkpoint, 0)
				a.history.Add(input)
				inputPrint := components.prompt.AsString()
				components.prompt.SetValue("")
				components.prompt.Focus()
				components.prompt, promptCmd = components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.Println(inputPrint),
					a.command.run(input),
					a.command.awaitChatCompleted(),
				)
			}

		// clear
		case tea.KeyCtrlL:
			if !a.isQuerying() {
				components.prompt.SetValue("")
				components.prompt, promptCmd = components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.ClearScreen,
					textinput.Blink,
				)
			}

		// reset
		case tea.KeyCtrlR:
			if !a.isQuerying() {
				a.reset()
				components.prompt.SetValue("")
				components.prompt, promptCmd = components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.ClearScreen,
					textinput.Blink,
				)
			}

		default:
			if a.state.confirming && components.confirm != nil {
				components.confirm, confirmCmd = components.confirm.Update(msg)
				return a, tea.Sequence(
					confirmCmd,
					textinput.Blink,
				)
			}
			components.prompt.Focus()
			components.prompt, promptCmd = components.prompt.Update(msg)
			cmds = append(
				cmds,
				promptCmd,
				textinput.Blink,
			)
		}

	// engine chat stream feedback
	case llm.EngineChatStreamOutput:
		if msg.IsLast() {
			output := components.renderer.RenderContent(a.state.buffer)
			components.prompt.Focus()
			return a, tea.Sequence(
				tea.Println(output),
				textinput.Blink,
			)
		} else {
			return a, a.command.awaitChatCompleted()
		}

	case error:
		a.state.error = msg
		return a, nil
	}

	return a, tea.Batch(cmds...)
}

func (a *AutoCoder) View() string {
	started := len(a.checkpoints) > 0
	done := started && a.checkpoints[len(a.checkpoints)-1].Done

	if a.state.confirming && components.confirm != nil {
		return components.confirm.View()
	}

	if started && a.checkpoints[len(a.checkpoints)-1].Error != nil {
		return fmt.Sprintf("\n%s\n\n%s\n",
			components.renderer.RenderError(fmt.Sprintf("%s", a.checkpoints[len(a.checkpoints)-1].Error)),
			components.prompt.View(),
		)
	}

	if started {
		doneMsg := ""
		for _, s := range a.checkpoints[:len(a.checkpoints)-1] {
			icon := checkpointIcon(s.Type)
			switch s.Type {
			case StatusLoading:
				if !done {
					doneMsg += icon + s.Desc + "\n"
				}
			case StatusSuccess:
				doneMsg += icon + components.renderer.RenderSuccess(s.Desc) + "\n"
			case StatusWarning:
				doneMsg += icon + components.renderer.RenderWarning(s.Desc) + "\n"
			default:
				doneMsg += icon + s.Desc + "\n"
			}
		}
		if !done {
			if len(a.state.buffer) > 0 {
				return components.renderer.RenderContent(a.state.buffer)
			}
			return components.spinner.ViewWithMessage(doneMsg, a.checkpoints[len(a.checkpoints)-1].Desc)
		}
		if len(doneMsg) > 0 {
			return fmt.Sprintf("\n%s\n%s",
				doneMsg,
				components.prompt.View(),
			)
		}
	}

	return components.prompt.View()
}

func (a *AutoCoder) reset() {
	a.checkpoints = make([]Checkpoint, 0)
	a.history.Reset()
	a.absFileNames = make(map[string]struct{})
	a.state.buffer = ""
}
