package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/charmbracelet/x/ansi"
)

func DialogSurface(style lipgloss.Style, content string) string {
	foreground := ansi.Style{}.
		ForegroundColor(style.GetForeground()).
		String()
	background := ansi.Style{}.
		BackgroundColor(style.GetBackground()).
		String()
	base := ansi.Style{}.
		ForegroundColor(style.GetForeground()).
		BackgroundColor(style.GetBackground()).
		String()

	return strings.NewReplacer(
		ansi.ResetStyle, ansi.ResetStyle+base,
		"\x1b[0m", "\x1b[0m"+base,
		"\x1b[39;49m", "\x1b[39;49m"+base,
		"\x1b[49;39m", "\x1b[49;39m"+base,
		"\x1b[39m", "\x1b[39m"+foreground,
		"\x1b[49m", "\x1b[49m"+background,
	).Replace(content)
}

func RenderDialog(style lipgloss.Style, content string, width, height int) string {
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	content = ui.Fit(content, innerWidth, innerHeight)
	content = DialogSurface(style, content)

	return style.
		Width(width).Height(height).
		MaxWidth(width).MaxHeight(height).
		Render(content)
}
