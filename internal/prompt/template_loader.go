package prompt

import (
	"embed"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"
	"k8s.io/klog/v2"
)

//go:embed templates/*
var templatesFS embed.FS

// Template file names
const (
	CodeReviewTemplate         = "code_review_file_diff.tmpl"
	SummarizeFileDiffTemplate  = "summarize_file_diff.tmpl"
	SummarizeTitleTemplate     = "summarize_title.tmpl"
	ConventionalCommitTemplate = "conventional_commit.tmpl"
	TranslationTemplate        = "translation.tmpl"
	CommitMessageTemplate      = "commit-msg.tmpl"

	SummarizePrefixKey  = "summarize_prefix"
	SummarizeTitleKey   = "summarize_title"
	SummarizeMessageKey = "summarize_message"
	SummarizePointsKey  = "summary_points"
	FileDiffsKey        = "file_diffs"
	OutputLanguageKey   = "output_language"
	OutputMessageKey    = "output_message"
)

type prompt struct {
	inputVars []string
	template  prompts.PromptTemplate
}

var (
	templatesDir    = "templates"
	promptTemplates = map[string]*prompt{
		CodeReviewTemplate: {
			inputVars: []string{FileDiffsKey},
		},
		SummarizeFileDiffTemplate: {
			inputVars: []string{FileDiffsKey},
		},
		SummarizeTitleTemplate: {
			inputVars: []string{SummarizePointsKey},
		},
		ConventionalCommitTemplate: {
			inputVars: []string{SummarizePointsKey},
		},
		TranslationTemplate: {
			inputVars: []string{OutputLanguageKey, OutputMessageKey},
		},
		CommitMessageTemplate: {
			inputVars: []string{SummarizePrefixKey, SummarizeTitleKey, SummarizeMessageKey},
		},
	}
)

// Initializes the prompt package by loading the templates from the embedded file system.
func init() { //nolint:gochecknoinits
	if err := loadPromptTemplates(templatesFS); err != nil {
		klog.Fatal(err)
	}
}

// LoadPromptTemplates loads all the prompt templates found in the templates directory.
func loadPromptTemplates(files embed.FS) error {
	for k, v := range promptTemplates {
		content, err := files.ReadFile(templatesDir + "/" + k)
		if err != nil {
			return err
		}
		v.template = prompts.NewPromptTemplate(string(content), v.inputVars)
	}
	return nil
}
