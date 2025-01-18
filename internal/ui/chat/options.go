package chat

import (
	"context"

	"github.com/charmbracelet/lipgloss"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/ui"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

type Options struct {
	ctx        context.Context
	runMode    ui.RunMode
	promptMode ui.PromptMode
	renderer   *lipgloss.Renderer
	wordWrap   int

	engine *ai.Engine

	content  string
	messages []llms.ChatMessage
}

type Option func(*Options)

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

func WithEngine(engine *ai.Engine) Option {
	return func(o *Options) {
		o.engine = engine
	}
}

func WithRunMode(runMode ui.RunMode) Option {
	return func(o *Options) {
		o.runMode = runMode
	}
}

func WithPromptMode(promptMode ui.PromptMode) Option {
	return func(o *Options) {
		o.promptMode = promptMode
	}
}

func WithContent(content string) Option {
	return func(o *Options) {
		o.content = content
	}
}

func WithMessages(messages []llms.ChatMessage) Option {
	return func(o *Options) {
		o.messages = messages
	}
}

func WithWordWrap(wordWrap int) Option {
	return func(o *Options) {
		o.wordWrap = wordWrap
	}
}

func WithRenderer(renderer *lipgloss.Renderer) Option {
	return func(o *Options) {
		o.renderer = renderer
	}
}

func NewOptions(opts ...Option) *Options {
	o := &Options{
		runMode:    ui.CliMode,
		promptMode: ui.ChatPromptMode,
		renderer:   console.StderrRenderer(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
