package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms/openai"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
)

const noExec = "[noexec]"

type Engine struct {
	mode         EngineMode
	config       *options.Config
	llm          *openai.Model
	execMessages []llms.MessageContent
	chatMessages []llms.MessageContent
	channel      chan EngineChatStreamOutput
	pipe         string
	running      bool
}

func NewDefaultEngine(mode EngineMode, config *options.Config) (*Engine, error) {
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

	return &Engine{
		mode:         mode,
		config:       config,
		llm:          llm,
		execMessages: make([]llms.MessageContent, 0),
		chatMessages: make([]llms.MessageContent, 0),
		channel:      make(chan EngineChatStreamOutput),
		pipe:         "",
		running:      false,
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
		e.execMessages = []llms.MessageContent{}
	} else {
		e.chatMessages = []llms.MessageContent{}
	}
}

func (e *Engine) Reset() {
	e.execMessages = []llms.MessageContent{}
	e.chatMessages = []llms.MessageContent{}
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

func (e *Engine) appendUserMessage(content string) {
	if len(strings.TrimSpace(content)) == 0 {
		klog.V(2).Info("user input content is empty")
		return
	}
	if e.mode == ExecEngineMode {
		e.execMessages = append(e.execMessages, llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(content)},
		})
	} else {
		e.chatMessages = append(e.chatMessages, llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(content)},
		})
	}
}

func (e *Engine) appendAssistantMessage(content string) {
	if e.mode == ExecEngineMode {
		e.execMessages = append(e.execMessages, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(content)},
		})
	} else {
		e.chatMessages = append(e.chatMessages, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(content)},
		})
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
		messages = append(messages, e.execMessages...)
	} else {
		messages = append(messages, e.chatMessages...)
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
