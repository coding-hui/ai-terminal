package commit

import (
	"errors"
	"html"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coding-hui/iam/pkg/cli/genericclioptions"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/prompt"
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/ui"
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

	genericclioptions.IOStreams
}

// NewCmdCommit returns a cobra command for commit msg.
func NewCmdCommit(ioStreams genericclioptions.IOStreams) *cobra.Command {
	ops := &Options{
		noConfirm: false,
		IOStreams: ioStreams,
	}

	commitCmd := &cobra.Command{
		Use:   "commit",
		Short: "Auto generate commit message",
		RunE:  ops.autoCommit,
	}

	commitCmd.Flags().StringVarP(&ops.commitMsgFile, "file", "f", "", "commit message file")
	commitCmd.Flags().BoolVar(&ops.preview, "preview", false, "preview commit message")
	commitCmd.Flags().IntVar(&ops.diffUnified, "diff-unified", 3, "generate diffs with <n> lines of context, default is 3")
	commitCmd.Flags().StringSliceVar(&ops.excludeList, "exclude-list", []string{}, "exclude file from git diff command")
	commitCmd.Flags().StringVar(&ops.templateFile, "template-file", "", "git commit message file")
	commitCmd.Flags().StringVar(&ops.templateString, "template-string", "", "git commit message string")
	commitCmd.Flags().BoolVar(&ops.commitAmend, "amend", false, "replace the tip of the current branch by creating a new commit.")
	commitCmd.Flags().BoolVar(&ops.noConfirm, "no-confirm", false, "skip confirmation prompt")

	return commitCmd
}

func (o *Options) autoCommit(cmd *cobra.Command, args []string) error {
	if !runner.IsCommandAvailable("git") {
		return errors.New("git command not found on your system's PATH. Please install Git and try again")
	}

	llmEngine, err := llm.NewLLMEngine(llm.ChatEngineMode, options.NewConfig())
	if err != nil {
		return err
	}

	g := git.New(
		git.WithDiffUnified(o.diffUnified),
		git.WithExcludeList(o.excludeList),
		git.WithEnableAmend(o.commitAmend),
	)
	diff, err := g.DiffFiles()
	if err != nil {
		return err
	}

	vars := map[string]any{
		prompt.FileDiffsKey: diff,
	}

	err = o.codeReview(llmEngine, vars)
	if err != nil {
		return err
	}

	err = o.summarizeTitle(llmEngine, vars)
	if err != nil {
		return err
	}

	err = o.summarizePrefix(llmEngine, vars)
	if err != nil {
		return err
	}

	commitMessage, err := o.generateCommitMsg(llmEngine, vars)
	if err != nil {
		return err
	}

	if o.commitMsgFile == "" {
		out, err := g.GitDir()
		if err != nil {
			return err
		}
		o.commitMsgFile = path.Join(strings.TrimSpace(out), "COMMIT_EDITMSG")
	}
	color.Cyan("Write the commit message to " + o.commitMsgFile + " file")
	err = os.WriteFile(o.commitMsgFile, []byte(commitMessage), 0o600)
	if err != nil {
		return err
	}

	if o.preview && !o.noConfirm {
		input := confirmation.New("Commit preview summary?", confirmation.Yes)
		ready, err := input.RunPrompt()
		if err != nil {
			return err
		}
		if !ready {
			return nil
		}
	}

	if !o.noConfirm {
		input := confirmation.New("Do you want to change the commit message?", confirmation.No)
		change, err := input.RunPrompt()
		if err != nil {
			return err
		}

		if change {
			m := ui.InitialTextareaPrompt(commitMessage)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return err
			}
			p.Wait()
			commitMessage = m.Textarea.Value()
		}
	}

	// git commit automatically
	color.Cyan("Git record changes to the repository")
	output, err := g.Commit(commitMessage)
	if err != nil {
		return err
	}
	color.Yellow(output)

	return nil
}

// codeReview summary code review message from diff datas
func (o *Options) codeReview(engine *llm.Engine, vars map[string]any) error {
	color.Cyan("We are trying to summarize a git diff")

	p, err := prompt.GetPromptStringByTemplateName(prompt.SummarizeFileDiffTemplate, vars)
	if err != nil {
		return err
	}

	resp, err := engine.ExecCompletion(strings.TrimSpace(p))
	if err != nil {
		return err
	}
	codeReviewResult := strings.TrimSpace(resp.Explanation)
	vars[prompt.SummarizePointsKey] = codeReviewResult
	vars[prompt.SummarizeMessageKey] = codeReviewResult

	return nil
}

func (o *Options) summarizeTitle(engine *llm.Engine, vars map[string]any) error {
	color.Cyan("We are trying to summarize a title for pull request")

	p, err := prompt.GetPromptStringByTemplateName(prompt.SummarizeTitleTemplate, vars)
	if err != nil {
		return err
	}

	resp, err := engine.ExecCompletion(strings.TrimSpace(p))
	if err != nil {
		return err
	}
	summarizeTitle := resp.Explanation
	summarizeTitle = strings.TrimRight(strings.ToLower(string(summarizeTitle[0]))+summarizeTitle[1:], ".")

	vars[prompt.SummarizeTitleKey] = summarizeTitle

	return nil
}

func (o *Options) summarizePrefix(engine *llm.Engine, vars map[string]any) error {
	message := "We are trying to get conventional commit prefix"
	color.Cyan(message + " (Tools)")

	p, err := prompt.GetPromptStringByTemplateName(prompt.ConventionalCommitTemplate, vars)
	if err != nil {
		return err
	}

	resp, err := engine.ExecCompletion(strings.TrimSpace(p))
	if err != nil {
		return err
	}

	vars[prompt.SummarizePrefixKey] = resp.Explanation

	return nil
}

func (o *Options) generateCommitMsg(engine *llm.Engine, vars map[string]any) (string, error) {
	color.Cyan("We are trying to translate a git commit message to " + prompt.GetLanguage(viper.GetString("output.lang")) + " language")

	commitMessage, err := prompt.GetPromptStringByTemplateName(prompt.CommitMessageTemplate, vars)
	if err != nil {
		return "", err
	}

	vars[prompt.OutputLanguageKey] = prompt.GetLanguage(viper.GetString("output.lang"))
	vars[prompt.OutputMessageKey] = commitMessage
	translationPrompt, err := prompt.GetPromptStringByTemplateName(prompt.TranslationTemplate, vars)
	if err != nil {
		return "", err
	}

	resp, err := engine.ExecCompletion(strings.TrimSpace(translationPrompt))
	if err != nil {
		return "", err
	}

	// unescape html entities in commit message
	commitMessage = html.UnescapeString(resp.Explanation)
	commitMessage = strings.TrimSpace(commitMessage)

	// Output commit summary data from AI
	color.Yellow("================Commit Summary====================")
	color.Yellow("\n" + commitMessage + "\n\n")
	color.Yellow("==================================================")

	return commitMessage, nil
}
