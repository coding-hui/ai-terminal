package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/caarlos0/go-shellwords"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coding-hui/common/util/slices"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/openai"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

const (
	noExec        = "[noexec]"
	defaultChatID = "temp_session"
)

type Engine struct {
	mode    EngineMode
	config  *options.Config
	channel chan StreamCompletionOutput
	pipe    string
	running bool
	chatID  string

	chatHistory convo.Store

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

	chatHistory, err := convo.GetConversationStore(*cfg)
	if err != nil {
		return nil, errbook.Wrap("Failed to get chat history store.", err)
	}

	charID := defaultChatID
	if cfg.ChatID != "" {
		charID = cfg.ChatID
	}

	return &Engine{
		mode:        mode,
		config:      cfg,
		Model:       llm,
		channel:     make(chan StreamCompletionOutput),
		pipe:        "",
		running:     false,
		chatID:      charID,
		chatHistory: chatHistory,
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

func (e *Engine) Interrupt() {
	e.channel <- StreamCompletionOutput{
		content:    "[Interrupt]",
		last:       true,
		interrupt:  true,
		executable: false,
	}

	e.running = false
}

func (e *Engine) Clear() {
	err := e.chatHistory.ClearConversations(context.Background())
	if err != nil {
		errbook.HandleError(errbook.Wrap("failed to clear chat history.", err))
	}
}

func (e *Engine) Reset() {
	err := e.chatHistory.ClearConversations(context.Background())
	if err != nil {
		errbook.HandleError(errbook.Wrap("failed to reset chat history.", err))
	}
}

func (e *Engine) CreateCompletion(messages []llms.ChatMessage) (*EngineExecOutput, error) {
	ctx := context.Background()

	e.running = true

	rsp, err := e.Model.GenerateContent(ctx, slices.Map(messages, convert), e.callOptions()...)
	if err != nil {
		return nil, errbook.Wrap("Failed to create completion.", err)
	}

	content := rsp.Choices[0].Content
	e.appendAssistantMessage(content)

	var output EngineExecOutput
	err = json.Unmarshal([]byte(content), &output)
	if err != nil {
		output = EngineExecOutput{
			Command:     "",
			Explanation: content,
			Executable:  false,
		}
	}

	return &output, nil
}

func (e *Engine) CreateStreamCompletion(messages []llms.ChatMessage) tea.Msg {
	ctx := context.Background()

	e.running = true

	streamingFunc := func(ctx context.Context, chunk []byte) error {
		if !e.config.Quiet {
			e.channel <- StreamCompletionOutput{
				content: string(chunk),
				last:    false,
			}
		}
		return nil
	}

	messageParts := slices.Map(messages, convert)
	rsp, err := e.Model.GenerateContent(ctx, messageParts, e.callOptions(streamingFunc)...)
	if err != nil {
		e.running = false
		return errbook.Wrap("Failed to create stream completion.", err)
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
			content:    "",
			last:       true,
			executable: executable,
		}
	}
	e.running = false
	e.appendAssistantMessage(output)

	return &StreamCompletionOutput{
		content:    output,
		last:       true,
		executable: executable,
	}
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

func (e *Engine) appendUserMessage(content string) {
	if len(strings.TrimSpace(content)) == 0 {
		errbook.HandleError(errbook.New("empty input is not allowed."))
		return
	}
	if e.chatHistory != nil {
		if err := e.chatHistory.AddUserMessage(context.Background(), e.chatID, content); err != nil {
			errbook.HandleError(errbook.Wrap("failed to add user chat input message to history", err))
		}
	}
}

func (e *Engine) appendAssistantMessage(content string) {
	if e.chatHistory != nil {
		if err := e.chatHistory.AddAIMessage(context.Background(), e.chatID, content); err != nil {
			errbook.HandleError(errbook.Wrap("failed to add assistant chat output message to history", err))
		}
	}
}

func (e *Engine) prepareCompletionMessages() []llms.MessageContent {
	var messages []llms.MessageContent

	if e.config.Ai.SystemPrompt == "" {
		messages = append(messages, llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(e.prepareSystemPrompt())},
		})
	}

	if e.pipe != "" {
		messages = append(
			messages,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart(e.preparePipePrompt())},
			},
		)
	}

	if e.chatHistory != nil {
		history, err := e.chatHistory.Messages(context.Background(), e.chatID)
		if err != nil {
			errbook.HandleError(errbook.Wrap("failed to get chat history", err))
		}
		messages = append(messages, slices.Map(history, convert)...)
	}

	return messages
}

func (e *Engine) preparePipePrompt() string {
	return fmt.Sprintf("I will work on the following input: %s", e.pipe)
}

func (e *Engine) prepareSystemPrompt() string {
	var bodyPart string
	if e.mode == ExecEngineMode {
		bodyPart = e.prepareSystemPromptExecPart()
	} else {
		bodyPart = e.prepareSystemPromptChatPart()
	}

	return bodyPart
}

func (e *Engine) prepareSystemPromptExecPart() string {
	return "Your are a powerful terminal assistant generating a JSON containing a command line for my input.\n" +
		"The language you are using is Chinese.\n" +
		"You will always reply using the following json structure: {\"cmd\":\"the command\", \"exp\": \"some explanation\", \"exec\": true}.\n" +
		"Your answer will always only contain the json structure, never add any advice or supplementary detail or information, even if I asked the same question before.\n" +
		"The field cmd will contain a single line command (don't use new lines, use separators like && and ; instead).\n" +
		"The field exp will contain an short explanation of the command if you managed to generate an executable command, otherwise it will contain the reason of your failure.\n" +
		"The field exec will contain true if you managed to generate an executable command, false otherwise." +
		"\n" +
		"Examples:\n" +
		"Me: list all files in my home dir\n" +
		"You: {\"cmd\":\"ls ~\", \"exp\": \"list all files in your home dir\", \"exec\\: true}\n" +
		"Me: list all pods of all namespaces\n" +
		"You: {\"cmd\":\"kubectl get pods --all-namespaces\", \"exp\": \"list pods form all k8s namespaces\", \"exec\": true}\n" +
		"Me: how are you ?\n" +
		"You: {\"cmd\":\"\", \"exp\": \"I'm good thanks but I cannot generate a command for this. Use the chat mode to discuss.\", \"exec\": false}"
}

func (e *Engine) prepareSystemPromptChatPart() string {
	if e.config.Ai.OutputFormat == options.RawOutputFormat {
		return `You are a powerful terminal assistant. Your primary language is English and you are good at answering users' questions.`
	}
	return `You are a powerful terminal assistant. Your primary language is English and you are good at answering users' questions in markdown format.`
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
