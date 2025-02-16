package coders

import (
	"fmt"
	"strings"
	"testing"
)

func TestChooseExistingFence(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedOpen  string
		expectedClose string
	}{
		{
			name:          "backtick fence with SEARCH/REPLACE",
			input:         "```go\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n```",
			expectedOpen:  "```",
			expectedClose: "```",
		},
		{
			name:          "code tag fence with SEARCH/REPLACE",
			input:         "<code>\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n</code>",
			expectedOpen:  "<code>",
			expectedClose: "</code>",
		},
		{
			name:          "pre tag fence with SEARCH/REPLACE",
			input:         "<pre>\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n</pre>",
			expectedOpen:  "<pre>",
			expectedClose: "</pre>",
		},
		{
			name:          "no matching fence",
			input:         "just some text\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n",
			expectedOpen:  "```",
			expectedClose: "```",
		},
		{
			name:          "code fence",
			input:         "I'll help implement these changes using *SEARCH/REPLACE* blocks:\n\n1. First, let's add the new option to chat.go:\n\n<code>go\ninternal/ui/chat/chat.go\n<<<<<<< SEARCH\ntype Options struct {\n\tengine   *ai.Engine\n\trenderer *lipgloss.Renderer\n\tmessages []llms.ChatMessage\n\tcontent  string\n}\n=======\ntype Options struct {\n\tengine         *ai.Engine\n\trenderer       *lipgloss.Renderer\n\tmessages       []llms.ChatMessage\n\tcontent        string\n\tCopyToClipboard bool\n}\n>>>>>>> REPLACE\n</code>\n\n<code>go\ninternal/ui/chat/chat.go\n<<<<<<< SEARCH\nfunc NewOptions(opts ...Option) *Options {\n\to := &Options{}\n\tfor _, opt := range opts {\n\t\topt(o)\n\t}\n\treturn o\n}\n=======\nfunc NewOptions(opts ...Option) *Options {\n\to := &Options{\n\t\tCopyToClipboard: false, // Default to false\n\t}\n\tfor _, opt := range opts {\n\t\topt(o)\n\t}\n\treturn o\n}\n\nfunc WithCopyToClipboard(copy bool) Option {\n\treturn func(o *Options) {\n\t\to.CopyToClipboard = copy\n\t}\n}\n>>>>>>> REPLACE\n</code>\n\n2. Now let's modify the Run() method:\n\n<code>go\ninternal/ui/chat/chat.go\n<<<<<<< SEARCH\n\tif term.IsOutputTTY() {\n\t\tif c.config.Raw && c.Output != \"\" {\n\t\t\tfmt.Print(c.Output)\n\t\t} else {\n\t\t\tswitch {\n\t\t\tcase c.GlamOutput != \"\":\n\t\t\t\tfmt.Print(c.GlamOutput)\n\t\t\tcase c.Output != \"\":\n\t\t\t\tfmt.Print(c.Output)\n\t\t\t}\n\t\t}\n\t}\n\n\tif c.config.Show != \"\" || c.config.ShowLast {\n=======\n\tif term.IsOutputTTY() {\n\t\tif c.config.Raw && c.Output != \"\" {\n\t\t\tfmt.Print(c.Output)\n\t\t} else {\n\t\t\tswitch {\n\t\t\tcase c.GlamOutput != \"\":\n\t\t\t\tfmt.Print(c.GlamOutput)\n\t\t\tcase c.Output != \"\":\n\t\t\t\tfmt.Print(c.Output)\n\t\t\t}\n\t\t}\n\t}\n\n\t// Add clipboard support\n\tif c.opts.CopyToClipboard {\n\t\tcontent := c.Output\n\t\tif c.GlamOutput != \"\" {\n\t\t\tcontent = c.GlamOutput\n\t\t}\n\t\tif content != \"\" {\n\t\t\tif err := clipboard.WriteAll(content); err != nil {\n\t\t\t\tconsole.RenderError(err, \"Failed to copy to clipboard\")\n\t\t\t} else {\n\t\t\t\tconsole.Render(\"Copied result to clipboard!\")\n\t\t\t}\n\t\t}\n\t}\n\n\tif c.config.Show != \"\" || c.config.ShowLast {\n>>>>>>> REPLACE\n</code>\n\n3. Update the command handlers:\n\n<code>go\ninternal/ui/coders/commands.go\n<<<<<<< SEARCH\n\tchatModel := chat.NewChat(c.coder.cfg,\n\t\tchat.WithContext(ctx),\n\t\tchat.WithMessages(messages),\n\t\tchat.WithEngine(c.coder.engine),\n\t)\n=======\n\tchatModel := chat.NewChat(c.coder.cfg,\n\t\tchat.WithContext(ctx),\n\t\tchat.WithMessages(messages),\n\t\tchat.WithEngine(c.coder.engine),\n\t\tchat.WithCopyToClipboard(true),\n\t)\n>>>>>>> REPLACE\n</code>\n\n<code>go\ninternal/ui/coders/commands.go\n<<<<<<< SEARCH\n\tchatModel := chat.NewChat(c.coder.cfg,\n\t\tchat.WithContext(ctx),\n\t\tchat.WithMessages(messages),\n\t\tchat.WithEngine(c.coder.engine),\n\t)\n=======\n\tchatModel := chat.NewChat(c.coder.cfg,\n\t\tchat.WithContext(ctx),\n\t\tchat.WithMessages(messages),\n\t\tchat.WithEngine(c.coder.engine),\n\t\tchat.WithCopyToClipboard(true),\n\t)\n>>>>>>> REPLACE\n</code>\n\n4. Add clipboard import to chat.go:\n\n<code>go\ninternal/ui/chat/chat.go\n<<<<<<< SEARCH\n\t\"unicode\"\n\n\t\"github.com/charmbracelet/bubbles/viewport\"\n=======\n\t\"unicode\"\n\n\t\"github.com/atotto/clipboard\"\n\t\"github.com/charmbracelet/bubbles/viewport\"\n>>>>>>> REPLACE\n</code>\n\nThese changes will:\n1. Add clipboard support to the chat package\n2. Enable clipboard copying by default for ask, design and coding commands\n3. Show success/error messages for clipboard operations\n4. Copy either raw or formatted output depending on what's available",
			expectedOpen:  "<code>",
			expectedClose: "</code>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, c := chooseExistingFence(tt.input)
			if o != tt.expectedOpen || c != tt.expectedClose {
				t.Errorf("chooseExistingFence() = (%q, %q), want (%q, %q)",
					o, c, tt.expectedOpen, tt.expectedClose)
			}
		})
	}
}

func BenchmarkChooseExistingFenceLargeStrings(b *testing.B) {
	largeTestCases := []struct {
		name  string
		input string
	}{
		{
			name:  "10k_lines_with_backticks",
			input: generateLargeString("```go\n", "```\n", 10000),
		},
		{
			name:  "100k_lines_no_fence",
			input: strings.Repeat("some random text\n<<<<<<< SEARCH\n=======\n>>>>>>> REPLACE\n", 25000),
		},
		{
			name:  "mixed_fences_1mb",
			input: generateMixedFences(1024 * 1024), // 1MB string
		},
	}

	for _, tc := range largeTestCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				chooseExistingFence(tc.input)
			}
		})
	}
}

// generateLargeString 帮助函数用于生成包含指定围栏的大字符串
func generateLargeString(open, close string, lines int) string {
	var sb strings.Builder
	sb.WriteString(open + "\n")
	for i := 0; i < lines; i++ {
		sb.WriteString(fmt.Sprintf("<<<<<<< SEARCH\nline %d\n=======\nline %d\n>>>>>>> REPLACE\n", i, i))
	}
	sb.WriteString(close + "\n")
	return sb.String()
}

// generateMixedFences 生成包含多种围栏类型的混合大字符串
func generateMixedFences(targetLen int) string {
	var sb strings.Builder
	fences := []string{"```", "<code>", "<pre>"}
	i := 0
	for sb.Len() < targetLen {
		fence := fences[i%len(fences)]
		sb.WriteString(fmt.Sprintf("%s\n<<<<<<< SEARCH\ncontent\n=======\ncontent\n>>>>>>> REPLACE\n%s\n",
			fence, strings.Replace(fence, "<", "</", 1)))
		i++
		// 添加随机文本
		sb.WriteString(strings.Repeat("some random text\n", 100))
	}
	return sb.String()[:targetLen]
}
