package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/layout"
)

const (
	mediumLayoutMinWidth = 88
	mediumLayoutMaxWidth = 88
	wideLayoutMinWidth   = 152
	wideLayoutMinHeight  = 20
	wideLayoutMaxWidth   = 152
	wideSidebarWidth     = 28
	wideAsideWidth       = 32
)

func (r *root) handleChromeClick(ctx *reactea.Ctx, msg tea.Msg) (tea.Cmd, bool) {
	if r.page.kind != timelinePage {
		return nil, false
	}
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return nil, false
	}
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil, false
	}
	if !r.wide {
		if x < r.contentLeft || x >= r.contentLeft+r.contentWidth {
			return nil, false
		}
		x -= r.contentLeft
		width := r.contentWidth
		if y != 1 {
			return nil, false
		}
		if r.feed.CurrentMode() == feed.Search {
			return r.feed.SetMode(ctx, feed.Home, ""), true
		}
		if width < 64 {
			if x < width/2 {
				return r.cycleTimeline(ctx, -1), true
			}

			return r.cycleTimeline(ctx, 1), true
		}
		if x < width/2 {
			return r.feed.SetMode(ctx, feed.Ranked, ""), true
		}

		return r.feed.SetMode(ctx, feed.Home, ""), true
	}

	return nil, false
}

func (r *root) configureLayout(width, height int) {
	wide := width >= wideLayoutMinWidth && height >= wideLayoutMinHeight
	medium := !wide && width >= mediumLayoutMinWidth
	if r.layoutSet && r.wide == wide && r.medium == medium && r.layoutWidth == width && r.layoutHeight == height {
		return
	}
	r.layoutSet = true
	r.wide = wide
	r.medium = medium
	r.layoutWidth = width
	r.layoutHeight = height

	if wide {
		contentWidth := min(width, wideLayoutMaxWidth)
		left := max(0, (width-contentWidth)/2)
		right := max(0, width-contentWidth-left)
		r.contentLeft = left
		r.contentWidth = contentWidth
		r.main.SetItems(
			layout.Fixed(wideSidebarWidth, r.sidebar).Key("navigation"),
			layout.Grow(1, r.surface).Bounds(58, 0).Focusable().Key("surface"),
			layout.Fixed(wideAsideWidth, r.aside).Key("aside"),
		)
		r.view.SetItems(
			layout.Spacer(left).Key("left-gutter"),
			layout.Fixed(contentWidth, r.main).Key("content"),
			layout.Spacer(right).Key("right-gutter"),
		)
		r.statusView.SetItems(
			layout.Spacer(left).Key("left-gutter"),
			layout.Fixed(contentWidth, r.statusBar).Key("content"),
			layout.Spacer(right).Key("right-gutter"),
		)
		r.body.SetItems(
			layout.Grow(1, r.view).Key("main"),
			layout.Fixed(1, r.statusView).Key("status"),
		)

		return
	}
	if medium {
		contentWidth := min(width, mediumLayoutMaxWidth)
		left := max(0, (width-contentWidth)/2)
		right := max(0, width-contentWidth-left)
		r.contentLeft = left
		r.contentWidth = contentWidth
		r.view.SetItems(
			layout.Spacer(left).Key("left-gutter"),
			layout.Fixed(contentWidth, r.surface).Focusable().Key("content"),
			layout.Spacer(right).Key("right-gutter"),
		)
		r.body.SetItems(layout.Grow(1, r.view).Key("main"))

		return
	}
	r.contentLeft = 0
	r.contentWidth = width

	r.body.SetItems(layout.Grow(1, r.surface).Focusable().Key("main"))
}
