package app

import (
	"context"
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/sessiondialog"
	"github.com/Hayao0819/nth/internal/components/termimage"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain/bookmark"
	messagedomain "github.com/Hayao0819/nth/internal/domain/message"
	"github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/domain/session"
	notificationpage "github.com/Hayao0819/nth/internal/pages/notification"
	postpage "github.com/Hayao0819/nth/internal/pages/post"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/layout"
	"github.com/Hayao0819/reactea/v2/modal"
)

type meLoadedMsg struct {
	target *root
	user   north.User
	resp   *north.Response
	err    error
}

type postCreatedMsg struct {
	target   *root
	draftKey string
	resp     *north.Response
	err      error
}

type postDeletedMsg struct {
	target  *root
	postID  string
	deleted bool
	resp    *north.Response
	err     error
}

type postEditedMsg struct {
	target *root
	postID string
	text   string
	resp   *north.Response
	err    error
}

type unreadLoadedMsg struct {
	target *root
	count  int
	resp   *north.Response
	err    error
}

type trendsLoadedMsg struct {
	target *root
	items  []north.Trend
	err    error
}

type trendAPI interface {
	Trends(context.Context, string) ([]north.Trend, *north.Response, error)
}

type root struct {
	reactea.Wrapper

	api           API
	notifications notification.API
	bookmarks     bookmarkdomain.API
	messages      messagedomain.API
	trendAPI      trendAPI
	images        *termimage.Renderer
	theme         ui.Theme
	feed          *feed.Feed
	body          *layout.Box
	main          *layout.Box
	view          *layout.Box
	statusView    *layout.Box
	compact       *layout.Box
	surface       *pageHost
	page          pageState
	history       []pageState

	compactHeader reactea.Component
	compactFooter reactea.Component
	timeline      reactea.Component
	sidebar       reactea.Component
	aside         reactea.Component
	statusBar     reactea.Component
	wide          bool
	medium        bool
	layoutSet     bool
	layoutWidth   int
	layoutHeight  int
	contentLeft   int
	contentWidth  int

	me         *north.User
	meErr      error
	posting    bool
	editing    bool
	deleting   bool
	drafts     map[string]string
	notice     string
	problem    error
	postFailed bool
	failureJob string
	resp       *north.Response
	respAt     time.Time
	unread     int
	unreadErr  error
	trends     []north.Trend
	trendErr   error
	trendBusy  bool
	linkedUser map[string]int
}

func newRoot(api API) *root {
	return newRootWithServices(api, nil, termimage.New(false))
}

func newRootWithTrends(api API, trends trendAPI) *root {
	return newRootWithServices(api, trends, termimage.New(false))
}

func newRootWithServices(api API, trends trendAPI, images *termimage.Renderer) *root {
	if images == nil {
		images = termimage.New(false)
	}
	theme := ui.NewTheme()
	feed := feed.NewWithImages(api, theme, images)
	r := &root{
		api:        api,
		trendAPI:   trends,
		images:     images,
		theme:      theme,
		feed:       feed,
		drafts:     make(map[string]string),
		linkedUser: make(map[string]int),
	}
	r.notifications, _ = api.(notification.API)
	r.bookmarks, _ = api.(bookmarkdomain.API)
	r.messages, _ = api.(messagedomain.API)

	r.compactHeader = layout.Memo(reactea.Func(r.renderHeader), func() any {
		return r.compactHeaderKey()
	})
	r.compactFooter = reactea.Func(r.renderFooter)
	sidebar := &sidebar{root: r}
	r.sidebar = layout.Memo(sidebar, func() any { return sidebar.renderKey() })
	r.aside = &inspector{root: r}
	r.statusBar = reactea.Func(r.renderStatusBar)
	header := &timelineHeader{root: r}
	r.timeline = layout.Framed(
		theme.Column,
		layout.Column(
			layout.Fixed(3, layout.Memo(header, func() any {
				return r.feed.CurrentMode()
			})),
			layout.Grow(1, feed).Focusable(),
		),
	)
	r.main = layout.Row()
	r.view = layout.Row()
	r.statusView = layout.Row()
	r.compact = layout.Column(
		layout.Fixed(2, r.compactHeader).Key("header"),
		layout.Grow(1, feed).Focusable().Key("timeline"),
		layout.Fixed(2, r.compactFooter).Key("footer"),
	)
	r.page = pageState{kind: timelinePage}
	r.surface = &pageHost{root: r}
	r.body = layout.Column()
	r.Wrapper = reactea.Wrap(r.body)
	r.configureLayout(0, 0)

	return r
}

func (r *root) Init(ctx *reactea.Ctx) tea.Cmd {
	r.configureLayout(ctx.Width(), ctx.Height())

	return tea.Batch(
		r.Wrapper.Init(ctx),
		reactea.SetMouseMode(tea.MouseModeCellMotion),
		r.loadMe(ctx.Context()),
		r.loadUnread(ctx.Context()),
		r.loadTrends(ctx.Context()),
	)
}

func (r *root) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if _, resized := msg.(tea.WindowSizeMsg); resized {
		r.configureLayout(ctx.Width(), ctx.Height())
	}
	if carrier, ok := msg.(interface{ Response() *north.Response }); ok {
		r.setResponse(carrier.Response())
	}
	if command, handled := r.images.Update(msg); handled {
		return command
	}

	switch msg := msg.(type) {
	case *session.RefreshRequest:
		if msg == nil {
			return nil
		}

		return modal.Push(ctx, sessiondialog.New(r.theme, msg))

	case unreadLoadedMsg:
		if msg.target != r {
			return nil
		}
		r.unread, r.unreadErr = msg.count, msg.err
		r.setResponse(msg.resp)
		return nil

	case trendsLoadedMsg:
		if msg.target != r {
			return nil
		}
		r.trendBusy = false
		r.trendErr = msg.err
		if msg.err == nil {
			r.trends = append([]north.Trend(nil), msg.items...)
		}

		return nil

	case notificationpage.PageLoadedMsg:
		return r.Wrapper.Update(ctx, msg)

	case notificationpage.ReadMsg:
		if msg.Error() == nil {
			r.unread = 0
			r.unreadErr = nil
		}
		return r.Wrapper.Update(ctx, msg)

	case pageheader.BackMsg:
		r.backPage()
		return nil

	case navigation.OpenPostMsg:
		if r.page.kind == timelinePage && ctx.InputCaptured() {
			return nil
		}
		return r.openPost(ctx, msg.Post)

	case navigation.OpenUserMsg:
		if r.page.kind == timelinePage && ctx.InputCaptured() {
			return nil
		}
		return r.openUser(ctx, msg.User)

	case postcomponent.ActionMsg:
		if r.page.kind == timelinePage && ctx.InputCaptured() {
			return nil
		}
		return r.handlePostAction(ctx, msg.Action, msg.Post)

	case meLoadedMsg:
		if msg.target != r {
			return nil
		}
		r.setResponse(msg.resp)
		r.meErr = msg.err
		if msg.err == nil {
			r.me = &msg.user
			management := func() tea.Msg { return postpage.ManagementUpdate{Account: msg.user} }
			return management
		}

		return nil

	case postCreatedMsg:
		if msg.target != r {
			return nil
		}
		r.posting = false
		r.setResponse(msg.resp)
		if msg.err != nil {
			r.problem, r.notice = msg.err, ""
			r.postFailed = true
			r.failureJob = "Post"

			return nil
		}
		r.problem = nil
		r.postFailed = false
		r.failureJob = ""
		r.notice = "Post sent"
		delete(r.drafts, msg.draftKey)

		return r.feed.Refresh(ctx)

	case postDeletedMsg:
		if msg.target != r {
			return nil
		}
		r.deleting = false
		r.setResponse(msg.resp)
		if msg.err != nil {
			r.problem, r.notice = msg.err, ""

			return nil
		}
		if !msg.deleted {
			r.problem = errors.New("north did not delete the post")
			r.notice = ""

			return nil
		}
		r.problem = nil
		r.notice = "Post deleted"
		r.feed.RemovePost(ctx, msg.postID)
		if r.page.kind == postPage && r.page.key == msg.postID {
			r.backPage()
		}

		return nil

	case postEditedMsg:
		if msg.target != r {
			return nil
		}
		r.editing = false
		r.setResponse(msg.resp)
		if msg.err != nil {
			r.problem, r.notice = msg.err, ""
			r.postFailed = true
			r.failureJob = "Update"

			return nil
		}
		r.problem = nil
		r.postFailed = false
		r.failureJob = ""
		r.notice = "Post updated"
		delete(r.drafts, "edit:"+msg.postID)
		editedAt := time.Now()
		r.feed.UpdatePost(ctx, msg.postID, msg.text, editedAt)
		if detail, ok := r.page.component.(*postpage.Screen); ok {
			detail.UpdatePost(msg.postID, msg.text, editedAt)
		}

		return nil

	case modal.Result[dialog.Submission]:
		if msg.Ok() {
			key := submissionDraftKey(msg.Value)
			r.rememberDraft(key, msg.Value.Text)
			if msg.Value.Canceled {
				if strings.TrimSpace(msg.Value.Text) != "" {
					r.notice = "Draft saved"
				}

				return nil
			}

			if msg.Value.EditID != "" {
				return r.editPost(ctx.Context(), msg.Value.EditID, msg.Value.Text, msg.Value.MediaIDs)
			}

			return r.submit(ctx.Context(), msg.Value)
		}

		return nil

	case navigationMsg:
		if ctx.InputCaptured() {
			return nil
		}
		switch msg.action {
		case navigateSearch:
			if msg.query != "" {
				return r.searchFor(ctx, msg.query)
			}
			return r.search(ctx)
		case navigateCompose:
			return r.compose(ctx, nil, nil)
		case navigateNotifications:
			return r.openNotifications(ctx)
		case navigateMessages:
			return r.openMessages(ctx)
		case navigateBookmarks:
			return r.openBookmarks(ctx)
		case navigateLists:
			return r.openLists(ctx)
		case navigateProfile:
			if r.me != nil {
				return r.openUser(ctx, *r.me)
			}
			return nil
		case navigateSettings:
			return r.openSettings(ctx)
		default:
			return r.openTimeline(ctx, msg.mode)
		}
	}

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
	if cmd, handled := r.handleChromeClick(ctx, msg); handled {
		return cmd
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
	case reactea.Key(msg, "tab"):
		r.showTimeline()
		return r.cycleTimeline(ctx, 1)
	case reactea.Key(msg, "shift+tab"):
		r.showTimeline()
		return r.cycleTimeline(ctx, -1)
	case reactea.Key(msg, ".") && (r.meErr != nil || r.unreadErr != nil):
		commands := []tea.Cmd{r.Wrapper.Update(ctx, msg)}
		if r.meErr != nil {
			commands = append(commands, r.loadMe(ctx.Context()))
		}
		if r.unreadErr != nil {
			commands = append(commands, r.loadUnread(ctx.Context()))
		}

		return tea.Batch(commands...)
	case reactea.Key(msg, ".") && r.trendErr != nil:
		return r.loadTrends(ctx.Context())
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

var _ reactea.Component = (*root)(nil)
