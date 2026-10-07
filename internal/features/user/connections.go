package user

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Connections struct {
	reactea.BasicComponent

	api       domain.UserConnectionsAPI
	theme     ui.Theme
	user      north.User
	following bool
	items     []north.User
	next      *string
	selected  int
	top       int
	loading   bool
	more      bool
	err       error
}

func NewConnections(api domain.UserConnectionsAPI, theme ui.Theme, user north.User, following bool) *Connections {
	return &Connections{api: api, theme: theme, user: user, following: following}
}

func (s *Connections) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.load(ctx.Context(), false)
}

func (s *Connections) room(height int) int {
	return max(0, height-pageheader.Height-1)
}

var _ reactea.Component = (*Connections)(nil)
