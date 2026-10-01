package info

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Screen struct {
	reactea.BasicComponent

	theme ui.Theme
	title string
	lines []string
}

func New(theme ui.Theme, title string, lines ...string) *Screen {
	return &Screen{theme: theme, title: title, lines: append([]string(nil), lines...)}
}

func (s *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && pageheader.BackAt(x, y) {
			return pageheader.Back()
		}
	}
	if reactea.Key(msg, "esc", "left") {
		return pageheader.Back()
	}

	return nil
}

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	header := pageheader.Render(s.theme, s.title, s.theme.Dim.Render("Esc/← back"), width)
	body := make([]string, 0, len(s.lines)+2)
	body = append(body, "")
	for _, line := range s.lines {
		for _, wrapped := range ui.WrappedLines(line, max(1, width-4)) {
			body = append(body, "  "+wrapped)
		}
		body = append(body, "")
	}

	return ui.Fit(header+strings.Join(body, "\n"), width, height)
}

var _ reactea.Component = (*Screen)(nil)
