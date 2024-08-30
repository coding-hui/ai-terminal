package coders

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	querying   bool
	confirming bool
}

// AutoCoder is an interface for auto-generating code.
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

	files              []string
	currentSuggestions []string
}

func StartAutCoder() error {
	coder := NewAutoCoder()
	program = tea.NewProgram(
		coder,
		// tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)
	if _, err := program.Run(); err != nil {
		fmt.Println("Error running auto chat program:", err)
		os.Exit(1)
	}
	return nil
}

func NewAutoCoder() *AutoCoder {
	var err error
	g := git.New()
	cfg := options.NewConfig()

	autoCoder := &AutoCoder{
		state: State{
			error:  nil,
			buffer: "",
		},
		gitRepo:        g,
		cfg:            cfg,
		checkpoints:    []Checkpoint{},
		history:        ui.NewHistory(),
		absFileNames:   map[string]struct{}{},
		checkpointChan: make(chan Checkpoint),
	}

	autoCoder.llmEngine, err = llm.NewLLMEngine(llm.ChatEngineMode, cfg)
	if err != nil {
		display.FatalErr(err)
	}

	root, err := g.GitDir()
	if err != nil {
		display.FatalErr(err)
	}

	autoCoder.codeBasePath = filepath.Dir(root)
	autoCoder.command = newCommand(autoCoder)

	// 获取所有文件
	files, err := autoCoder.gitRepo.ListAllFiles()
	if err != nil {
		display.FatalErr(err)
	}
	autoCoder.files = files

	// 初始时只设置命令作为建议
	components.prompt.SetSuggestions(getSupportedCommands())

	return autoCoder
}

func (a *AutoCoder) Init() tea.Cmd {
	components.prompt.SetSuggestions(getSupportedCommands())

	return tea.Sequence(
		tea.ClearScreen,
		tea.Println(components.renderer.RenderContent(components.renderer.RenderWelcomeMessage(a.cfg.System.GetUsername()))),
		textinput.Blink,
		a.statusTickCmd(),
	)
}

//nolint:golint,gocyclo
func (a *AutoCoder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds       []tea.Cmd
		promptCmd  tea.Cmd
		spinnerCmd tea.Cmd
		confirmCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if a.state.querying {
			components.spinner, spinnerCmd = components.spinner.Update(msg)
			cmds = append(
				cmds,
				spinnerCmd,
			)
		}

	case tea.WindowSizeMsg:
		components.width = msg.Width
		components.height = msg.Height
		components.prompt.SetWidth(msg.Width)
		if a.state.confirming {
			components.confirm.SetWidth(msg.Width)
			components.confirm.SetHeight(msg.Height - components.prompt.Height())
			components.confirm.GotoBottom()
		}

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
			if !a.state.querying && !a.state.confirming {
				components.prompt.SetValue("")
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
			if !a.state.querying && !a.state.confirming {
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
			if !a.state.querying && !a.state.confirming && input != "" {
				a.state.buffer = ""
				a.checkpoints = make([]Checkpoint, 0)
				a.history.Add(input)
				inputPrint := components.prompt.AsString()
				components.prompt.SetValue("")
				components.prompt.Blur()
				components.prompt, promptCmd = components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.Println(inputPrint),
					a.command.run(input),
					a.command.awaitChatCompleted(),
				)
				components.prompt.Focus()
			}

		// clear
		case tea.KeyCtrlL:
			if !a.state.querying && !a.state.confirming {
				a.checkpoints = make([]Checkpoint, 0)
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
			if !a.state.querying && !a.state.confirming {
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

			// 更新建议
			currentInput := components.prompt.GetValue()
			a.updateSuggestions(currentInput)

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
	if components.width == 0 || components.height == 0 {
		return "Initializing..."
	}

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

func (a *AutoCoder) updateSuggestions(input string) {
	a.currentSuggestions = a.filterSuggestions(input)

	var suggestionsToSet []string

	parts := strings.Fields(input)
	if len(parts) == 0 {
		// 如果输入为空，显示所有命令
		suggestionsToSet = getSupportedCommands()
	} else if strings.HasPrefix(parts[0], "/") {
		// 如果输入以 "/" 开头
		if len(parts) == 1 {
			// 只有命令部分，设置匹配的命令作为建议
			suggestionsToSet = a.getMatchingCommands(input)
		} else {
			// 命令后面还有内容，设置匹配的文件作为建议
			fileInput := parts[len(parts)-1]
			matchingFiles := a.getMatchingFiles(fileInput)
			prefix := strings.Join(parts[:len(parts)-1], " ") + " "
			for _, file := range matchingFiles {
				suggestionsToSet = append(suggestionsToSet, prefix+file)
			}
		}
	} else {
		// 如果不是以 "/" 开头，设置匹配的文件作为建议
		fileInput := parts[len(parts)-1]
		matchingFiles := a.getMatchingFiles(fileInput)
		suggestionsToSet = matchingFiles
	}

	// 设置建议
	components.prompt.SetSuggestions(suggestionsToSet)
}

func (a *AutoCoder) filterSuggestions(input string) []string {
	if input == "" {
		return getSupportedCommands()
	}

	var filtered []string
	parts := strings.Fields(input)

	if strings.HasPrefix(parts[0], "/") {
		cmd := parts[0]
		filtered = append(filtered, a.getMatchingCommands(cmd)...)

		// 对于命令后的每个参数，都尝试匹配文件
		for i := 1; i < len(parts); i++ {
			matchingFiles := a.getMatchingFiles(parts[i])
			prefix := strings.Join(parts[:i], " ") + " "
			for _, file := range matchingFiles {
				filtered = append(filtered, prefix+file)
			}
		}
	} else {
		filtered = append(filtered, a.getMatchingCommands(input)...)
		filtered = append(filtered, a.getMatchingFiles(parts[len(parts)-1])...)
	}

	// 排序和限制数量的逻辑
	sort.Slice(filtered, func(i, j int) bool {
		iIsCmd := strings.HasPrefix(filtered[i], "/")
		jIsCmd := strings.HasPrefix(filtered[j], "/")
		if iIsCmd && !jIsCmd {
			return true
		}
		if !iIsCmd && jIsCmd {
			return false
		}
		return filtered[i] < filtered[j]
	})

	maxSuggestions := 10
	if len(filtered) > maxSuggestions {
		filtered = filtered[:maxSuggestions]
	}

	return filtered
}

// 获取匹配的命令
func (a *AutoCoder) getMatchingCommands(input string) []string {
	var matched []string
	for _, cmd := range getSupportedCommands() {
		if strings.HasPrefix(cmd, input) {
			matched = append(matched, cmd)
		}
	}
	return matched
}

// 获取匹配的文件
func (a *AutoCoder) getMatchingFiles(input string) []string {
	var matched []string
	inputParts := strings.Split(input, "/")

	for _, file := range a.files {

		// 检查是否所有输入部分都匹配文件路径的某个部分
		allPartsMatch := true
		for _, inputPart := range inputParts {
			partMatched := false
			if strings.Contains(file, inputPart) {
				partMatched = true
				break
			}
			if !partMatched {
				allPartsMatch = false
				break
			}
		}

		if allPartsMatch {
			matched = append(matched, file)
		}
	}

	// 对匹配结果进行排序，使更相关的结果排在前面
	sort.Slice(matched, func(i, j int) bool {
		iRelevance := getRelevanceScore(input, matched[i])
		jRelevance := getRelevanceScore(input, matched[j])
		if iRelevance == jRelevance {
			return matched[i] < matched[j] // 字母顺序作为次要排序标准，保留大小写
		}
		return iRelevance > jRelevance
	})

	return matched
}

// 计算匹配相关度分数
func getRelevanceScore(input, file string) int {
	score := 0

	// 如果是精确前缀匹配，给予最高分
	if strings.HasPrefix(file, input) {
		score += 100
	}

	// 根据匹配位置给予不同的分数
	index := strings.Index(file, input)
	if index != -1 {
		score += 50 - index // 匹配位置越靠前，分数越高
	}

	// 计算匹配的字符数
	matchedChars := 0
	for _, ch := range input {
		if strings.ContainsRune(file, ch) {
			matchedChars++
		}
	}
	score += matchedChars

	return score
}

// 添加一个新方法来获取当前的建议
func (a *AutoCoder) GetCurrentSuggestions() []string {
	return a.currentSuggestions
}
