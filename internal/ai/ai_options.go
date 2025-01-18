package ai

import (
	"fmt"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/openai"
)

type EngineOption func(*Engine)

func WithMode(mode EngineMode) EngineOption {
	return func(e *Engine) {
		e.mode = mode
	}
}

func WithConfig(cfg *options.Config) EngineOption {
	return func(e *Engine) {
		e.Config = cfg
	}
}

func WithStore(store convo.Store) EngineOption {
	return func(a *Engine) {
		a.convoStore = store
	}
}

func applyEngineOptions(engineOpts ...EngineOption) (engine *Engine, err error) {
	engine = &Engine{
		channel: make(chan StreamCompletionOutput),
		running: false,
	}

	for _, option := range engineOpts {
		option(engine)
	}

	cfg := engine.Config
	if cfg == nil {
		return nil, errbook.New("Failed to initialize engine. Config is nil.")
	}

	if engine.convoStore == nil {
		engine.convoStore, err = convo.GetConversationStore(cfg)
		if err != nil {
			return nil, errbook.Wrap("Failed to get chat convo store.", err)
		}
	}

	var api options.API
	mod, ok := cfg.Models[cfg.Model]
	if !ok {
		if cfg.API == "" {
			return nil, errbook.Wrap(
				fmt.Sprintf(
					"model %s is not in the settings file.",
					console.StderrStyles().InlineCode.Render(cfg.Model),
				),
				errbook.NewUserErrorf(
					"Please specify an API endpoint with %s or configure the model in the settings: %s",
					console.StderrStyles().InlineCode.Render("--api"),
					console.StderrStyles().InlineCode.Render("ai -s"),
				),
			)
		}
		mod.Name = cfg.Model
		mod.API = cfg.API
		mod.MaxChars = cfg.MaxInputChars
	}
	if cfg.API != "" {
		mod.API = cfg.API
	}
	for _, a := range cfg.APIs {
		if mod.API == a.Name {
			api = a
			break
		}
	}
	if api.Name == "" {
		eps := make([]string, 0)
		for _, a := range cfg.APIs {
			eps = append(eps, console.StderrStyles().InlineCode.Render(a.Name))
		}
		return nil, errbook.Wrap(
			fmt.Sprintf(
				"The API endpoint %s is not configured.",
				console.StderrStyles().InlineCode.Render(cfg.API),
			),
			errbook.NewUserErrorf(
				"Your configured API endpoints are: %s",
				eps,
			),
		)
	}

	key, err := ensureApiKey(api)
	if err != nil {
		return nil, err
	}

	var opts []openai.Option
	opts = append(opts,
		openai.WithModel(mod.Name),
		openai.WithBaseURL(api.BaseURL),
		openai.WithToken(key),
	)
	engine.model, err = openai.New(opts...)
	if err != nil {
		return nil, err
	}

	return engine, nil
}
