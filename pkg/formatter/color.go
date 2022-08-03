package formatter

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	defaultStyle = lipgloss.NewStyle()
	// styles
	text    = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	basicTextStyle      = lipgloss.NewStyle().Foreground(text).Padding(0, 1)
	boldItalicTextStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				Italic(true).
				Foreground(lipgloss.Color("#FFFFFF"))
	TitleHighlight = lipgloss.NewStyle().
			Background(lipgloss.Color("#F25D94")).
			MarginTop(1).
			MarginLeft(10).
			MarginRight(10).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB"))

	textHighlight = lipgloss.NewStyle().
			MarginRight(1).
			MarginLeft(1).
			Padding(0, 1).
			Foreground(special)

	mainStype = defaultStyle.Copy()

	tableStype = defaultStyle.Copy()

	vethPairStyle = defaultStyle.Copy()

	bridgeStyle = lipgloss.NewStyle().
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

	unusedVethStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#F25D94")).
			MarginLeft(5).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB"))
)

func colorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#F25D94")
	x1y0, _ := colorful.Hex("#EDFF82")
	x0y1, _ := colorful.Hex("#643AFF")
	x1y1, _ := colorful.Hex("#14F9D5")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}
