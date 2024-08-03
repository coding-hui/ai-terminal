package history

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/session"
	"github.com/coding-hui/ai-terminal/internal/util"
	"github.com/coding-hui/ai-terminal/internal/util/templates"
)

var lsHistoryExample = templates.Examples(`
		# Managing session history:
          ai history ls
`)

type historyItem struct {
	title, desc string
}

func (i historyItem) Title() string       { return i.title }
func (i historyItem) Description() string { return i.desc }
func (i historyItem) FilterValue() string { return i.title }

type listModel struct {
	list list.Model

	itemStyle, quitTextStyle lipgloss.Style
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := m.itemStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	return m.itemStyle.Render(m.list.View())
}

type ls struct {
	model     *options.ModelOptions
	datastore *options.DataStoreOptions
}

// newLs returns initialized ls.
func newLs(model *options.ModelOptions, datastore *options.DataStoreOptions) *ls {
	return &ls{
		model:     model,
		datastore: datastore,
	}
}

func newCmdLsHistory(model *options.ModelOptions, datastore *options.DataStoreOptions) *cobra.Command {
	o := newLs(model, datastore)
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "show chat session history.",
		Long:    "show chat session history.",
		Example: lsHistoryExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run(args))
		},
		PostRunE: func(c *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = viper.BindPFlag(flag.Name, flag)
	})

	return cmd
}

// Validate validates the provided options.
func (o *ls) Validate() error {
	return nil
}

// Run executes history command.
func (o *ls) Run(_ []string) error {
	cfg, err := options.NewConfig()
	if err != nil {
		return err
	}
	engine, err := llm.NewLLMEngine(llm.ChatEngineMode, cfg)
	if err != nil {
		return err
	}

	chatHistory, err := session.GetHistoryStore(*cfg, llm.ChatEngineMode.String())
	if err != nil {
		return err
	}

	allSession, err := chatHistory.Sessions(context.Background())
	if err != nil {
		return err
	}

	var (
		items []list.Item
		mutex sync.Mutex
		wg    sync.WaitGroup
	)
	for _, sessionId := range allSession {
		wg.Add(1)
		go func(sessionId string) {
			defer wg.Done()

			messages, err := chatHistory.Messages(context.Background(), sessionId)
			if err != nil {
				klog.Error(err)
			}

			summary, err := engine.SummaryMessages(messages)
			if err != nil {
				klog.Error(err)
			}

			mutex.Lock()
			defer mutex.Unlock()
			items = append(items, historyItem{
				title: sessionId,
				desc:  summary,
			})
		}(sessionId)
	}
	wg.Wait()

	m := listModel{
		list:          list.New(items, list.NewDefaultDelegate(), 0, 0),
		itemStyle:     lipgloss.NewStyle().Margin(1, 2),
		quitTextStyle: lipgloss.NewStyle().Margin(1, 0, 2, 4),
	}
	m.list.Title = "Chat History"

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return nil
}
