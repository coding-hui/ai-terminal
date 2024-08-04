package llm

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/callbacks"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

// LogHandler is a callback handler that prints to the standard output.
type LogHandler struct{}

var _ callbacks.Handler = LogHandler{}

func (l LogHandler) HandleLLMGenerateContentStart(_ context.Context, ms []llms.MessageContent) {
	info := strings.Builder{}
	info.WriteString("Entering LLM with messages:\n")
	for _, m := range ms {
		var buf strings.Builder
		for _, t := range m.Parts {
			if t, ok := t.(llms.TextContent); ok {
				buf.WriteString(t.Text)
			}
		}
		info.WriteString(fmt.Sprintf(" Role: %s\n", m.Role))
		info.WriteString(fmt.Sprintf(" Text: %s\n", strings.TrimSpace(buf.String())))
	}
	klog.V(2).Info(info.String())
}

func (l LogHandler) HandleLLMGenerateContentEnd(_ context.Context, res *llms.ContentResponse) {
	info := strings.Builder{}
	info.WriteString("Exiting LLM with response:\n")
	for _, c := range res.Choices {
		if c.Content != "" {
			info.WriteString(fmt.Sprintf(" Content: %s\n", c.Content))
		}
		if c.StopReason != "" {
			info.WriteString(fmt.Sprintf(" StopReason: %s\n", c.StopReason))
		}
		if len(c.GenerationInfo) > 0 {
			info.WriteString(" GenerationInfo:\n")
			for k, v := range c.GenerationInfo {
				info.WriteString(fmt.Sprintf("%20s: %v\n", k, v))
			}
		}
		if c.FuncCall != nil {
			info.WriteString(fmt.Sprintf(" FuncCall: %s %s\n", c.FuncCall.Name, c.FuncCall.Arguments))
		}
	}
	klog.V(2).Info(info.String())
}

func (l LogHandler) HandleStreamingFunc(_ context.Context, chunk []byte) {
	fmt.Println(string(chunk))
}

func (l LogHandler) HandleText(_ context.Context, text string) {
	fmt.Println(text)
}

func (l LogHandler) HandleLLMStart(_ context.Context, prompts []string) {
	fmt.Println("Entering LLM with prompt:", prompts)
}

func (l LogHandler) HandleLLMError(_ context.Context, err error) {
	fmt.Println("Exiting LLM with error:", err)
}

func (l LogHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	fmt.Println("Entering chain with inputs:", formatChainValues(inputs))
}

func (l LogHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	fmt.Println("Exiting chain with outputs:", formatChainValues(outputs))
}

func (l LogHandler) HandleChainError(_ context.Context, err error) {
	fmt.Println("Exiting chain with error:", err)
}

func (l LogHandler) HandleToolStart(_ context.Context, input string) {
	fmt.Println("Entering tool with input:", removeNewLines(input))
}

func (l LogHandler) HandleToolEnd(_ context.Context, output string) {
	fmt.Println("Exiting tool with output:", removeNewLines(output))
}

func (l LogHandler) HandleToolError(_ context.Context, err error) {
	fmt.Println("Exiting tool with error:", err)
}

func formatChainValues(values map[string]any) string {
	output := ""
	for key, value := range values {
		output += fmt.Sprintf("\"%s\" : \"%s\", ", removeNewLines(key), removeNewLines(value))
	}

	return output
}

func removeNewLines(s any) string {
	return strings.ReplaceAll(fmt.Sprint(s), "\n", " ")
}
