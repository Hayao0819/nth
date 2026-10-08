package pageheader

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

func RenderWithAction(theme ui.Theme, title, right, action string, width int) string {
	if action == "" {
		return Render(theme, title, right, width)
	}

	left := theme.Active.Render("←") + "  " + theme.Heading.Render(title)
	button := theme.Button.Padding(0, 1).Render(action)
	available := max(0, width-lipgloss.Width(left)-lipgloss.Width(button)-3)
	rightSide := button
	if right != "" && available > 0 {
		rightSide = ui.Clip(right, available) + "  " + button
	}

	return strings.Join([]string{"", ui.Sides(left, rightSide, width), ""}, "\n")
}

func BackAt(x, y int) bool {
	return x >= 0 && x < 6 && y >= 0 && y < Height
}

func ActionAt(theme ui.Theme, x, y, width int, action string) bool {
	if action == "" || width <= 0 || y < 0 || y >= Height {
		return false
	}
	actionWidth := lipgloss.Width(theme.Button.Padding(0, 1).Render(action))

	return x >= max(0, width-actionWidth) && x < width
}
