package coders

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditBlockCoder(t *testing.T) {
	t.Run("splitRawBlocks", testSplitRawBlocks)
	t.Run("findFilename", testFindFilename)
	t.Run("replaceMostSimilarChunk", testReplaceMostSimilarChunk)
	t.Run("findOriginalUpdateBlocks", testFindOriginalUpdateBlocks)
	t.Run("parseEditBlock", testParseEditBlock)
	t.Run("findAllCodeBlocks", testFindAllCodeBlocks)
	t.Run("stripQuotedWrapping", testStripQuotedWrapping)
	t.Run("doReplace", testDoReplace)
	t.Run("perfectOrWhitespace", testPerfectOrWhitespace)
	t.Run("findSimilarLines", testFindSimilarLines)
	t.Run("ld", testLd)
}

func testSplitRawBlocks(t *testing.T) {
	content := `To add comments explaining the methods in **internal/coders/code_editor.go**, we will add them right
above each method declaration.

Here are the *SEARCH/REPLACE* blocks:

` + "```go" + `
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
` + "```" + `
`

	splitRe := regexp.MustCompile("(?m)^((?:" + separators + ")[ ]*\n)")
	s := splitRe.Split(content, -1)
	assert.Equal(t, 4, len(s))
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

` + "```go" + `
internal/coders/code_editor.go`,
				[]string{"```go", "```"},
			},
			want: "internal/coders/code_editor.go",
		},
		{
			name: "t2",
			args: args{
				`To add comments explaining the methods in **internal/coders/code_editor.go**, we will add them right above each method declaration.

Here are the *SEARCH/REPLACE* blocks:

` + "```go" + `
internal/coders/code_editor.go
`,
				[]string{"```go", "```"},
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
		{
			"replace_part_with_whole_file",
			args{
				`package console

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("9"))

func Error(err error, msgs ...string) {
	if err == nil {
		return
	}

	errMsg := err.Error()
	if errMsg == "" {
		return
	}

	if len(msgs) > 0 {
		ErrorMsg(msgs...)
	}
	ErrorMsg(err.Error())
}

func ErrorMsg(msgs ...string) {
	for _, msg := range msgs {
		fmt.Println(style.Render(msg))
	}
}

func FatalErr(err error, msgs ...string) {
	Error(err, msgs...)
	os.Exit(1)
}
`,
				"func Error(err error, msgs ...string) {",
				`// Error handles and displays an error message along with optional additional messages.
func Error(err error, msgs ...string) {`,
			},
			`package console

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("9"))

// Error handles and displays an error message along with optional additional messages.
func Error(err error, msgs ...string) {
	if err == nil {
		return
	}

	errMsg := err.Error()
	if errMsg == "" {
		return
	}

	if len(msgs) > 0 {
		ErrorMsg(msgs...)
	}
	ErrorMsg(err.Error())
}

func ErrorMsg(msgs ...string) {
	for _, msg := range msgs {
		fmt.Println(style.Render(msg))
	}
}

func FatalErr(err error, msgs ...string) {
	Error(err, msgs...)
	os.Exit(1)
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, replaceMostSimilarChunk(tt.args.whole, tt.args.part, tt.args.replace), "replaceMostSimilarChunk(%v, %v, %v)", tt.args.whole, tt.args.part, tt.args.replace)
		})
	}
}

func testFindOriginalUpdateBlocks(t *testing.T) {
	content := `
` + "```go" + `
file1.go
<<<<<<< SEARCH
old content 1
=======
new content 1
>>>>>>> REPLACE
` + "```" + `

` + "```go" + `
file2.go
<<<<<<< SEARCH
old content 2
=======
new content 2
>>>>>>> REPLACE
` + "```" + `
`
	fence := []string{"```go", "```"}

	blocks, err := findOriginalUpdateBlocks(content, fence)
	assert.NoError(t, err)
	assert.Len(t, blocks, 2)
	assert.Equal(t, "file1.go", blocks[0].Path)
	assert.Equal(t, "old content 1", blocks[0].OriginalText)
	assert.Equal(t, "new content 1", blocks[0].UpdatedText)
	assert.Equal(t, "file2.go", blocks[1].Path)
	assert.Equal(t, "old content 2", blocks[1].OriginalText)
	assert.Equal(t, "new content 2", blocks[1].UpdatedText)
}

func testParseEditBlock(t *testing.T) {
	edit := pathAndCode{
		path: "test.go",
		code: `<<<<<<< SEARCH
old content
=======
new content
>>>>>>> REPLACE`,
	}

	block, err := parseEditBlock(edit)
	assert.NoError(t, err)
	assert.Equal(t, "test.go", block.Path)
	assert.Equal(t, "old content", block.OriginalText)
	assert.Equal(t, "new content", block.UpdatedText)
}

func testFindAllCodeBlocks(t *testing.T) {
	content := `
` + "```go" + `
file1.go
<<<<<<< SEARCH
old content 1
=======
new content 1
>>>>>>> REPLACE
` + "```" + `

` + "```go" + `
file2.go
<<<<<<< SEARCH
old content 2
=======
new content 2
>>>>>>> REPLACE
` + "```" + `
`
	fence := []string{"```go", "```"}

	blocks := findAllCodeBlocks(content, fence)
	assert.Len(t, blocks, 2)
	assert.Equal(t, "file1.go", blocks[0].path)
	assert.Contains(t, blocks[0].code, "old content 1")
	assert.Contains(t, blocks[0].code, "new content 1")
	assert.Equal(t, "file2.go", blocks[1].path)
	assert.Contains(t, blocks[1].code, "old content 2")
	assert.Contains(t, blocks[1].code, "new content 2")
}

func testStripQuotedWrapping(t *testing.T) {
	res := "```go\nfunc test() {}\n```"
	fname := "test.go"
	fence := []string{"```go", "```"}

	stripped := stripQuotedWrapping(res, fname, fence)
	assert.Equal(t, "func test() {}\n", stripped)
}

func testDoReplace(t *testing.T) {
	fileName := "test.go"
	content := "func oldFunction() {}\n"
	beforeText := "func oldFunction() {}"
	afterText := "func newFunction() {}"
	fence := []string{"```go", "```"}

	result := doReplace(fileName, content, beforeText, afterText, fence)
	assert.Equal(t, "func newFunction() {}\n", result)
}

func testPerfectOrWhitespace(t *testing.T) {
	wholeLines := []string{"    line1", "    line2", "    line3"}
	partLines := []string{"line1", "line2"}
	replaceLines := []string{"newline1", "newline2"}

	result := perfectOrWhitespace(wholeLines, partLines, replaceLines)
	assert.Equal(t, "    newline1    newline2    line3", result)
}

func testFindSimilarLines(t *testing.T) {
	rawSearch := "This is a test line\nAnother test line"
	rawContent := "This is a test line\nAnother test line\nYet another line"

	result := findSimilarLines(rawSearch, rawContent)
	assert.Equal(t, "This is a test line\nAnother test line\n", result)
}

func testLd(t *testing.T) {
	s1 := "kitten"
	s2 := "sitting"

	distance := ld(s1, s2, false)
	assert.Equal(t, 3, distance)

	distance = ld(s1, s2, true)
	assert.Equal(t, 3, distance)
}
