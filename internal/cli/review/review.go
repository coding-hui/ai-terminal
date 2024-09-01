package review

import (
	"errors"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/prompt"
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type Options struct {
	diffUnified int
	excludeList []string
	commitAmend bool
	commitLang  string

	genericclioptions.IOStreams
}

// NewCmdCommit returns a cobra command for commit msg.
func NewCmdCommit(ioStreams genericclioptions.IOStreams) *cobra.Command {
	ops := &Options{
		IOStreams: ioStreams,
	}

	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "Auto review code changes",
		RunE:  ops.reviewCode,
	}

	reviewCmd.Flags().IntVar(&ops.diffUnified, "diff-unified", 3, "generate diffs with <n> lines of context, default is 3")
	reviewCmd.Flags().StringSliceVar(&ops.excludeList, "exclude-list", []string{}, "exclude file from git diff command")
	reviewCmd.Flags().BoolVar(&ops.commitAmend, "amend", false, "replace the tip of the current branch by creating a new commit.")
	reviewCmd.Flags().StringVar(&ops.commitLang, "lang", "en", "summarizing language uses English by default. "+
		"support en, zh-cn, zh-tw, ja, pt, pt-br.")

	return reviewCmd
}

func (o *Options) reviewCode(cmd *cobra.Command, args []string) error {
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

	vars := map[string]any{prompt.FileDiffsKey: diff}

	reviewPrompt, err := prompt.GetPromptStringByTemplateName(prompt.CodeReviewTemplate, vars)
	if err != nil {
		return err
	}

	// Get summarize comment from diff datas
	color.Cyan("We are trying to review code changes")
	reviewResp, err := llmEngine.ExecCompletion(strings.TrimSpace(reviewPrompt))
	if err != nil {
		return err
	}

	reviewMessage := reviewResp.Explanation
	if prompt.GetLanguage(o.commitLang) != prompt.DefaultLanguage {
		translationPrompt, err := prompt.GetPromptStringByTemplateName(
			prompt.TranslationTemplate, map[string]any{
				prompt.OutputLanguageKey: prompt.GetLanguage(o.commitLang),
				prompt.OutputMessageKey:  reviewMessage,
			},
		)
		if err != nil {
			return err
		}

		color.Cyan("we are trying to translate code review to " + o.commitLang + " language")
		translationResp, err := llmEngine.ExecCompletion(strings.TrimSpace(translationPrompt))
		if err != nil {
			return err
		}
		reviewMessage = translationResp.Explanation
	}

	// Output core review summary
	color.Yellow("================Review Summary====================")
	color.Yellow("\n" + strings.TrimSpace(reviewMessage) + "\n\n")
	color.Yellow("==================================================")

	return nil
}
