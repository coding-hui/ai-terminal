package commit

import (
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/options"
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
	TokenUsage llms.Usage

	// Step token usages
	CodeReviewUsage      llms.Usage
	SummarizeTitleUsage  llms.Usage
	SummarizePrefixUsage llms.Usage
	TranslationUsage     llms.Usage
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

// WithCommitLang sets the commit language
func WithCommitLang(lang string) Option {
	return func(o *Options) {
		o.commitLang = lang
	}
}
