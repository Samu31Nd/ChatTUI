package styles

import "github.com/charmbracelet/lipgloss"

var (
	FocusedStyleForm = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyleForm = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyleForm  = FocusedStyleForm
	NoStyle          = lipgloss.NewStyle()
)
