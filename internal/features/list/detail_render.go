package list

import (
	"strings"

	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Detail) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	title := ui.SafeInline(s.item.Name)
	if title == "" {
		title = "List"
	}
	right := ""
	if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-12)))
	} else {
		states := make([]string, 0, 2)
		if s.item.FollowedByViewer {
			states = append(states, "following")
		}
		if s.item.PinnedByViewer {
			states = append(states, "pinned")
		}
		right = s.theme.Dim.Render(strings.Join(states, " · "))
	}
	header := pageheader.Render(s.theme, title, right, width)
	body := s.feed.Render(s.feedCtx(ctx))
	footer := "j/k move · Enter open · u profile · o owner · f follow · p pin"
	if s.members != nil {
		footer += " · m members"
	}
	if s.item.OwnedByViewer && s.editor != nil {
		footer += " · e edit · d delete"
	}
	footer += " · Esc/← back"
	if progress := s.feed.Progress(); progress != "" {
		footer = ui.Sides(footer, progress, width)
	}

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Detail) feedCtx(ctx *reactea.Ctx) *reactea.Ctx {
	return ctx.Inset(0, pageheader.Height, ctx.Width(), max(0, ctx.Height()-pageheader.Height-1))
}
