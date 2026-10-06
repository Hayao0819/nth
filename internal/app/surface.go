package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/router"
)

// pageSurface keeps the timeline mounted and delegates every other page to the
// reactea router. Keeping the timeline alive preserves its loaded posts and
// selection when another page is open.
type pageSurface struct {
	reactea.BasicComponent

	root  *root
	pages *router.Component
}

func (h *pageSurface) Init(ctx *reactea.Ctx) tea.Cmd {
	commands := []tea.Cmd{h.timeline().Init(ctx)}
	if ctx.Route() != timelineRoute {
		commands = append(commands, h.pages.Init(h.componentCtx(ctx)))
	}
	h.sync(ctx)

	return tea.Batch(commands...)
}

func (h *pageSurface) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if reactea.IsInput(msg) {
		if ctx.Route() == timelineRoute {
			return h.timeline().Update(ctx, msg)
		}

		child := h.componentCtx(ctx)
		if reactea.IsMouse(msg) {
			outerX, outerY := ctx.Origin()
			innerX, innerY := child.Origin()
			msg = reactea.TranslateMouse(msg, innerX-outerX, innerY-outerY)
			if _, _, inside := reactea.Mouse(child, msg); !inside {
				return nil
			}
		}

		return h.pages.Update(child, msg)
	}

	commands := []tea.Cmd{h.timeline().Update(ctx, msg)}
	if ctx.Route() != timelineRoute || h.pages.Current() != nil {
		commands = append(commands, h.pages.Update(h.componentCtx(ctx), msg))
	}
	h.sync(ctx)

	return tea.Batch(commands...)
}

func (h *pageSurface) Render(ctx *reactea.Ctx) string {
	if ctx.Route() == timelineRoute {
		return h.timeline().Render(ctx)
	}

	child := h.componentCtx(ctx)
	content := h.pages.Render(child)
	if !h.root.wide {
		return ui.Fit(content, ctx.Width(), ctx.Height())
	}

	return h.root.theme.Column.
		Width(ctx.Width()).
		Height(ctx.Height()).
		MaxWidth(ctx.Width()).
		MaxHeight(ctx.Height()).
		Render(ui.Fit(content, child.Width(), child.Height()))
}

func (h *pageSurface) FocusNext() bool { return reactea.FocusOf(h.active()).FocusNext() }

func (h *pageSurface) FocusPrev() bool { return reactea.FocusOf(h.active()).FocusPrev() }

func (h *pageSurface) FocusFirst() { reactea.FocusOf(h.active()).FocusFirst() }

func (h *pageSurface) FocusLast() { reactea.FocusOf(h.active()).FocusLast() }

func (h *pageSurface) HasFocusable() bool { return reactea.FocusOf(h.active()).HasFocusable() }

func (h *pageSurface) active() reactea.Component {
	if h.root.page.kind == timelinePage {
		return h.timeline()
	}

	return h.pages
}

func (h *pageSurface) timeline() reactea.Component {
	if h.root.wide {
		return h.root.timeline
	}

	return h.root.compact
}

func (h *pageSurface) componentCtx(ctx *reactea.Ctx) *reactea.Ctx {
	if !h.root.wide {
		return ctx
	}
	style := h.root.theme.Column

	return ctx.Inset(
		style.GetMarginLeft()+style.GetBorderLeftSize()+style.GetPaddingLeft(),
		style.GetMarginTop()+style.GetBorderTopSize()+style.GetPaddingTop(),
		max(0, ctx.Width()-style.GetHorizontalFrameSize()),
		max(0, ctx.Height()-style.GetVerticalFrameSize()),
	)
}

func (h *pageSurface) sync(ctx *reactea.Ctx) {
	h.root.page = pageStateForRoute(ctx.Route())
}

var _ reactea.Component = (*pageSurface)(nil)
var _ reactea.Focuser = (*pageSurface)(nil)
