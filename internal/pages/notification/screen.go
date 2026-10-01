package notification

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/termimage"
	notificationdomain "github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Screen struct {
	reactea.BasicComponent

	api         notificationdomain.API
	theme       ui.Theme
	images      *termimage.Renderer
	items       []notificationdomain.Item
	nextCursor  *string
	selected    int
	top         int
	loading     bool
	loadingMore bool
	err         error
	notice      string
	markedRead  bool
	markingRead bool
}

func NewPage(api notificationdomain.API, theme ui.Theme) *Screen {
	return NewPageWithImages(api, theme, nil)
}

func NewPageWithImages(api notificationdomain.API, theme ui.Theme, images *termimage.Renderer) *Screen {
	return &Screen{api: api, theme: theme, images: images}
}

func (d *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	return d.load(ctx.Context(), false)
}

var _ reactea.Component = (*Screen)(nil)
