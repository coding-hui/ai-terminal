package console

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	AppName,
	CliArgs,
	Comment,
	CyclingChars,
	ErrorHeader,
	ErrorDetails,
	ErrPadding,
	Flag,
	FlagComma,
	FlagDesc,
	InlineCode,
	Link,
	Pipe,
	Quote,
	ConversationList,
	SHA1,
	Timeago,
	CommitStep,
	CommitSuccess,
	DiffHeader,
	DiffFileHeader,
	DiffHunkHeader,
	DiffAdded,
	DiffRemoved,
	DiffContext lipgloss.Style
}

func MakeStyles(r *lipgloss.Renderer) (s Styles) {
	const horizontalEdgePadding = 2
	s.AppName = r.NewStyle().Bold(true)
	s.CliArgs = r.NewStyle().Foreground(lipgloss.Color("#585858"))
	s.Comment = r.NewStyle().Foreground(lipgloss.Color("#757575"))
	s.CyclingChars = r.NewStyle().Foreground(lipgloss.Color("#FF87D7"))
	s.ErrorHeader = r.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#FF5F87")).Bold(true).Padding(0, 1).SetString("ERROR")
	s.ErrorDetails = s.Comment
	s.ErrPadding = r.NewStyle().Padding(0, horizontalEdgePadding)
	s.Flag = r.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00B594", Dark: "#3EEFCF"}).Bold(true)
	s.FlagComma = r.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#5DD6C0", Dark: "#427C72"}).SetString(",")
	s.FlagDesc = s.Comment
	s.InlineCode = r.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Background(lipgloss.Color("#3A3A3A")).Padding(0, 1)
	s.Link = r.NewStyle().Foreground(lipgloss.Color("#00AF87")).Underline(true)
	s.Quote = r.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#FF71D0", Dark: "#FF78D2"})
	s.Pipe = r.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8470FF", Dark: "#745CFF"})
	s.ConversationList = r.NewStyle().Padding(0, 1)
	s.SHA1 = s.Flag
	s.Timeago = r.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#999", Dark: "#555"})

	// Commit message styles
	s.CommitStep = r.NewStyle().Foreground(lipgloss.Color("#00CED1")).Bold(true)
	s.CommitSuccess = r.NewStyle().Foreground(lipgloss.Color("#32CD32"))

	// Diff styles
	s.DiffHeader = r.NewStyle().Bold(true)
	s.DiffFileHeader = r.NewStyle().Foreground(lipgloss.Color("#00CED1")).Bold(true)
	s.DiffHunkHeader = r.NewStyle().Foreground(lipgloss.Color("#888888")).Bold(true)
	s.DiffAdded = r.NewStyle().Foreground(lipgloss.Color("#00AA00"))
	s.DiffRemoved = r.NewStyle().Foreground(lipgloss.Color("#AA0000"))
	s.DiffContext = r.NewStyle().Foreground(lipgloss.Color("#888888"))

	return s
}
