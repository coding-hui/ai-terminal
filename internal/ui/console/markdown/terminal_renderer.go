package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// TerminalRenderer implements Renderer for terminal output
type TerminalRenderer struct {
	config     *Config
	listStack  []*listContext
	currentBuf *bytes.Buffer
}

// NewTerminalRenderer creates a new terminal renderer instance
func NewTerminalRenderer(opts ...Option) *TerminalRenderer {
	cfg := NewConfig(opts...)
	return &TerminalRenderer{
		config:    cfg,
		listStack: make([]*listContext, 0),
	}
}

func (t *TerminalRenderer) RenderHeader(level int, text string) {
	t.config.HeaderStyle.Print(text + " ")
}

func (t *TerminalRenderer) RenderCodeBlock(lang, code string) {
	style := styles.Get(t.config.CodeTheme)
	formatter := formatters.Get("terminal256")
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	iterator, _ := lexer.Tokenise(nil, code)
	var buf bytes.Buffer
	formatter.Format(&buf, style, iterator)

	fmt.Printf("\n\x1b[48;5;%dm%s\x1b[0m\n", t.config.Background, buf.String())
}

func (t *TerminalRenderer) RenderListItem(level int, ordered bool, text string) {
	if level >= len(t.listStack) {
		t.listStack = append(t.listStack, &listContext{ordered, 0})
	}

	prefix := "• "
	if ordered {
		t.listStack[level].counter++
		prefix = fmt.Sprintf("%d. ", t.listStack[level].counter)
	}

	// Add indentation and prefix
	t.config.ListStyle.Printf("%s%s", strings.Repeat(" ", level), prefix)
	
	// Print the text and add newline
	fmt.Println(text)
}

func (t *TerminalRenderer) RenderParagraph(text string) {
	fmt.Println()
}

func (t *TerminalRenderer) RenderText(text string, emphasized bool) {
	if text != "" {
		if emphasized {
			t.config.TextEmphasis.Print(text + " ")
		} else {
			fmt.Print(text + " ")
		}
	}
}

func (t *TerminalRenderer) Finalize() {
	// Clean up any resources if needed
}
