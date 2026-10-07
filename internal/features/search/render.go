package search

import (
	"strings"

	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	if width <= 0 || height <= 0 {
		return ""
	}

	input := s.input.Render(s.inputCtx(ctx))
	header := "\n" + s.theme.Active.Render("←") + strings.Repeat(" ", 5) + input + "\n"
	room := max(0, height-pageheader.Height)
	body := ""
	if strings.TrimSpace(s.input.Widget.Value()) == "" {
		body = "\n  " + s.theme.Heading.Render("Search north") +
			"\n  " + s.theme.Dim.Render("Results appear when you pause typing")
	} else if s.pending {
		body = "\n  " + s.theme.Dim.Render("Searching…")
	} else {
		body = s.feed.Render(s.feedCtx(ctx))
	}

	return ui.Fit(header+"\n"+ui.Fit(body, width, room), width, height)
}
