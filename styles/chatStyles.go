package styles

import "github.com/charmbracelet/lipgloss"

var (
	ViewportsStyle = lipgloss.NewStyle().
			Padding(0, 2, 0, 1)

	FullChatViewStyle = ViewportsStyle.BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder())

	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	SenderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	UnactiveButtonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Border(lipgloss.NormalBorder())
	ActiveButtonStyle   = UnactiveButtonStyle.BorderBackground(lipgloss.Color("#fff")).Background(lipgloss.Color("#fff"))
)
