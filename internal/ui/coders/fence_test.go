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
