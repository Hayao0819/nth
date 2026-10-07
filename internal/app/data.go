package app

import (
	"context"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postfeature "github.com/Hayao0819/nth/internal/features/post"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/state"
)

type accountResult struct {
	user north.User
	resp *north.Response
	err  error
}

type unreadResult struct {
	count int
	resp  *north.Response
	err   error
}

type messageUnreadResult struct {
	count int
	resp  *north.Response
	err   error
}

type trendsResult struct {
	items []north.Trend
	err   error
}

type trendDismissedMsg struct {
	target   *root
	tag      string
	response *north.Response
	err      error
}

func (m trendDismissedMsg) Response() *north.Response { return m.response }

type rootResources struct {
	account  state.Resource[accountResult]
	unread   state.Resource[unreadResult]
	dmUnread state.Resource[messageUnreadResult]
	trends   state.Resource[trendsResult]
}

func (r *root) loadMe(ctx *reactea.Ctx) tea.Cmd {
	return r.resources.account.Load(ctx, func(request context.Context) (accountResult, error) {
		user, response, err := r.api.Me(request)

		return accountResult{user: user, resp: response, err: err}, nil
	})
}

func (r *root) loadUnread(ctx *reactea.Ctx) tea.Cmd {
	if r.notifications == nil {
		return nil
	}

	return r.resources.unread.Load(ctx, func(request context.Context) (unreadResult, error) {
		count, response, err := r.notifications.NotificationUnreadCount(request)

		return unreadResult{count: count, resp: response, err: err}, nil
	})
}

func (r *root) loadDMUnread(ctx *reactea.Ctx) tea.Cmd {
	if r.messageUnread == nil {
		return nil
	}

	return r.resources.dmUnread.Load(ctx, func(request context.Context) (messageUnreadResult, error) {
		count, response, err := r.messageUnread.DMUnreadCount(request)

		return messageUnreadResult{count: count, resp: response, err: err}, nil
	})
}

func (r *root) loadTrends(ctx *reactea.Ctx) tea.Cmd {
	if r.trendAPI == nil || r.resources.trends.Loading() {
		return nil
	}
	r.trendBusy = true

	return r.resources.trends.Load(ctx, func(request context.Context) (trendsResult, error) {
		items, _, err := r.trendAPI.Trends(request, "")

		return trendsResult{items: items, err: err}, nil
	})
}

func (r *root) handleResource(msg tea.Msg) (tea.Cmd, bool) {
	if dismissed, ok := msg.(trendDismissedMsg); ok {
		if dismissed.target != r {
			return nil, false
		}
		if dismissed.err != nil {
			r.notice = ui.FriendlyError(dismissed.err)

			return nil, true
		}
		r.trends = slices.DeleteFunc(r.trends, func(trend north.Trend) bool {
			return trend.Tag == dismissed.tag
		})
		r.notice = "Trend hidden"

		return nil, true
	}
	switch {
	case r.resources.account.Handle(msg):
		result := r.resources.account.Value()
		r.setResponse(result.resp)
		r.meErr = result.err
		if result.err != nil {
			return nil, true
		}
		r.me = &result.user
		if page, ok := r.currentPage().(interface{ SetViewer(north.User) }); ok {
			page.SetViewer(result.user)
		}

		return func() tea.Msg { return postfeature.ManagementUpdate{Account: result.user} }, true

	case r.resources.unread.Handle(msg):
		result := r.resources.unread.Value()
		r.unread, r.unreadErr = result.count, result.err
		r.setResponse(result.resp)

		return nil, true

	case r.resources.dmUnread.Handle(msg):
		result := r.resources.dmUnread.Value()
		r.dmUnread, r.dmUnreadErr = result.count, result.err
		r.setResponse(result.resp)

		return nil, true

	case r.resources.trends.Handle(msg):
		result := r.resources.trends.Value()
		r.trendBusy = false
		r.trendErr = result.err
		if result.err == nil {
			r.trends = append([]north.Trend(nil), result.items...)
		}

		return nil, true
	}

	return nil, false
}

func (r *root) trendQuery(msg tea.Msg) (string, bool) {
	tag, ok := r.trendTag(msg, "")
	if !ok {
		return "", false
	}
	for _, trend := range r.trends {
		if trend.Tag != tag {
			continue
		}
		if trend.IsHashtag && !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}

		return tag, true
	}

	return "", false
}

func (r *root) trendTag(msg tea.Msg, prefix string) (string, bool) {
	keys := [...]string{"6", "7", "8", "9", "0"}
	for index, key := range keys {
		if !reactea.Key(msg, prefix+key) || index >= len(r.trends) {
			continue
		}

		return r.trends[index].Tag, true
	}

	return "", false
}

func (r *root) dismissTrend(ctx context.Context, tag string) tea.Cmd {
	if r.trendDismiss == nil || strings.TrimSpace(tag) == "" {
		return nil
	}
	r.notice = "Hiding trend…"

	return func() tea.Msg {
		_, response, err := r.trendDismiss.DismissTrend(ctx, tag)

		return trendDismissedMsg{target: r, tag: tag, response: response, err: err}
	}
}
