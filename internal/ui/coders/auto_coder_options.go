package coders

import (
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/options"

	"github.com/coding-hui/common/version"
)

type AutoCoderOption func(*AutoCoder)

func WithConfig(cfg *options.Config) AutoCoderOption {
	return func(a *AutoCoder) {
		a.cfg = cfg
	}
}

func WithEngine(engine *llm.Engine) AutoCoderOption {
	return func(a *AutoCoder) {
		a.engine = engine
	}
}

func WithRepo(repo *git.Command) AutoCoderOption {
	return func(a *AutoCoder) {
		a.repo = repo
	}
}

func WithCodeBasePath(path string) AutoCoderOption {
	return func(a *AutoCoder) {
		a.codeBasePath = path
	}
}

func WithStore(store convo.Store) AutoCoderOption {
	return func(a *AutoCoder) {
		a.store = store
	}
}

func WithLoadedContexts(contexts []*convo.LoadContext) AutoCoderOption {
	return func(a *AutoCoder) {
		a.loadedContexts = contexts
	}
}

func WithPrompt(prompt string) AutoCoderOption {
	return func(a *AutoCoder) {
		a.prompt = prompt
	}
}

func applyAutoCoderOptions(options ...AutoCoderOption) *AutoCoder {
	ac := &AutoCoder{
		versionInfo:    version.Get(),
		loadedContexts: []*convo.LoadContext{},
	}

	for _, option := range options {
		option(ac)
	}

	if ac.loadedContexts == nil {
		ac.loadedContexts = []*convo.LoadContext{}
	}

	return ac
}
