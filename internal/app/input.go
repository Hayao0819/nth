package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/feed"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (r *root) handleInput(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if reactea.Key(msg, "ctrl+c") {
		return tea.Quit
	}
	if ctx.InputCaptured() {
		return r.Wrapper.Update(ctx, msg)
	}
	if reactea.IsInput(msg) {
		r.notice = ""
		r.problem = nil
		r.postFailed = false
		r.failureJob = ""
	}
	if r.page.kind == searchPage {
		return r.Wrapper.Update(ctx, msg)
	}
	if command, handled := r.handleChromeClick(ctx, msg); handled {
		return command
	}

	switch {
	case reactea.Key(msg, "q"):
		return tea.Quit
	case reactea.Key(msg, "?"):
		return modal.PushAt(
			ctx,
			dialog.NewHelp(r.theme, r.notifications != nil, r.messages != nil, r.bookmarks != nil),
			dialog.Placement(ctx, 72, 26),
		)
	case reactea.Key(msg, "c", "n"):
		return r.compose(ctx, nil, nil)
	case reactea.Key(msg, "/"):
		return r.search(ctx)
	case reactea.Key(msg, "1"):
		return r.openTimeline(ctx, feed.Home)
	case r.notifications != nil && reactea.Key(msg, "2", "v"):
		return r.openNotifications(ctx)
	case r.messages != nil && reactea.Key(msg, "3"):
		return r.openMessages(ctx)
	case r.bookmarks != nil && reactea.Key(msg, "4"):
		return r.openBookmarks(ctx)
	case r.page.kind == timelinePage && reactea.Key(msg, "tab"):
		return r.cycleTimeline(ctx, 1)
	case r.page.kind == timelinePage && reactea.Key(msg, "shift+tab"):
		return r.cycleTimeline(ctx, -1)
	case reactea.Key(msg, ".") && (r.meErr != nil || r.unreadErr != nil):
		commands := []tea.Cmd{r.Wrapper.Update(ctx, msg)}
		if r.meErr != nil {
			commands = append(commands, r.loadMe(ctx))
		}
		if r.unreadErr != nil {
			commands = append(commands, r.loadUnread(ctx))
		}

		return tea.Batch(commands...)
	case reactea.Key(msg, ".") && r.trendErr != nil:
		return r.loadTrends(ctx)
	case r.wide && r.trendDismiss != nil && reactea.Key(msg, "alt+6", "alt+7", "alt+8", "alt+9", "alt+0"):
		if tag, ok := r.trendTag(msg, "alt+"); ok {
			return r.dismissTrend(ctx.Context(), tag)
		}
	case r.wide && reactea.Key(msg, "6", "7", "8", "9", "0"):
		if query, ok := r.trendQuery(msg); ok {
			return r.searchFor(ctx, query)
		}
	case reactea.Key(msg, "esc") && r.page.kind == timelinePage && r.feed.CurrentMode() == feed.Search:
		return r.feed.SetMode(ctx, feed.Home, "")
	case reactea.Key(msg, "r", "R") && r.page.kind == timelinePage:
		if post := r.feed.SelectedPost(); post != nil {
			return r.handlePostAction(ctx, postcomponent.Reply, *post)
		}
	case reactea.Key(msg, "u") && r.page.kind == timelinePage:
		if post := r.feed.SelectedPost(); post != nil && post.DisplayPost() != nil {
			return r.openUser(ctx, post.DisplayPost().Author)
		}
	case reactea.Key(msg, "U") && r.page.kind == timelinePage:
		if post := r.feed.SelectedPost(); post != nil {
			return r.openNextLinkedUser(ctx, *post)
		}
	case reactea.Key(msg, "a"):
		if r.me != nil {
			return r.openUser(ctx, *r.me)
		}
	case reactea.Key(msg, "Q") && r.page.kind == timelinePage:
		if post := r.feed.SelectedPost(); post != nil {
			return r.handlePostAction(ctx, postcomponent.Quote, *post)
		}
	case reactea.Key(msg, "enter") && r.page.kind == timelinePage:
		if post := r.feed.SelectedPost(); post != nil {
			return r.openPost(ctx, *post)
		}
	}

	return r.Wrapper.Update(ctx, msg)
}
