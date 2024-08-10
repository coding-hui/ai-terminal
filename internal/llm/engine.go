package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/coding-hui/common/util/slices"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/openai"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/session"
)

const (
	noExec = "[noexec]"
)

type Engine struct {
	mode    EngineMode
	config  *options.Config
	llm     *openai.Model
	channel chan EngineChatStreamOutput
	pipe    string
	running bool
	chatID  string

	execHistory, chatHistory session.History
}

func NewLLMEngine(mode EngineMode, config *options.Config) (*Engine, error) {
	var opts []openai.Option
	if config.Ai.Model != "" {
		opts = append(opts, openai.WithModel(config.Ai.Model))
	}
	if config.Ai.ApiBase != "" {
		opts = append(opts, openai.WithBaseURL(config.Ai.ApiBase))
	}
	if config.Ai.Token != "" {
		opts = append(opts, openai.WithToken(config.Ai.Token))
	}
	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	if klog.V(2).Enabled() {
		llm.CallbacksHandler = LogHandler{}
	}

	var execHistory, chatHistory session.History
	chatHistory, err = session.GetHistoryStore(*config, ChatEngineMode.String())
	if err != nil {
		return nil, err
	}
	execHistory, err = session.GetHistoryStore(*config, ExecEngineMode.String())
	if err != nil {
		return nil, err
	}

	return &Engine{
		mode:        mode,
		config:      config,
		llm:         llm,
		channel:     make(chan EngineChatStreamOutput),
		pipe:        "",
		running:     false,
		chatID:      config.ChatID,
		chatHistory: chatHistory,
		execHistory: execHistory,
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

func (e *Engine) GetChannel() chan EngineChatStreamOutput {
	return e.channel
}

func (e *Engine) Interrupt() {
	e.channel <- EngineChatStreamOutput{
		content:    "[Interrupt]",
		last:       true,
		interrupt:  true,
		executable: false,
	}

	e.running = false
}

func (e *Engine) Clear() {
	if e.mode == ExecEngineMode {
		err := e.execHistory.Clear(context.Background(), e.chatID)
		if err != nil {
			klog.Fatal("failed to clean exec history.", err)
		}
	} else {
		err := e.chatHistory.Clear(context.Background(), e.chatID)
		if err != nil {
			klog.Fatal("failed to clean chat history.", err)
		}
	}
}

func (e *Engine) Reset() {
	err := e.execHistory.Clear(context.Background(), e.chatID)
	if err != nil {
		klog.Fatal("failed to clean exec history.", err)
	}
	err = e.chatHistory.Clear(context.Background(), e.chatID)
	if err != nil {
		klog.Fatal("failed to clean chat history.", err)
	}
}

func (e *Engine) SummaryMessages(messages []llms.ChatMessage) (string, error) {
	ctx := context.Background()

	messageParts := slices.Map(messages, convert)
	messageParts = append(messageParts, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart(summaryMessagesPrompt)},
	})

	rsp, err := e.llm.GenerateContent(ctx, messageParts,
		llms.WithModel(e.config.Ai.Model),
		llms.WithTemperature(e.config.Ai.Temperature),
		llms.WithTopP(e.config.Ai.TopP),
		llms.WithMaxTokens(256),
	)
	if err != nil {
		return "", err
	}

	return rsp.Choices[0].Content, nil
}

func (e *Engine) ExecCompletion(input string) (*EngineExecOutput, error) {
	ctx := context.Background()

	e.running = true

	e.appendUserMessage(input)

	messages := e.prepareCompletionMessages()
	rsp, err := e.llm.GenerateContent(ctx, messages,
		llms.WithModel(e.config.Ai.Model),
		llms.WithMaxTokens(e.config.Ai.MaxTokens),
		llms.WithTemperature(e.config.Ai.Temperature),
		llms.WithTopP(e.config.Ai.TopP),
	)
	if err != nil {
		return nil, err
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

func (e *Engine) ChatStreamCompletion(input string) error {
	ctx := context.Background()

	e.running = true

	e.appendUserMessage(input)

	streamingFunc := func(ctx context.Context, chunk []byte) error {
		e.channel <- EngineChatStreamOutput{
			content: string(chunk),
			last:    false,
		}
		return nil
	}

	messages := e.prepareCompletionMessages()
	klog.V(2).InfoS("prepareCompletionMessages", "input", input, "messages", messages)
	rsp, err := e.llm.GenerateContent(ctx, messages,
		llms.WithModel(e.config.Ai.Model),
		llms.WithMaxTokens(e.config.Ai.MaxTokens),
		llms.WithTemperature(e.config.Ai.Temperature),
		llms.WithTopP(e.config.Ai.TopP),
		llms.WithStreamingFunc(streamingFunc),
	)
	if err != nil {
		e.running = false
		return err
	}

	executable := false
	output := rsp.Choices[0].Content

	if e.mode == ExecEngineMode {
		if !strings.HasPrefix(output, noExec) && !strings.Contains(output, "\n") {
			executable = true
		}
	}

	e.channel <- EngineChatStreamOutput{
		content:    "",
		last:       true,
		executable: executable,
	}
	e.running = false
	e.appendAssistantMessage(output)

	return nil
}

func (e *Engine) ChatStream(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) error {
	e.running = true

	streamingFunc := func(ctx context.Context, chunk []byte) error {
		e.channel <- EngineChatStreamOutput{
			content: string(chunk),
			last:    false,
		}
		return nil
	}

	ops := []llms.CallOption{
		llms.WithModel(e.config.Ai.Model),
		llms.WithMaxTokens(e.config.Ai.MaxTokens),
		llms.WithTemperature(e.config.Ai.Temperature),
		llms.WithTopP(e.config.Ai.TopP),
		llms.WithStreamingFunc(streamingFunc),
	}

	for _, o := range options {
		ops = append(ops, o)
	}

	rsp, err := e.llm.GenerateContent(ctx, messages, ops...)
	if err != nil {
		e.running = false
		return err
	}

	executable := false
	output := rsp.Choices[0].Content

	if e.mode == ExecEngineMode {
		if !strings.HasPrefix(output, noExec) && !strings.Contains(output, "\n") {
			executable = true
		}
	}

	e.channel <- EngineChatStreamOutput{
		content:    "",
		last:       true,
		executable: executable,
	}
	e.running = false

	return nil
}

func (e *Engine) appendUserMessage(content string) {
	if len(strings.TrimSpace(content)) == 0 {
		klog.V(2).Info("user input content is empty")
		return
	}
	if e.mode == ExecEngineMode {
		if err := e.execHistory.AddUserMessage(context.Background(), e.chatID, content); err != nil {
			klog.Fatal("failed to add user exec input message to history", err)
		}
	} else {
		if err := e.chatHistory.AddUserMessage(context.Background(), e.chatID, content); err != nil {
			klog.Fatal("failed to add user chat input message to history", err)
		}
	}
}

func (e *Engine) appendAssistantMessage(content string) {
	if e.mode == ExecEngineMode {
		if err := e.execHistory.AddAIMessage(context.Background(), e.chatID, content); err != nil {
			klog.Fatal("failed to add assistant exec output message to history", err)
		}
	} else {
		if err := e.chatHistory.AddAIMessage(context.Background(), e.chatID, content); err != nil {
			klog.Fatal("failed to add assistant chat output message to history", err)
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

	if e.mode == ExecEngineMode {
		history, err := e.execHistory.Messages(context.Background(), e.chatID)
		if err != nil {
			klog.Fatal("failed to get exec history messages", err)
		}
		messages = append(messages, slices.Map(history, convert)...)
	} else {
		history, err := e.chatHistory.Messages(context.Background(), e.chatID)
		if err != nil {
			klog.Fatal("failed to get chat history messages", err)
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
		return `You are a powerful terminal assistant. Your primary language is Chinese and you are good at answering users' questions.`
	}
	return `You are a powerful terminal assistant. Your primary language is Chinese and you are good at answering users' questions in markdown format.`
}

func convert(msg llms.ChatMessage) llms.MessageContent {
	return llms.MessageContent{
		Role:  msg.GetType(),
		Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
	}
}
