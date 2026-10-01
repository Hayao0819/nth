package pageheader

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/ui"
)

const Height = 3

type BackMsg struct{}

func Back() tea.Cmd {
	return func() tea.Msg { return BackMsg{} }
}

func Render(theme ui.Theme, title, right string, width int) string {
	center := theme.Active.Render("←") + "  " + theme.Heading.Render(title)
	if right != "" {
		center = ui.Sides(center, right, width)
	} else {
		center = ui.Left(center, width)
	}

	return strings.Join([]string{"", center, ""}, "\n")
}

func BackAt(x, y int) bool {
	return x >= 0 && x < 6 && y >= 0 && y < Height
}
