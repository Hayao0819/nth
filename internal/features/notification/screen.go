package notification

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	notificationdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Screen struct {
	reactea.BasicComponent

	api         notificationdomain.NotificationAPI
	streamAPI   notificationdomain.NotificationStreamAPI
	theme       ui.Theme
	images      *termimage.Renderer
	tab         north.NotificationTab
	stream      *north.NotificationStream
	streaming   bool
	items       []notificationdomain.NotificationItem
	nextCursor  *string
	selected    int
	top         int
	loading     bool
	loadingMore bool
	fillLoads   int
	err         error
	notice      string
	markedRead  bool
	markingRead bool
}

func NewPage(api notificationdomain.NotificationAPI, theme ui.Theme) *Screen {
	return NewPageWithImages(api, theme, nil)
}

func NewPageWithImages(api notificationdomain.NotificationAPI, theme ui.Theme, images *termimage.Renderer) *Screen {
	screen := &Screen{api: api, theme: theme, images: images, tab: north.NotificationsAll}
	if reporter, ok := api.(interface{ SupportsNotificationStream() bool }); !ok || reporter.SupportsNotificationStream() {
		screen.streamAPI, _ = api.(notificationdomain.NotificationStreamAPI)
	}

	return screen
}

func (d *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	return tea.Batch(d.load(ctx.Context(), false), d.connectStream(ctx.Context()))
}

var _ reactea.Component = (*Screen)(nil)
