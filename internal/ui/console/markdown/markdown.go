package markdown

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func renderDocument(md string, renderer Renderer) {
	// Create Markdown parser
	markdown := goldmark.New()
	doc := markdown.Parser().Parse(text.NewReader([]byte(md)))

	// Walk AST and delegate rendering to the renderer
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch v := n.(type) {
		case *ast.Heading:
			if entering {
				renderer.RenderHeader(v.Level, getHeaderText(v, md))
			}

		case *ast.FencedCodeBlock:
			if entering {
				lang := string(v.Language([]byte(md)))
				renderer.RenderCodeBlock(lang, extractCodeContent(v, md))
			}

		case *ast.List:
			if entering {
				renderer.RenderListItem(0, v.IsOrdered(), "")
			} else {
				renderer.RenderParagraph("")
			}

		case *ast.ListItem:
			if entering {
				text := getTextContent(v, md)
				renderer.RenderListItem(1, v.Parent().(*ast.List).IsOrdered(), text)
			}

		case *ast.Text:
			if entering {
				text := strings.TrimSpace(string(v.Segment.Value([]byte(md))))
				if text != "" {
					emphasized := hasParentOfType(v, ast.KindEmphasis)
					renderer.RenderText(text, emphasized)
				}
			}

		case *ast.Paragraph:
			if !entering {
				renderer.RenderParagraph("")
			}
		}
		return ast.WalkContinue, nil
	})
}

// getTextContent extracts text content from a node
func getTextContent(n ast.Node, md string) string {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if text, ok := c.(*ast.Text); ok {
			buf.Write(text.Segment.Value([]byte(md)))
		}
	}
	return buf.String()
}

// 提取代码块内容
func extractCodeContent(v *ast.FencedCodeBlock, md string) string {
	var code strings.Builder
	lines := v.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		seg := &line
		code.Write(seg.Value([]byte(md)))
	}
	return code.String()
}

// 获取标题文本
func getHeaderText(h *ast.Heading, md string) string {
	var buf bytes.Buffer
	for n := h.FirstChild(); n != nil; n = n.NextSibling() {
		if text, ok := n.(*ast.Text); ok {
			buf.Write(text.Segment.Value([]byte(md)))
		}
	}
	return buf.String()
}

// 检查父节点类型
func hasParentOfType(n ast.Node, kind ast.NodeKind) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if p.Kind() == kind {
			return true
		}
	}
	return false
}
