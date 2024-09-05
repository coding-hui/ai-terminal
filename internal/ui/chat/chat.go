package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"k8s.io/klog/v2"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/ui"
)

const (
	defaultWidth  = 120
	defaultHeight = 120
)

type State struct {
	error       error
	runMode     ui.RunMode
	promptMode  ui.PromptMode
	configuring bool
	querying    bool
	confirming  bool
	executing   bool
	args        string
	pipe        string
	buffer      string
	command     string
}

type Dimensions struct {
	width  int
	height int
}

type Components struct {
	prompt   *Prompt
	renderer *Renderer
	spinner  *ui.Spinner
}

type Ui struct {
	state           State
	dimensions      Dimensions
	components      Components
	config          *options.Config
	waitForUserChan chan struct{}
	engine          *llm.Engine
	history         *ui.History
}

func NewUi(input *ui.Input) *Ui {
	return &Ui{
		state: State{
			error:       nil,
			runMode:     input.GetRunMode(),
			promptMode:  input.GetPromptMode(),
			configuring: false,
			querying:    false,
			confirming:  false,
			executing:   false,
			args:        input.GetArgs(),
			pipe:        input.GetPipe(),
			buffer:      "",
			command:     "",
		},
		dimensions: Dimensions{
			defaultWidth,
			defaultHeight,
		},
		components: Components{
			prompt: NewPrompt(input.GetPromptMode()),
			renderer: NewRenderer(
				glamour.WithEmoji(),
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(defaultWidth),
			),
			spinner: ui.NewSpinner(),
		},
		history:         ui.NewHistory(),
		waitForUserChan: make(chan struct{}, 1),
	}
}

func (u *Ui) Init() tea.Cmd {
	cfg := options.NewConfig()
	u.config = cfg
	klog.V(2).InfoS("begin init tea model.", "cfg", cfg)
	if cfg.Ai.Token == "" || cfg.Ai.Model == "" || cfg.Ai.ApiBase == "" {
		if u.state.runMode == ui.ReplMode {
			return tea.Sequence(
				tea.ClearScreen,
				u.startConfig(ui.ModelConfigPromptMode),
				u.startConfig(ui.ApiBaseConfigPromptMode),
				u.startConfig(ui.TokenConfigPromptMode),
			)
		} else {
			return tea.Sequence(
				u.startConfig(ui.ModelConfigPromptMode),
				u.startConfig(ui.ApiBaseConfigPromptMode),
				u.startConfig(ui.TokenConfigPromptMode),
			)
		}
	}
	if u.state.runMode == ui.ReplMode {
		return u.startRepl(cfg)
	} else {
		return u.startCli(cfg)
	}
}

//nolint:golint,gocyclo
func (u *Ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds       []tea.Cmd
		promptCmd  tea.Cmd
		spinnerCmd tea.Cmd
	)

	switch msg := msg.(type) {
	// spinner
	case spinner.TickMsg:
		if u.state.querying {
			u.components.spinner, spinnerCmd = u.components.spinner.Update(msg)
			cmds = append(
				cmds,
				spinnerCmd,
			)
		}
	// size
	case tea.WindowSizeMsg:
		u.dimensions.width = msg.Width
		u.dimensions.height = msg.Height
		u.components.prompt.input.Width = msg.Width
		u.components.prompt.input.SetCursor(len(u.components.prompt.input.Value()) - 1)
		u.components.renderer = NewRenderer(
			glamour.WithEmoji(),
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(u.dimensions.width),
		)
	// keyboard
	case tea.KeyMsg:
		switch msg.Type {
		// quit
		case tea.KeyCtrlC:
			return u, tea.Quit
		// history
		case tea.KeyUp, tea.KeyDown:
			if !u.state.querying && !u.state.confirming {
				var input *string
				if msg.Type == tea.KeyUp {
					input = u.history.GetPrevious()
				} else {
					input = u.history.GetNext()
				}
				if input != nil {
					u.components.prompt.SetValue(*input)
					u.components.prompt, promptCmd = u.components.prompt.Update(msg)
					cmds = append(
						cmds,
						promptCmd,
					)
				}
			}
		// switch mode
		case tea.KeyTab:
			if !u.state.querying && !u.state.confirming {
				if u.state.promptMode == ui.ChatPromptMode {
					u.state.promptMode = ui.ExecPromptMode
					u.components.prompt.SetMode(ui.ExecPromptMode)
					u.engine.SetMode(llm.ExecEngineMode)
				} else {
					u.state.promptMode = ui.ChatPromptMode
					u.components.prompt.SetMode(ui.ChatPromptMode)
					u.engine.SetMode(llm.ChatEngineMode)
				}
				u.engine.Reset()
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					textinput.Blink,
				)
			}
		// enter
		case tea.KeyEnter:
			promptVal := u.components.prompt.GetValue()
			if u.state.configuring && promptVal != "" {
				return u, u.finishConfig(promptVal)
			}
			if !u.state.querying && !u.state.confirming {
				input := u.components.prompt.GetValue()
				if input != "" {
					inputPrint := u.components.prompt.AsString()
					u.history.Add(input)
					u.components.prompt.SetValue("")
					u.components.prompt.Blur()
					u.components.prompt, promptCmd = u.components.prompt.Update(msg)
					if u.state.promptMode == ui.ChatPromptMode {
						cmds = append(
							cmds,
							promptCmd,
							tea.Println(inputPrint),
							u.startChatStream(input),
							u.awaitChatStream(),
						)
					} else {
						cmds = append(
							cmds,
							promptCmd,
							tea.Println(inputPrint),
							u.startExec(input),
							u.components.spinner.Tick,
						)
					}
				}
			}

		// help
		case tea.KeyCtrlH:
			if !u.state.configuring && !u.state.querying && !u.state.confirming {
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.Println(u.components.renderer.RenderContent(u.components.renderer.RenderHelpMessage())),
					textinput.Blink,
				)
			}

		// clear
		case tea.KeyCtrlL:
			if !u.state.querying && !u.state.confirming {
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.ClearScreen,
					textinput.Blink,
				)
			}

		// reset
		case tea.KeyCtrlR:
			if !u.state.querying && !u.state.confirming {
				u.history.Reset()
				u.engine.Reset()
				u.components.prompt.SetValue("")
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.ClearScreen,
					textinput.Blink,
				)
			}

		// edit settings
		case tea.KeyCtrlS:
			if !u.state.querying && !u.state.confirming && !u.state.configuring && !u.state.executing {
				u.state.executing = true
				u.state.buffer = ""
				u.state.command = ""
				u.components.prompt.Blur()
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					u.editSettings(),
				)
			}

		default:
			if u.state.confirming {
				if strings.ToLower(msg.String()) == "y" {
					u.state.confirming = false
					u.state.executing = true
					u.state.buffer = ""
					u.components.prompt.SetValue("")
					return u, tea.Sequence(
						promptCmd,
						u.execCommand(u.state.command),
					)
				} else {
					u.state.confirming = false
					u.state.executing = false
					u.state.buffer = ""
					u.components.prompt, promptCmd = u.components.prompt.Update(msg)
					u.components.prompt.SetValue("")
					u.components.prompt.Focus()
					if u.state.runMode == ui.ReplMode {
						cmds = append(
							cmds,
							promptCmd,
							tea.Println(fmt.Sprintf("\n%s\n", u.components.renderer.RenderWarning("[cancel]"))),
							textinput.Blink,
						)
					} else {
						return u, tea.Sequence(
							promptCmd,
							tea.Println(fmt.Sprintf("\n%s\n", u.components.renderer.RenderWarning("[cancel]"))),
							tea.Quit,
						)
					}
				}
				u.state.command = ""
			} else {
				u.components.prompt.Focus()
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					textinput.Blink,
				)
			}
		}
	// engine exec feedback
	case llm.EngineExecOutput:
		var output string
		if msg.IsExecutable() {
			u.state.confirming = true
			u.state.command = msg.GetCommand()
			output = u.components.renderer.RenderContent(fmt.Sprintf("`%s`", u.state.command))
			output += fmt.Sprintf("  %s\n\n  confirm execution? [y/N]", u.components.renderer.RenderHelp(msg.GetExplanation()))
			u.components.prompt.Blur()
		} else {
			output = u.components.renderer.RenderContent(msg.GetExplanation())
			u.components.prompt.Focus()
			if u.state.runMode == ui.CliMode {
				return u, tea.Sequence(
					tea.Println(output),
					tea.Quit,
				)
			}
		}
		u.components.prompt, promptCmd = u.components.prompt.Update(msg)
		return u, tea.Sequence(
			promptCmd,
			textinput.Blink,
			tea.Println(output),
		)
	// engine chat stream feedback
	case llm.EngineChatStreamOutput:
		if msg.IsLast() {
			output := u.components.renderer.RenderContent(u.state.buffer)
			u.state.buffer = ""
			u.components.prompt.Focus()
			if u.state.runMode == ui.CliMode {
				return u, tea.Sequence(
					tea.Println(output),
					tea.Quit,
				)
			} else {
				return u, tea.Sequence(
					tea.Println(output),
					textinput.Blink,
				)
			}
		} else {
			return u, u.awaitChatStream()
		}
	// runner feedback
	case runner.Output:
		u.state.querying = false
		u.components.prompt, promptCmd = u.components.prompt.Update(msg)
		u.components.prompt.Focus()
		output := u.components.renderer.RenderSuccess(fmt.Sprintf("\n%s\n", msg.GetSuccessMessage()))
		if msg.HasError() {
			output = u.components.renderer.RenderError(fmt.Sprintf("\n%s\n", msg.GetErrorMessage()))
		}
		if u.state.runMode == ui.CliMode {
			return u, tea.Sequence(
				tea.Println(output),
				tea.Quit,
			)
		} else {
			return u, tea.Sequence(
				tea.Println(output),
				promptCmd,
				textinput.Blink,
			)
		}
	// errors
	case error:
		u.state.error = msg
		return u, nil
	}

	return u, tea.Batch(cmds...)
}

func (u *Ui) View() string {
	if u.state.error != nil {
		return u.components.renderer.RenderError(fmt.Sprintf("[ERROR] %s", u.state.error))
	}

	if u.state.configuring {
		return fmt.Sprintf(
			"%s\n%s",
			u.components.renderer.RenderContent(u.state.buffer),
			u.components.prompt.View(),
		)
	}

	if !u.state.querying && !u.state.confirming && !u.state.executing {
		return u.components.prompt.View()
	}

	if u.state.promptMode == ui.ChatPromptMode {
		if u.state.querying && len(u.state.buffer) <= 0 {
			return u.components.spinner.View()
		}
		return u.components.renderer.RenderContent(u.state.buffer)
	} else {
		if u.state.querying {
			return u.components.spinner.View()
		} else {
			if !u.state.executing {
				return u.components.renderer.RenderContent(u.state.buffer)
			}
		}
	}

	return ""
}

func (u *Ui) startRepl(config *options.Config) tea.Cmd {
	return tea.Sequence(
		tea.ClearScreen,
		tea.Println(u.components.renderer.RenderContent(u.components.renderer.RenderHelpMessage())),
		textinput.Blink,
		func() tea.Msg {
			u.config = config

			if u.state.promptMode == ui.DefaultPromptMode {
				u.state.promptMode = ui.GetPromptModeFromString(config.DefaultPromptMode)
			}

			engineMode := llm.ExecEngineMode
			if u.state.promptMode == ui.ChatPromptMode {
				engineMode = llm.ChatEngineMode
			}

			engine, err := llm.NewLLMEngine(engineMode, config)
			if err != nil {
				return err
			}

			if u.state.pipe != "" {
				engine.SetPipe(u.state.pipe)
			}

			u.engine = engine
			u.state.buffer = "Welcome! **" + u.config.System.GetUsername() + "** 👋  \n\n"
			u.state.command = ""
			u.components.prompt = NewPrompt(u.state.promptMode)

			return nil
		},
	)
}

func (u *Ui) startCli(config *options.Config) tea.Cmd {
	u.config = config

	if u.state.promptMode == ui.DefaultPromptMode {
		u.state.promptMode = ui.GetPromptModeFromString(config.DefaultPromptMode)
	}

	engineMode := llm.ExecEngineMode
	if u.state.promptMode == ui.ChatPromptMode {
		engineMode = llm.ChatEngineMode
	}

	engine, err := llm.NewLLMEngine(engineMode, config)
	if err != nil {
		u.state.error = err
		return nil
	}

	if u.state.pipe != "" {
		engine.SetPipe(u.state.pipe)
	}

	u.engine = engine
	u.state.querying = true
	u.state.confirming = false
	u.state.buffer = ""
	u.state.command = ""

	if u.state.promptMode == ui.ExecPromptMode {
		return tea.Sequence(
			u.components.spinner.Tick,
			func() tea.Msg {
				output, err := u.engine.ExecCompletion(u.state.args)
				u.state.querying = false
				if err != nil {
					return err
				}

				return *output
			},
		)
	} else {
		return tea.Sequence(
			u.components.spinner.Tick,
			u.startChatStream(u.state.args),
			u.awaitChatStream(),
		)
	}
}

func (u *Ui) startConfig(configPromptMode ui.PromptMode) tea.Cmd {
	return func() tea.Msg {
		u.state.configuring = true
		u.state.querying = false
		u.state.confirming = false
		u.state.executing = false

		switch configPromptMode {
		case ui.ModelConfigPromptMode:
			u.state.buffer = u.components.renderer.RenderConfigMessage(u.config.System.GetUsername())
		case ui.ApiBaseConfigPromptMode:
			u.state.buffer = u.components.renderer.RenderApiBaseConfigMessage()
		default:
			u.state.buffer = u.components.renderer.RenderApiTokenConfigMessage()
		}

		u.state.command = ""
		u.components.prompt = NewPrompt(configPromptMode)

		<-u.waitForUserChan

		return nil
	}
}

func (u *Ui) finishConfig(key string) tea.Cmd {
	switch u.components.prompt.mode {
	case ui.ModelConfigPromptMode:
		u.config.Ai.Model = key
	case ui.ApiBaseConfigPromptMode:
		u.config.Ai.ApiBase = key
	default:
		u.config.Ai.Token = key
	}

	finished := false
	if u.config.Ai.Model != "" && u.config.Ai.Token != "" && u.config.Ai.ApiBase != "" {
		finished = true
	}
	if !finished {
		u.waitForUserChan <- struct{}{}
		return tea.ClearScreen
	}

	u.state.configuring = false
	config, err := options.WriteConfig(u.config.Ai.Model, u.config.Ai.ApiBase, u.config.Ai.Token, true)
	if err != nil {
		u.state.error = err
		return nil
	}
	u.config = config

	engine, err := llm.NewLLMEngine(llm.ChatEngineMode, config)
	if err != nil {
		u.state.error = err
		return nil
	}

	if u.state.pipe != "" {
		engine.SetPipe(u.state.pipe)
	}

	u.engine = engine

	if u.state.runMode == ui.ReplMode {
		return tea.Sequence(
			tea.ClearScreen,
			tea.Println(u.components.renderer.RenderSuccess("\n[settings ok]\n")),
			textinput.Blink,
			func() tea.Msg {
				u.state.buffer = ""
				u.state.command = ""
				u.components.prompt = NewPrompt(ui.ExecPromptMode)

				return nil
			},
		)
	} else {
		if u.state.promptMode == ui.ExecPromptMode {
			u.state.querying = true
			u.state.configuring = false
			u.state.buffer = ""
			return tea.Sequence(
				tea.Println(u.components.renderer.RenderSuccess("\n[settings ok]")),
				u.components.spinner.Tick,
				func() tea.Msg {
					output, err := u.engine.ExecCompletion(u.state.args)
					u.state.querying = false
					if err != nil {
						return err
					}

					return *output
				},
			)
		} else {
			return tea.Batch(
				u.startChatStream(u.state.args),
				u.awaitChatStream(),
			)
		}
	}
}

func (u *Ui) startExec(input string) tea.Cmd {
	return func() tea.Msg {
		u.state.querying = true
		u.state.confirming = false
		u.state.buffer = ""
		u.state.command = ""

		output, err := u.engine.ExecCompletion(input)
		u.state.querying = false
		if err != nil {
			return err
		}

		return *output
	}
}

func (u *Ui) startChatStream(input string) tea.Cmd {
	return func() tea.Msg {
		u.state.querying = true
		u.state.executing = false
		u.state.confirming = false
		u.state.buffer = ""
		u.state.command = ""

		err := u.engine.ChatStreamCompletion(input)
		if err != nil {
			u.state.querying = false
			return err
		}

		return nil
	}
}

func (u *Ui) awaitChatStream() tea.Cmd {
	return func() tea.Msg {
		output := <-u.engine.GetChannel()
		u.state.buffer += output.GetContent()
		u.state.querying = !output.IsLast()

		return output
	}
}

func (u *Ui) execCommand(input string) tea.Cmd {
	u.state.querying = false
	u.state.confirming = false
	u.state.executing = true

	c := runner.PrepareInteractiveCommand(input)

	return tea.ExecProcess(c, func(error error) tea.Msg {
		u.state.executing = false
		u.state.command = ""

		return runner.NewRunOutput(error, "[ERROR]", "[ok]")
	})
}

func (u *Ui) editSettings() tea.Cmd {
	u.state.querying = false
	u.state.confirming = false
	u.state.executing = true

	c := runner.PrepareEditSettingsCommand(u.config.System.GetEditor(), u.config.System.GetConfigFile())

	return tea.ExecProcess(c, func(err error) tea.Msg {
		u.state.executing = false
		u.state.command = ""

		if err != nil {
			return runner.NewRunOutput(err, "[settings error]", "")
		}

		cfg := options.NewConfig()

		u.config = cfg
		engine, err := llm.NewLLMEngine(llm.ExecEngineMode, cfg)
		if u.state.pipe != "" {
			engine.SetPipe(u.state.pipe)
		}
		if err != nil {
			return runner.NewRunOutput(err, "[settings error]", "")
		}
		u.engine = engine

		return runner.NewRunOutput(nil, "", "[settings ok]")
	})
}
