package app

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/termimage"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain/bookmark"
	messagedomain "github.com/Hayao0819/nth/internal/domain/message"
	"github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/layout"
	"github.com/Hayao0819/reactea/v2/router"
)

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
	surface       *pageSurface
	pages         *router.Component
	page          pageState
	history       []string
	posts         map[string]north.Post
	users         map[string]north.User

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
	resources  rootResources
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
		posts:      make(map[string]north.Post),
		users:      make(map[string]north.User),
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
	r.page = pageStateForRoute(timelineRoute)
	r.pages = router.NewWithRoutes(r.pageRoutes())
	r.surface = &pageSurface{root: r, pages: r.pages}
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
		r.loadMe(ctx),
		r.loadUnread(ctx),
		r.loadTrends(ctx),
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
	if command, handled := r.handleResource(msg); handled {
		return command
	}
	if command, handled := r.handleEvent(ctx, msg); handled {
		return command
	}

	return r.handleInput(ctx, msg)
}

var _ reactea.Component = (*root)(nil)
