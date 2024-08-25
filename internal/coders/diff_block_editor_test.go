package coders

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditBlockCoder(t *testing.T) {
	t.Run("splitRawBlocks", testSplitRawBlocks)
	t.Run("findOriginalUpdateBlocks", testFindOriginalUpdateBlocks)
	t.Run("findFilename", testFindFilename)
	t.Run("replaceMostSimilarChunk", testReplaceMostSimilarChunk)
}

func testSplitRawBlocks(t *testing.T) {
	content := `To add comments explaining the methods in **internal/coders/code_editor.go**, we will add them right
above each method declaration.

Here are the *SEARCH/REPLACE* blocks:

<code>go
internal/coders/code_editor.go
<<<<<<< SEARCH
type Coder interface {
	// Name Returns the name of the code editor.
	Name() string

	// Prompt Returns the prompt template used by the code editor.
	Prompt() prompts.ChatPromptTemplate

	// FormatMessages formats the messages with the values and returns the formatted messages.
	FormatMessages(values map[string]any) ([]llms.MessageContent, error)

	GetEdits(ctx context.Context) ([]PartialCodeBlock, error)

	ApplyEdits(ctx context.Context, edits []PartialCodeBlock) error

	// Execute Executes the code editor with the given input.
	Execute(ctx context.Context, messages []llms.MessageContent) error
}
=======
type Coder interface {
	// Name returns the name of the code editor.
	Name() string

	// Prompt returns the prompt template used by the code editor.
	Prompt() prompts.ChatPromptTemplate

	// FormatMessages formats the messages with the provided values and returns the formatted messages.
	FormatMessages(values map[string]any) ([]llms.MessageContent, error)

	// GetEdits retrieves the list of edits made to the code.
	GetEdits(ctx context.Context) ([]PartialCodeBlock, error)

	// ApplyEdits applies the given list of edits to the code.
	ApplyEdits(ctx context.Context, edits []PartialCodeBlock) error

	// Execute executes the code editor with the given input messages.
	Execute(ctx context.Context, messages []llms.MessageContent) error
}
>>>>>>> REPLACE
</code>
`

	splitRe := regexp.MustCompile("(?m)^((?:" + separators + ")[ ]*\n)")
	s := splitRe.Split(content, -1)
	assert.Equal(t, len(s), 4)
}

func testFindOriginalUpdateBlocks(t *testing.T) {

	text := `为了优化 supportCommands，我们需要为注册的方法添加 context.Context 参数。这将允许我们在这些方法中使用上下文进行取消操作或其他上下文相关的操作。

以下是需要进行的更改：

1. 修改 supportCommands 的定义，使其包含 context.Context 参数。
2. 更新所有注册的方法，使其接受 context.Context 参数。
3. 更新 run 方法，使其传递 context.Context 参数给调用的方法。

以下是具体的更改步骤：

<code>go
internal\coders\commands.go
<<<<<<< SEARCH
var supportCommands = map[string]func(...string) tea.Msg{}
=======
var supportCommands = map[string]func(context.Context, ...string) tea.Msg{}
>>>>>>> REPLACE
</code>

<code>go
internal\coders\commands.go
<<<<<<< SEARCH
func (c *command) registryCmds() {
	supportCommands["/add"] = c.addFiles
	supportCommands["/list"] = c.listFiles
	supportCommands["/remove"] = c.removeFiles
	supportCommands["/ask"] = c.askFiles
	supportCommands["/drop"] = c.dropFiles
	supportCommands["/coding"] = c.coding
}
=======
func (c *command) registryCmds() {
	supportCommands["/add"] = c.addFiles
	supportCommands["/list"] = c.listFiles
	supportCommands["/remove"] = c.removeFiles
	supportCommands["/ask"] = c.askFiles
	supportCommands["/drop"] = c.dropFiles
	supportCommands["/coding"] = c.coding
}

func (c *command) addFiles(ctx context.Context, files ...string) tea.Msg {
	return c.addFiles(files...)
}

func (c *command) listFiles(ctx context.Context, _ ...string) tea.Msg {
	return c.listFiles()
}

func (c *command) removeFiles(ctx context.Context, files ...string) tea.Msg {
	return c.removeFiles(files...)
}

func (c *command) askFiles(ctx context.Context, args ...string) tea.Msg {
	return c.askFiles(args...)
}

func (c *command) dropFiles(ctx context.Context, _ ...string) tea.Msg {
	return c.dropFiles()
}

func (c *command) coding(ctx context.Context, args ...string) tea.Msg {
	return c.coding(args...)
}
>>>>>>> REPLACE
</code>

<code>go
internal\coders\commands.go
<<<<<<< SEARCH
func (c *command) run(input string) tea.Cmd {
	return func() tea.Msg {
		cmd, args := "/ask", []string{input}
		if c.isCommand(input) {
			cmd, args = extractCmdArgs(input)
		}

		cmdFunc, ok := supportCommands[cmd]
		if !ok {
			return c.coder.Errorf("Invalid command %s", cmd)
		}

		// do run
		return cmdFunc(args...)
	}
}
=======
func (c *command) run(ctx context.Context, input string) tea.Cmd {
	return func() tea.Msg {
		cmd, args := "/ask", []string{input}
		if c.isCommand(input) {
			cmd, args = extractCmdArgs(input)
		}

		cmdFunc, ok := supportCommands[cmd]
		if !ok {
			return c.coder.Errorf("Invalid command %s", cmd)
		}

		// do run
		return cmdFunc(ctx, args...)
	}
}
>>>>>>> REPLACE
</code>

这些更改将为 supportCommands 中的所有方法添加 context.Context 参数，并确保在调用这些方法时传递上下文。
`
	editor := NewEditBlockCoder(NewAutoCoder(), []string{"<code>", "</code>"})
	editor.partialResponseContent = text

	go func() {
		<-editor.coder.checkpointChan
	}()

	edits, err := editor.GetEdits(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, edits)
	assert.Equal(t, 3, len(edits))
}

func testFindFilename(t *testing.T) {
	type args struct {
		line  string
		fence []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				`To add comments explaining the methods in **internal/coders/code_editor.go**, we will add them right above each method declaration.

Here are the *SEARCH/REPLACE* blocks:

<code>go
internal/coders/code_editor.go`,
				[]string{"<code>", "</code>"},
			},
			want: "internal/coders/code_editor.go",
		},
		{
			name: "t2",
			args: args{
				`To add comments explaining the methods in **internal/coders/code_editor.go**, we will add them right above each method declaration.

Here are the *SEARCH/REPLACE* blocks:

<code>go
internal/coders/code_editor.go
`,
				[]string{"<code>", "</code>"},
			},
			want: "internal/coders/code_editor.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, findFilename(tt.args.line, tt.args.fence), "findFilename(%v, %v)", tt.args.line, tt.args.fence)
		})
	}
}

func testReplaceMostSimilarChunk(t *testing.T) {
	type args struct {
		whole   string
		part    string
		replace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"replace_most_similar_chunk",
			args{
				"This is a sample text.\nAnother line of text.\nYet another line.\n",
				"This is a sample text\n",
				"This is a replaced text.\n",
			},
			"This is a replaced text.\nAnother line of text.\nYet another line.\n",
		},
		{
			"replace_most_similar_chunk_not_perfect_match",
			args{
				"This is a sample text.\nAnother line of text.\nYet another line.\n",
				"This was a sample text.\nAnother line of txt\n",
				"This is a replaced text.\nModified line of text.\n",
			},
			"This is a replaced text.\nModified line of text.\nYet another line.\n",
		},
		{
			"replace_part_with_missing_varied_leading_whitespace",
			args{
				`
    line1
    line2
        line3
    line4
`,
				"line2\n    line3\n",
				"new_line2\n    new_line3\n",
			},
			`
    line1
    new_line2
        new_line3
    line4
`,
		},
		{
			"replace_part_with_missing_leading_whitespace",
			args{
				"    line1\n    line2\n    line3\n",
				"line1\nline2\n",
				"new_line1\nnew_line2\n",
			},
			"    new_line1\n    new_line2\n    line3\n",
		},
		{
			"replace_part_with_just_some_missing_leading_whitespace",
			args{
				"    line1\n    line2\n    line3\n",
				" line1\n line2\n",
				" new_line1\n     new_line2\n",
			},
			"    new_line1\n    new_line2\n    line3\n",
		},
		{
			"replace_part_with_missing_leading_whitespace_including_blank_line",
			args{
				"    line1\n    line2\n    line3\n",
				"\n  line1\n  line2\n",
				"  new_line1\n  new_line2\n",
			},
			"    new_line1\n    new_line2\n    line3\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, replaceMostSimilarChunk(tt.args.whole, tt.args.part, tt.args.replace), "replaceMostSimilarChunk(%v, %v, %v)", tt.args.whole, tt.args.part, tt.args.replace)
		})
	}
}
