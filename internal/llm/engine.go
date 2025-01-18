package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/caarlos0/go-shellwords"
	"github.com/coding-hui/common/util/slices"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/openai"

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
	config  *options.Config
	channel chan StreamCompletionOutput
	pipe    string
	running bool

	convoStore convo.Store

	Model *openai.Model
}

func NewLLMEngine(mode EngineMode, cfg *options.Config) (*Engine, error) {
	var api options.API
	mod, ok := cfg.Models[cfg.Model]
	if !ok {
		if cfg.API == "" {
			return nil, errbook.Wrap(
				fmt.Sprintf(
					"Model %s is not in the settings file.",
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
	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	chatHistory, err := convo.GetConversationStore(cfg)
	if err != nil {
		return nil, errbook.Wrap("Failed to get chat convo store.", err)
	}

	return &Engine{
		mode:       mode,
		config:     cfg,
		Model:      llm,
		channel:    make(chan StreamCompletionOutput),
		pipe:       "",
		running:    false,
		convoStore: chatHistory,
	}, nil
}

func (e *Engine) SetMode(m EngineMode) {
	e.mode = m
}

func (e *Engine) GetMode() EngineMode {
	return e.mode
}

func (e *Engine) SetPipe(pipe string) {
	e.pipe = pipe
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

	rsp, err := e.Model.GenerateContent(ctx, slices.Map(messages, convert), e.callOptions()...)
	if err != nil {
		return nil, errbook.Wrap("Failed to create completion.", err)
	}

	content := rsp.Choices[0].Content

	e.appendAssistantMessage(content)

	var output CompletionOutput
	err = json.Unmarshal([]byte(content), &output)
	if err != nil {
		output = CompletionOutput{
			Command:     "",
			Explanation: content,
			Executable:  false,
		}
	}

	return &output, nil
}

func (e *Engine) CreateStreamCompletion(ctx context.Context, messages []llms.ChatMessage) (*StreamCompletionOutput, error) {
	e.running = true

	streamingFunc := func(ctx context.Context, chunk []byte) error {
		if !e.config.Quiet {
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
		err := e.convoStore.AddMessage(ctx, e.config.CacheWriteToID, v)
		if err != nil {
			errbook.HandleError(errbook.Wrap("Failed to add user chat input message to convo", err))
		}
	}

	messageParts := slices.Map(messages, convert)
	rsp, err := e.Model.GenerateContent(ctx, messageParts, e.callOptions(streamingFunc)...)
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

	if !e.config.Quiet {
		e.channel <- StreamCompletionOutput{
			Content:    "",
			Last:       true,
			Executable: executable,
		}
	}
	e.running = false

	e.appendAssistantMessage(output)

	return &StreamCompletionOutput{
		Content:    output,
		Last:       true,
		Executable: executable,
	}, nil
}

func (e *Engine) callOptions(streamingFunc ...func(ctx context.Context, chunk []byte) error) []llms.CallOption {
	var opts []llms.CallOption
	if e.config.MaxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(e.config.MaxTokens))
	}
	if len(streamingFunc) > 0 && streamingFunc[0] != nil {
		opts = append(opts, llms.WithStreamingFunc(streamingFunc[0]))
	}
	opts = append(opts, llms.WithModel(e.config.Model))
	opts = append(opts, llms.WithMaxLength(e.config.MaxInputChars))
	opts = append(opts, llms.WithTemperature(e.config.Temperature))
	opts = append(opts, llms.WithTopP(e.config.TopP))
	opts = append(opts, llms.WithTopK(e.config.TopK))
	opts = append(opts, llms.WithMultiContent(false))

	return opts
}

func (e *Engine) setupChatContext(ctx context.Context, messages *[]llms.ChatMessage) error {
	store := e.convoStore
	if store == nil {
		return errbook.New("no chat convo store found")
	}

	if !e.config.NoCache && e.config.CacheReadFromID != "" {
		history, err := store.Messages(ctx, e.config.CacheReadFromID)
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
	if e.convoStore != nil {
		if err := e.convoStore.AddAIMessage(context.Background(), e.config.CacheWriteToID, content); err != nil {
			errbook.HandleError(errbook.Wrap("failed to add assistant chat output message to convo", err))
		}
	}
}

func ensureApiKey(api options.API) (string, error) {
	key := api.APIKey
	if key == "" && api.APIKeyEnv != "" && api.APIKeyCmd == "" {
		key = os.Getenv(api.APIKeyEnv)
	}
	if key == "" && api.APIKeyCmd != "" {
		args, err := shellwords.Parse(api.APIKeyCmd)
		if err != nil {
			return "", errbook.Wrap("Failed to parse api-key-cmd", err)
		}
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput() //nolint:gosec
		if err != nil {
			return "", errbook.Wrap("Cannot exec api-key-cmd", err)
		}
		key = strings.TrimSpace(string(out))
	}
	if key != "" {
		return key, nil
	}
	return "", errbook.Wrap(
		fmt.Sprintf(
			"%[1]s required; set the environment variable %[1]s or update %[2]s through %[3]s.",
			console.StderrStyles().InlineCode.Render(api.APIKeyEnv),
			console.StderrStyles().InlineCode.Render("config.yaml"),
			console.StderrStyles().InlineCode.Render("ai config"),
		),
		errbook.NewUserErrorf(
			"You can grab one at %s.",
			console.StderrStyles().Link.Render(api.BaseURL),
		),
	)
}

func convert(msg llms.ChatMessage) llms.MessageContent {
	return llms.MessageContent{
		Role:  msg.GetType(),
		Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
	}
}
