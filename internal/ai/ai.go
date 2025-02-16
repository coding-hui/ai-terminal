package ai

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/coding-hui/common/util/slices"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

const (
	noExec = "[noexec]"
)

type Engine struct {
	mode    EngineMode
	running bool
	channel chan StreamCompletionOutput

	convoStore convo.Store
	model      Model

	Config *options.Config
}

func New(ops ...Option) (*Engine, error) {
	return applyOptions(ops...)
}

func (e *Engine) SetMode(m EngineMode) {
	e.mode = m
}

func (e *Engine) GetMode() EngineMode {
	return e.mode
}

func (e *Engine) GetChannel() chan StreamCompletionOutput {
	return e.channel
}

func (e *Engine) GetConvoStore() convo.Store {
	return e.convoStore
}

func (e *Engine) Interrupt() {
	e.channel <- StreamCompletionOutput{
		Content:    "[Interrupt]",
		Last:       true,
		Interrupt:  true,
		Executable: false,
	}

	e.running = false
}

func (e *Engine) CreateCompletion(ctx context.Context, messages []llms.ChatMessage) (*CompletionOutput, error) {
	e.running = true

	if err := e.setupChatContext(ctx, &messages); err != nil {
		return nil, err
	}

	rsp, err := e.model.GenerateContent(ctx, slices.Map(messages, convert), e.callOptions()...)
	if err != nil {
		return nil, errbook.Wrap("Failed to create completion.", err)
	}

	content := rsp.Choices[0].Content
	content = html.UnescapeString(content)

	e.appendAssistantMessage(content)

	e.running = false

	return &CompletionOutput{
		Command:     "",
		Explanation: content,
		Executable:  false,
		Usage:       rsp.Usage,
	}, nil
}

func (e *Engine) CreateStreamCompletion(ctx context.Context, messages []llms.ChatMessage) (*StreamCompletionOutput, error) {
	e.running = true

	streamingFunc := func(ctx context.Context, chunk []byte) error {
		if !e.Config.Quiet {
			e.channel <- StreamCompletionOutput{
				Content: string(chunk),
				Last:    false,
			}
		}
		return nil
	}

	if err := e.setupChatContext(ctx, &messages); err != nil {
		return nil, err
	}

	for _, v := range messages {
		err := e.convoStore.AddMessage(ctx, e.Config.CacheWriteToID, v)
		if err != nil {
			errbook.HandleError(errbook.Wrap("Failed to add user chat input message to convo", err))
		}
	}

	messageParts := slices.Map(messages, convert)
	rsp, err := e.model.GenerateContent(ctx, messageParts, e.callOptions(streamingFunc)...)
	if err != nil {
		e.running = false
		return nil, errbook.Wrap("Failed to create stream completion.", err)
	}

	executable := false
	output := rsp.Choices[0].Content

	if e.mode == ExecEngineMode {
		if !strings.HasPrefix(output, noExec) && !strings.Contains(output, "\n") {
			executable = true
		}
	}

	output = html.UnescapeString(output)

	if !e.Config.Quiet {
		e.channel <- StreamCompletionOutput{
			Content:    "",
			Last:       true,
			Executable: executable,
			Usage:      rsp.Usage,
		}
	}
	e.running = false

	e.appendAssistantMessage(output)

	return &StreamCompletionOutput{
		Content:    output,
		Last:       true,
		Executable: executable,
		Usage:      rsp.Usage,
	}, nil
}

func (e *Engine) callOptions(streamingFunc ...func(ctx context.Context, chunk []byte) error) []llms.CallOption {
	var opts []llms.CallOption
	if e.Config.MaxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(e.Config.MaxTokens))
	}
	if len(streamingFunc) > 0 && streamingFunc[0] != nil {
		opts = append(opts, llms.WithStreamingFunc(streamingFunc[0]))
	}
	opts = append(opts, llms.WithModel(e.Config.CurrentModel.Name))
	opts = append(opts, llms.WithMaxLength(e.Config.CurrentModel.MaxChars))
	opts = append(opts, llms.WithTemperature(e.Config.Temperature))
	opts = append(opts, llms.WithTopP(e.Config.TopP))
	opts = append(opts, llms.WithTopK(e.Config.TopK))
	opts = append(opts, llms.WithMultiContent(false))

	return opts
}

func (e *Engine) setupChatContext(ctx context.Context, messages *[]llms.ChatMessage) error {
	store := e.convoStore
	if store == nil {
		return errbook.New("no chat convo store found")
	}

	if !e.Config.NoCache && e.Config.CacheReadFromID != "" {
		history, err := store.Messages(ctx, e.Config.CacheReadFromID)
		if err != nil {
			return errbook.Wrap(fmt.Sprintf(
				"There was a problem reading the cache. Use %s / %s to disable it.",
				console.StderrStyles().InlineCode.Render("--no-cache"),
				console.StderrStyles().InlineCode.Render("NO_CACHE"),
			), err)
		}
		*messages = append(*messages, history...)
	}

	return nil
}

func (e *Engine) appendAssistantMessage(content string) {
	if e.convoStore != nil && e.Config.CacheWriteToID != "" {
		if err := e.convoStore.AddAIMessage(context.Background(), e.Config.CacheWriteToID, content); err != nil {
			errbook.HandleError(errbook.Wrap("failed to add assistant chat output message to convo", err))
		}
	}
}

func convert(msg llms.ChatMessage) llms.MessageContent {
	return llms.MessageContent{
		Role:  msg.GetType(),
		Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
	}
}
