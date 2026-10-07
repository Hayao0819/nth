package user

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

type profileTab uint8

const (
	postsTab profileTab = iota
	likesTab
)

func (d *Screen) setTab(ctx context.Context, tab profileTab) tea.Cmd {
	if tab == d.tab || d.postsLoading || tab == likesTab && d.activity == nil {
		return nil
	}
	d.tab = tab
	d.posts = nil
	d.nextCursor = nil
	d.selected = 0
	d.offset = 0
	d.postsErr = nil
	d.fillLoads = 0

	return d.loadPosts(ctx, false)
}

func (d *Screen) cycleTab(ctx context.Context) tea.Cmd {
	if d.activity == nil {
		return nil
	}
	if d.tab == postsTab {
		return d.setTab(ctx, likesTab)
	}

	return d.setTab(ctx, postsTab)
}
