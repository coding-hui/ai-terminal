package ai

import (
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/openai"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/volcengine"
)

type Option func(*Engine)

func WithMode(mode EngineMode) Option {
	return func(e *Engine) {
		e.mode = mode
	}
}

func WithConfig(cfg *options.Config) Option {
	return func(e *Engine) {
		e.Config = cfg
	}
}

func WithStore(store convo.Store) Option {
	return func(a *Engine) {
		a.convoStore = store
	}
}

func applyOptions(engineOpts ...Option) (engine *Engine, err error) {
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

	cfg.CurrentModel, err = cfg.GetModel(cfg.Model)
	if err != nil {
		return nil, err
	}

	cfg.CurrentAPI, err = cfg.GetAPI(cfg.API)
	if err != nil {
		return nil, err
	}

	mod, api := cfg.CurrentModel, cfg.CurrentAPI
	switch api.Name {
	case ModelTypeARK:
		engine.model, err = volcengine.NewClientWithApiKey(
			cfg.CurrentAPI.APIKey,
			arkruntime.WithBaseUrl(api.BaseURL),
			arkruntime.WithRegion(api.Region),
			arkruntime.WithTimeout(api.Timeout),
			arkruntime.WithRetryTimes(api.RetryTimes),
		)
		if err != nil {
			return nil, err
		}
	default:
		engine.model, err = openai.New(
			openai.WithModel(mod.Name),
			openai.WithBaseURL(api.BaseURL),
			openai.WithToken(api.APIKey),
		)
		if err != nil {
			return nil, err
		}
	}

	return engine, nil
}
