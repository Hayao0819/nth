package app

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/feed"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	bookmarkpage "github.com/Hayao0819/nth/internal/pages/bookmark"
	infopage "github.com/Hayao0819/nth/internal/pages/info"
	messagepage "github.com/Hayao0819/nth/internal/pages/message"
	notificationpage "github.com/Hayao0819/nth/internal/pages/notification"
	postpage "github.com/Hayao0819/nth/internal/pages/post"
	searchpage "github.com/Hayao0819/nth/internal/pages/search"
	userpage "github.com/Hayao0819/nth/internal/pages/user"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (r *root) loadMe(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		user, resp, err := r.api.Me(ctx)

		return meLoadedMsg{target: r, user: user, resp: resp, err: err}
	}
}

func (r *root) loadUnread(ctx context.Context) tea.Cmd {
	if r.notifications == nil {
		return nil
	}

	return func() tea.Msg {
		count, response, err := r.notifications.NotificationUnreadCount(ctx)
		return unreadLoadedMsg{target: r, count: count, resp: response, err: err}
	}
}

func (r *root) loadTrends(ctx context.Context) tea.Cmd {
	if r.trendAPI == nil || r.trendBusy {
		return nil
	}
	r.trendBusy = true

	return func() tea.Msg {
		items, _, err := r.trendAPI.Trends(ctx, "")

		return trendsLoadedMsg{target: r, items: items, err: err}
	}
}

func (r *root) trendQuery(msg tea.Msg) (string, bool) {
	keys := [...]string{"6", "7", "8", "9", "0"}
	for index, key := range keys {
		if !reactea.Key(msg, key) || index >= len(r.trends) {
			continue
		}
		trend := r.trends[index]
		query := trend.Tag
		if trend.IsHashtag && !strings.HasPrefix(query, "#") {
			query = "#" + query
		}

		return query, true
	}

	return "", false
}

func (r *root) openNotifications(ctx *reactea.Ctx) tea.Cmd {
	if r.notifications == nil {
		return nil
	}

	return r.pushPage(ctx, pageState{
		kind:      notificationsPage,
		key:       "notifications",
		component: notificationpage.NewPageWithImages(r.notifications, r.theme, r.images),
	})
}

func (r *root) openMessages(ctx *reactea.Ctx) tea.Cmd {
	if r.messages == nil {
		return nil
	}

	return r.pushPage(ctx, pageState{
		kind:      messagesPage,
		key:       "messages",
		component: messagepage.New(r.messages, r.theme),
	})
}

func (r *root) openBookmarks(ctx *reactea.Ctx) tea.Cmd {
	if r.bookmarks == nil {
		return nil
	}

	return r.pushPage(ctx, pageState{
		kind:      bookmarksPage,
		key:       "bookmarks",
		component: bookmarkpage.New(r.api, r.bookmarks, r.theme, r.images),
	})
}

func (r *root) openLists(ctx *reactea.Ctx) tea.Cmd {
	return r.pushPage(ctx, pageState{
		kind: listsPage,
		key:  "lists",
		component: infopage.New(
			r.theme,
			"Lists",
			"Lists are not available in this version of nth.",
		),
	})
}

func (r *root) openSettings(ctx *reactea.Ctx) tea.Cmd {
	return r.pushPage(ctx, pageState{
		kind: settingsPage,
		key:  "settings",
		component: infopage.New(
			r.theme,
			"Settings",
			"Run nth setup to change authentication, browser, and image settings.",
		),
	})
}

func (r *root) search(ctx *reactea.Ctx) tea.Cmd {
	current := ""
	if page, ok := r.page.component.(*searchpage.Screen); ok {
		current = page.Query()
	} else if r.feed.CurrentMode() == feed.Search {
		current = r.feed.SearchQuery()
	}

	return r.searchFor(ctx, current)
}

func (r *root) searchFor(ctx *reactea.Ctx, query string) tea.Cmd {
	return r.pushPage(ctx, pageState{
		kind:      searchPage,
		key:       "search",
		component: searchpage.NewWithImages(r.api, r.theme, query, r.images),
	})
}

func (r *root) cycleTimeline(ctx *reactea.Ctx, delta int) tea.Cmd {
	modes := [...]feed.Mode{feed.Home, feed.Ranked}
	index := 0
	for i, mode := range modes {
		if mode == r.feed.CurrentMode() {
			index = i
			break
		}
	}
	if r.feed.CurrentMode() == feed.Search {
		if delta < 0 {
			index = len(modes) - 1
		} else {
			index = 0
		}
	} else {
		index = (index + delta + len(modes)) % len(modes)
	}

	return r.feed.SetMode(ctx, modes[index], "")
}

func (r *root) openTimeline(ctx *reactea.Ctx, mode feed.Mode) tea.Cmd {
	r.showTimeline()
	if r.feed.CurrentMode() == feed.Home || r.feed.CurrentMode() == feed.Ranked {
		return nil
	}

	return r.feed.SetMode(ctx, mode, "")
}

func (r *root) compose(ctx *reactea.Ctx, replyTo, quote *north.Post) tea.Cmd {
	if r.posting {
		r.notice = "A post is already being sent"

		return nil
	}

	key := composeDraftKey(postID(replyTo), postID(quote))

	return modal.PushAt(ctx, dialog.NewCompose(r.theme, replyTo, quote, r.drafts[key]), dialog.Placement(ctx, 72, 16))
}

func (r *root) edit(ctx *reactea.Ctx, post north.Post) tea.Cmd {
	if r.editing {
		r.notice = "A post is already being updated"

		return nil
	}
	target := post.DisplayPost()
	if target == nil {
		r.notice = "This post cannot be edited"

		return nil
	}

	return modal.PushAt(
		ctx,
		dialog.NewEdit(r.theme, *target, r.drafts["edit:"+target.ID]),
		dialog.Placement(ctx, 72, 16),
	)
}

func (r *root) submit(ctx context.Context, post dialog.Submission) tea.Cmd {
	if r.posting {
		return nil
	}
	r.posting = true
	r.notice = "Posting…"
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""
	draftKey := composeDraftKey(post.ReplyTo, post.QuoteID)

	req := north.CreatePostRequest{Text: post.Text, QuotePostID: post.QuoteID}
	if post.ReplyTo != "" {
		req.Reply = &north.CreatePostReply{InReplyToPostID: post.ReplyTo}
	}

	return func() tea.Msg {
		_, resp, err := r.api.CreatePost(ctx, req)

		return postCreatedMsg{target: r, draftKey: draftKey, resp: resp, err: err}
	}
}

func (r *root) openPost(ctx *reactea.Ctx, post north.Post) tea.Cmd {
	editor, _ := r.api.(postpage.Editor)
	key := post.ID
	if target := post.DisplayPost(); target != nil && strings.TrimSpace(target.ID) != "" {
		key = target.ID
	}
	if strings.TrimSpace(key) == "" {
		r.notice = "This post is unavailable"

		return nil
	}

	return r.pushPage(ctx, pageState{
		kind:      postPage,
		key:       key,
		component: postpage.NewPageWithEditorAndImages(r.api, editor, r.theme, post, r.ownsPost(post), r.images),
	})
}

func (r *root) openUser(ctx *reactea.Ctx, user north.User) tea.Cmd {
	handle := strings.TrimSpace(user.Handle)
	if handle == "" {
		r.notice = "This account is unavailable"

		return nil
	}

	return r.pushPage(ctx, pageState{
		kind:      profilePage,
		key:       strings.ToLower(handle),
		component: userpage.NewPageWithImages(r.api, r.theme, user, r.images),
	})
}

func (r *root) openNextLinkedUser(ctx *reactea.Ctx, post north.Post) tea.Cmd {
	users := postcomponent.LinkedUsers(post)
	if len(users) == 0 {
		r.notice = "This post has no available account"

		return nil
	}
	target := post.DisplayPost()
	key := post.ID
	if target != nil && target.ID != "" {
		key = target.ID
	}
	index := r.linkedUser[key]
	if len(users) > 1 {
		index = (index + 1) % len(users)
	} else {
		index = 0
	}
	r.linkedUser[key] = index

	return r.openUser(ctx, users[index])
}

func (r *root) handlePostAction(ctx *reactea.Ctx, action postcomponent.Action, post north.Post) tea.Cmd {
	if action == postcomponent.ViewPost {
		return r.openPost(ctx, post)
	}
	if action == postcomponent.ViewAuthor {
		target := post.DisplayPost()
		if target == nil {
			r.notice = "This account is unavailable"

			return nil
		}

		return r.openUser(ctx, target.Author)
	}
	if !postcomponent.CanInteract(&post) {
		r.notice = "This post cannot be acted on"

		return nil
	}
	target := post.DisplayPost()
	switch action {
	case postcomponent.Reply:
		return r.compose(ctx, target, nil)
	case postcomponent.Repost:
		return r.feed.ToggleRepostPost(ctx.Context(), &post)
	case postcomponent.Like:
		return r.feed.ToggleLikePost(ctx.Context(), &post)
	case postcomponent.Quote:
		return r.compose(ctx, nil, target)
	case postcomponent.Edit:
		if !r.ownsPost(post) {
			r.notice = "Only your own posts can be edited"

			return nil
		}
		if _, ok := r.api.(postpage.Editor); !ok {
			r.notice = "Editing requires browser authentication"

			return nil
		}

		return r.edit(ctx, post)
	case postcomponent.Delete:
		if !r.ownsPost(post) {
			r.notice = "Only your own posts can be deleted"

			return nil
		}

		return r.deletePost(ctx.Context(), target.ID)
	default:
		return nil
	}
}

func (r *root) editPost(ctx context.Context, id, text string, mediaIDs []string) tea.Cmd {
	if r.editing {
		return nil
	}
	editor, ok := r.api.(postpage.Editor)
	if !ok {
		r.notice = "Editing requires browser authentication"

		return nil
	}
	r.editing = true
	r.notice = "Updating post…"
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""

	return func() tea.Msg {
		response, err := editor.EditPost(ctx, id, text, mediaIDs)

		return postEditedMsg{target: r, postID: id, text: text, resp: response, err: err}
	}
}

func (r *root) deletePost(ctx context.Context, id string) tea.Cmd {
	if r.deleting {
		r.notice = "A post is already being deleted"

		return nil
	}
	r.deleting = true
	r.notice = "Deleting…"
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""

	return func() tea.Msg {
		deleted, resp, err := r.api.DeletePost(ctx, id)

		return postDeletedMsg{target: r, postID: id, deleted: deleted, resp: resp, err: err}
	}
}

func (r *root) ownsPost(post north.Post) bool {
	if r.me == nil {
		return false
	}
	target := post.DisplayPost()
	if target == nil {
		return false
	}
	if r.me.ID != "" && target.Author.ID != "" {
		return r.me.ID == target.Author.ID
	}

	return strings.EqualFold(r.me.Handle, target.Author.Handle)
}

func (r *root) rememberDraft(key, text string) {
	if strings.TrimSpace(text) == "" {
		delete(r.drafts, key)

		return
	}
	r.drafts[key] = text
}

func composeDraftKey(replyTo, quoteID string) string {
	switch {
	case replyTo != "":
		return "reply:" + replyTo
	case quoteID != "":
		return "quote:" + quoteID
	default:
		return "post"
	}
}

func submissionDraftKey(submission dialog.Submission) string {
	if submission.EditID != "" {
		return "edit:" + submission.EditID
	}

	return composeDraftKey(submission.ReplyTo, submission.QuoteID)
}

func postID(post *north.Post) string {
	if post == nil {
		return ""
	}

	return post.ID
}
