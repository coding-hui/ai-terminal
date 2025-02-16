package markdown

// Renderer interface defines the contract for different rendering implementations
type Renderer interface {
	RenderHeader(level int, text string)
	RenderCodeBlock(lang, code string)
	RenderListItem(level int, ordered bool, text string)
	RenderParagraph(text string)
	RenderText(text string, emphasized bool)
	Finalize() // Finalize rendering
}

// listContext maintains state for list rendering
type listContext struct {
	isOrdered bool
	counter   int
}
