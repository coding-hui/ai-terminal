package commit

import (
	"html"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/prompt"
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/ui"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type Options struct {
	commitMsgFile  string
	preview        bool
	diffUnified    int
	excludeList    []string
	templateFile   string
	templateString string
	commitAmend    bool
	noConfirm      bool
	commitLang     string
	userPrompt     string
	commitPrefix   string

	cfg *options.Config
	genericclioptions.IOStreams

	FilesToAdd []string
}

// Option defines a function type for configuring Options
type Option func(*Options)

// WithNoConfirm sets the noConfirm flag
func WithNoConfirm(noConfirm bool) Option {
	return func(o *Options) {
		o.noConfirm = noConfirm
	}
}

// WithFilesToAdd sets the files to add
func WithFilesToAdd(files []string) Option {
	return func(o *Options) {
		o.FilesToAdd = files
	}
}

// WithIOStreams sets the IO streams
func WithIOStreams(ioStreams genericclioptions.IOStreams) Option {
	return func(o *Options) {
		o.IOStreams = ioStreams
	}
}

// WithConfig sets the configuration
func WithConfig(cfg *options.Config) Option {
	return func(o *Options) {
		o.cfg = cfg
	}
}

// WithCommitPrefix sets the commit prefix
func WithCommitPrefix(prefix string) Option {
	return func(o *Options) {
		o.commitPrefix = prefix
	}
}

// New creates a new Options instance with optional configurations
func New(opts ...Option) *Options {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// NewCmdCommit returns a cobra command for commit msg.
func NewCmdCommit(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	ops := New(
		WithNoConfirm(false),
		WithFilesToAdd([]string{}),
		WithIOStreams(ioStreams),
		WithConfig(cfg),
		WithCommitPrefix(""), // Initialize with empty prefix
	)

	commitCmd := &cobra.Command{
		Use:   "commit",
		Short: "Auto generate commit message",
		RunE:  ops.AutoCommit,
	}

	commitCmd.Flags().StringVarP(&ops.commitMsgFile, "file", "f", "", "File to store the generated commit message")
	commitCmd.Flags().BoolVar(&ops.preview, "preview", false, "Preview the commit message before committing")
	commitCmd.Flags().IntVar(&ops.diffUnified, "diff-unified", 3, "Number of lines of context to show in diffs (e.g., 3)")
	commitCmd.Flags().StringSliceVar(&ops.excludeList, "exclude-list", []string{}, "List of files to exclude from the diff (e.g., '*.lock')")
	commitCmd.Flags().StringVar(&ops.templateFile, "template-file", "", "File containing the template for the commit message")
	commitCmd.Flags().StringVar(&ops.templateString, "template-string", "", "Inline template string for the commit message")
	commitCmd.Flags().BoolVar(&ops.commitAmend, "amend", false, "Amend the most recent commit")
	commitCmd.Flags().BoolVar(&ops.noConfirm, "no-confirm", false, "Skip the confirmation prompt before committing")
	commitCmd.Flags().StringVar(&ops.commitLang, "lang", "en", "Language for summarizing the commit message (e.g., 'zh-cn', 'en', 'zh-tw', 'ja', 'pt', 'pt-br')")
	commitCmd.Flags().StringSliceVar(&ops.FilesToAdd, "add", []string{}, "Files to add to the commit (e.g., 'file1.txt file2.txt')")
	commitCmd.Flags().StringVar(&ops.commitPrefix, "prefix", "", "Specify conventional commit prefix (e.g., 'feat', 'fix', 'docs', 'style', 'refactor', 'test', 'chore')")

	return commitCmd
}

func (o *Options) AutoCommit(_ *cobra.Command, args []string) error {
	if !runner.IsCommandAvailable("git") {
		return errbook.New("git command not found on your system's PATH. Please install Git and try again")
	}

	o.userPrompt = ""
	if len(args) > 0 {
		o.userPrompt = strings.TrimSpace(strings.Join(args, " "))
	}

	llmEngine, err := llm.NewLLMEngine(llm.ChatEngineMode, o.cfg)
	if err != nil {
		return err
	}

	g := git.New(
		git.WithDiffUnified(o.diffUnified),
		git.WithExcludeList(o.excludeList),
		git.WithEnableAmend(o.commitAmend),
	)

	// Add files specified by the user
	if len(o.FilesToAdd) > 0 {
		err := g.AddFiles(o.FilesToAdd)
		if err != nil {
			return errbook.Wrap("Could not add files.", err)
		}
	}

	diff, err := g.DiffFiles()
	if err != nil {
		return errbook.Wrap("Could not get diff files.", err)
	}

	vars := map[string]any{
		prompt.FileDiffsKey:         diff,
		prompt.UserAdditionalPrompt: o.userPrompt,
		prompt.OutputLanguageKey:    prompt.GetLanguage(o.commitLang),
	}

	err = o.codeReview(llmEngine, vars)
	if err != nil {
		return errbook.Wrap("Could not generate code review.", err)
	}

	err = o.summarizeTitle(llmEngine, vars)
	if err != nil {
		return errbook.Wrap("Could not generate summarize title.", err)
	}

	// If prefix is specified, use it directly
	if o.commitPrefix != "" {
		vars[prompt.SummarizePrefixKey] = strings.ToLower(o.commitPrefix)
	} else {
		// Otherwise generate prefix from LLM
		err = o.summarizePrefix(llmEngine, vars)
		if err != nil {
			return errbook.Wrap("Could not generate summarize prefix.", err)
		}
	}

	commitMessage, err := o.generateCommitMsg(llmEngine, vars)
	if err != nil {
		return errbook.Wrap("Could not generate commit message.", err)
	}

	if o.commitMsgFile == "" {
		out, err := g.GitDir()
		if err != nil {
			return errbook.Wrap("Could not get git dir.", err)
		}
		o.commitMsgFile = path.Join(strings.TrimSpace(out), "COMMIT_EDITMSG")
	}
	console.RenderCommitStep("Writing commit message to %s", o.commitMsgFile)
	err = os.WriteFile(o.commitMsgFile, []byte(commitMessage), 0o600)
	if err != nil {
		return errbook.Wrap("Could not write commit message to file: "+o.commitMsgFile, err)
	}

	if o.preview && !o.noConfirm {
		input := confirmation.New("Commit preview summary?", confirmation.Yes)
		ready, err := input.RunPrompt()
		if err != nil {
			return errbook.Wrap("Could not run prompt.", err)
		}
		if !ready {
			return nil
		}
	}

	if !o.noConfirm {
		input := confirmation.New("Do you want to change the commit message?", confirmation.No)
		change, err := input.RunPrompt()
		if err != nil {
			return errbook.Wrap("Could not run prompt.", err)
		}

		if change {
			m := ui.InitialTextareaPrompt(commitMessage)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return errbook.Wrap("Could not start Bubble Tea program.", err)
			}
			p.Wait()
			commitMessage = m.Textarea.Value()
		}
	}

	// git commit automatically
	console.RenderCommitStep("Recording changes to repository...")
	output, err := g.Commit(commitMessage)
	if err != nil {
		return errbook.Wrap("Could not commit changes to the repository.", err)
	}
	color.Yellow(output)

	return nil
}

// codeReview summary code review message from diff datas
func (o *Options) codeReview(engine *llm.Engine, vars map[string]any) error {
	console.RenderCommitStep("Analyzing code changes...")

	p, err := prompt.GetPromptStringByTemplateName(prompt.SummarizeFileDiffTemplate, vars)
	if err != nil {
		return err
	}

	resp, err := engine.CreateCompletion(strings.TrimSpace(p))
	if err != nil {
		return err
	}
	codeReviewResult := strings.TrimSpace(resp.Explanation)
	vars[prompt.SummarizePointsKey] = codeReviewResult
	vars[prompt.SummarizeMessageKey] = codeReviewResult

	return nil
}

func (o *Options) summarizeTitle(engine *llm.Engine, vars map[string]any) error {
	console.RenderCommitStep("Generating commit title...")

	p, err := prompt.GetPromptStringByTemplateName(prompt.SummarizeTitleTemplate, vars)
	if err != nil {
		return err
	}

	resp, err := engine.CreateCompletion(strings.TrimSpace(p))
	if err != nil {
		return err
	}
	summarizeTitle := resp.Explanation
	summarizeTitle = strings.TrimRight(strings.ToLower(string(summarizeTitle[0]))+summarizeTitle[1:], ".")

	vars[prompt.SummarizeTitleKey] = summarizeTitle

	return nil
}

func (o *Options) summarizePrefix(engine *llm.Engine, vars map[string]any) error {
	console.RenderCommitStep("Determining commit type...")

	p, err := prompt.GetPromptStringByTemplateName(prompt.ConventionalCommitTemplate, vars)
	if err != nil {
		return err
	}

	resp, err := engine.CreateCompletion(strings.TrimSpace(p))
	if err != nil {
		return err
	}

	vars[prompt.SummarizePrefixKey] = strings.ToLower(resp.Explanation)

	return nil
}

func (o *Options) generateCommitMsg(engine *llm.Engine, vars map[string]any) (commitMessage string, err error) {
	if o.templateFile != "" {
		format, err := os.ReadFile(o.templateFile)
		if err != nil {
			return "", err
		}
		commitMessage, err = prompt.GetPromptStringByTemplate(string(format), vars)
		if err != nil {
			return "", err
		}
	} else if o.templateString != "" {
		commitMessage, err = prompt.GetPromptStringByTemplate(o.templateString, vars)
		if err != nil {
			return "", err
		}
	} else {
		commitMessage, err = prompt.GetPromptStringByTemplateName(prompt.CommitMessageTemplate, vars)
		if err != nil {
			return "", err
		}
	}

	if o.commitLang != prompt.DefaultLanguage {
		console.RenderCommitStep("Translating commit message to %s...", o.commitLang)
		translationPrompt, err := prompt.GetPromptStringByTemplateName(prompt.TranslationTemplate, map[string]any{
			prompt.OutputLanguageKey: prompt.GetLanguage(o.commitLang),
			prompt.OutputMessageKey:  commitMessage,
		})
		if err != nil {
			return "", err
		}

		resp, err := engine.CreateCompletion(strings.TrimSpace(translationPrompt))
		if err != nil {
			return "", err
		}
		commitMessage = resp.Explanation
	}

	// unescape html entities in commit message
	commitMessage = html.UnescapeString(commitMessage)
	commitMessage = strings.TrimSpace(commitMessage)

	// Output simplified commit summary
	lines := strings.Split(commitMessage, "\n")
	if len(lines) > 0 {
		console.RenderCommitSuccess("Commit Summary:")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				console.RenderCommitSuccess("  %s", line)
			}
		}
	}

	return commitMessage, nil
}
