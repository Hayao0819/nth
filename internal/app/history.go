package app

import (
	tea "charm.land/bubbletea/v2"
	searchfeature "github.com/Hayao0819/nth/internal/features/search"
	"github.com/Hayao0819/reactea/v2"
)

func (r *root) pushRoute(ctx *reactea.Ctx, target string) tea.Cmd {
	if target == "" || target == ctx.Route() {
		return nil
	}

	current := ctx.Route()
	if search, ok := r.currentPage().(*searchfeature.Screen); ok {
		current = searchRoute(search.Query())
	}
	r.history = append(r.history, current)
	if len(r.history) > 32 {
		r.history = append([]string(nil), r.history[len(r.history)-32:]...)
	}
	r.notice = ""
	r.problem = nil

	return ctx.SetRoute(target)
}

func (r *root) backPage(ctx *reactea.Ctx) tea.Cmd {
	target := timelineRoute
	if len(r.history) > 0 {
		target = r.history[len(r.history)-1]
		r.history = r.history[:len(r.history)-1]
	}
	if target == "" {
		target = timelineRoute
	}
	if target == ctx.Route() {
		return nil
	}

	return ctx.SetRoute(target)
}

func (r *root) showTimeline(ctx *reactea.Ctx) tea.Cmd {
	r.history = nil
	if ctx.Route() == timelineRoute {
		return nil
	}

	return ctx.SetRoute(timelineRoute)
}
