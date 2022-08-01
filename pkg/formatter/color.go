package formatter

import "github.com/charmbracelet/lipgloss"

var (
	// styles

	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	mainStype = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63"))

	vethPairStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63"))

	bridgeStyle = lipgloss.NewStyle().
			MarginRight(5).
			Padding(0, 1).
			Bold(true).
			Italic(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"})

	netNsStyle = lipgloss.NewStyle().
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"})

	vethStyle = lipgloss.NewStyle().
			MarginRight(5).
			Padding(0, 1).
			Italic(false).
			Foreground(lipgloss.Color("#FFFFFF"))
)
