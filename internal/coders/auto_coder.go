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
	error   error
	buffer  string
	command string
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
		tea.Println(components.renderer.RenderContent(components.renderer.RenderHelpMessage())),
		textinput.Blink,
		a.statusTickCmd(),
	)
}

func (a *AutoCoder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds       []tea.Cmd
		promptCmd  tea.Cmd
		spinnerCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if len(a.checkpoints) > 0 && !a.checkpoints[len(a.checkpoints)-1].Done {
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

	case tea.KeyMsg:
		switch msg.Type {
		// quit
		case tea.KeyCtrlC:
			return a, tea.Quit
		// help
		case tea.KeyCtrlH:
			components.prompt, promptCmd = components.prompt.Update(msg)
			cmds = append(
				cmds,
				promptCmd,
				tea.Println(components.renderer.RenderContent(components.renderer.RenderHelpMessage())),
				textinput.Blink,
			)

		// history
		case tea.KeyUp, tea.KeyDown:
			if len(a.checkpoints) <= 0 || (len(a.checkpoints) > 0 && a.checkpoints[len(a.checkpoints)-1].Done) {
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
			if len(input) > 0 {
				a.state.buffer = ""
				a.checkpoints = make([]Checkpoint, 0)
				a.history.Add(input)
				inputPrint := components.prompt.AsString()
				components.prompt.SetValue("")
				components.prompt.Focus()
				components.prompt, promptCmd = components.prompt.Update(msg)
				if a.command.isCommand(input) {
					cmds = append(
						cmds,
						promptCmd,
						tea.Println(inputPrint),
						a.command.run(input),
						a.command.awaitChatCompleted(),
					)
				}
			}

		// clear
		case tea.KeyCtrlL:
			components.prompt, promptCmd = components.prompt.Update(msg)
			cmds = append(
				cmds,
				promptCmd,
				tea.ClearScreen,
				textinput.Blink,
			)

		// reset
		case tea.KeyCtrlR:
			a.reset()
			components.prompt.SetValue("")
			components.prompt, promptCmd = components.prompt.Update(msg)
			cmds = append(
				cmds,
				promptCmd,
				tea.ClearScreen,
				textinput.Blink,
			)

		default:
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

	if started && a.checkpoints[len(a.checkpoints)-1].Error != nil {
		return fmt.Sprintf("%s\n\n%s\n",
			components.renderer.RenderError(fmt.Sprintf("%s", a.checkpoints[len(a.checkpoints)-1].Error)),
			components.prompt.View(),
		)
	}

	if started {
		doneMsg := ""
		for _, s := range a.checkpoints[:len(a.checkpoints)-1] {
			switch s.Type {
			case StatusLoading:
				if !done {
					doneMsg += " ✅  " + s.Desc + "\n"
				}
			case StatusSuccess:
				doneMsg += " " + components.renderer.RenderSuccess(s.Desc) + "\n"
			case StatusWarning:
				doneMsg += " " + components.renderer.RenderWarning(s.Desc) + "\n"
			default:
				doneMsg += " " + s.Desc + "\n"
			}
		}
		if !done {
			return components.spinner.ViewWithMessage(doneMsg, a.state.buffer, a.checkpoints[len(a.checkpoints)-1].Desc)
		}

		return fmt.Sprintf("\n%s\n%s",
			doneMsg,
			components.prompt.View(),
		)
	}

	return components.prompt.View()
}

func (a *AutoCoder) reset() {
	a.checkpoints = make([]Checkpoint, 0)
	a.history.Reset()
	a.absFileNames = make(map[string]struct{})
	a.state.buffer = ""
}
