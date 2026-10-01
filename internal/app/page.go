package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type pageKind uint8

const (
	timelinePage pageKind = iota
	notificationsPage
	messagesPage
	bookmarksPage
	listsPage
	profilePage
	settingsPage
	postPage
	searchPage
)

type pageState struct {
	kind      pageKind
	key       string
	component reactea.Component
}

type pageHost struct {
	reactea.BasicComponent
	root *root
}

func (h *pageHost) Init(ctx *reactea.Ctx) tea.Cmd {
	return h.timeline().Init(ctx)
}

func (h *pageHost) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if reactea.IsInput(msg) {
		child := h.componentCtx(ctx, h.root.page)
		if reactea.IsMouse(msg) && child != ctx {
			outerX, outerY := ctx.Origin()
			innerX, innerY := child.Origin()
			msg = reactea.TranslateMouse(msg, innerX-outerX, innerY-outerY)
			if _, _, inside := reactea.Mouse(child, msg); !inside {
				return nil
			}
		}

		return h.active().Update(child, msg)
	}

	commands := []tea.Cmd{h.timeline().Update(ctx, msg)}
	states := append(append([]pageState(nil), h.root.history...), h.root.page)
	for _, state := range states {
		if state.kind == timelinePage || state.component == nil {
			continue
		}
		commands = append(commands, state.component.Update(h.componentCtx(ctx, state), msg))
	}

	return tea.Batch(commands...)
}

func (h *pageHost) Render(ctx *reactea.Ctx) string {
	if h.root.page.kind == timelinePage {
		return h.timeline().Render(ctx)
	}

	child := h.componentCtx(ctx, h.root.page)
	content := h.root.page.component.Render(child)
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

func (h *pageHost) active() reactea.Component {
	if h.root.page.kind == timelinePage || h.root.page.component == nil {
		return h.timeline()
	}

	return h.root.page.component
}

func (h *pageHost) timeline() reactea.Component {
	if h.root.wide {
		return h.root.timeline
	}

	return h.root.compact
}

func (h *pageHost) componentCtx(ctx *reactea.Ctx, state pageState) *reactea.Ctx {
	if !h.root.wide || state.kind == timelinePage {
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

func (r *root) pushPage(ctx *reactea.Ctx, next pageState) tea.Cmd {
	if next.component == nil {
		return nil
	}
	if r.page.kind == next.kind && r.page.key == next.key {
		r.page = next

		return next.component.Init(r.surface.componentCtx(ctx, next))
	}
	r.history = append(r.history, r.page)
	if len(r.history) > 32 {
		r.history = append([]pageState(nil), r.history[len(r.history)-32:]...)
	}
	r.page = next
	r.notice = ""
	r.problem = nil

	return next.component.Init(r.surface.componentCtx(ctx, next))
}

func (r *root) showTimeline() {
	r.page = pageState{kind: timelinePage}
	r.history = nil
}

func (r *root) backPage() {
	if len(r.history) == 0 {
		r.page = pageState{kind: timelinePage}

		return
	}
	r.page = r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
}

var _ reactea.Component = (*pageHost)(nil)
