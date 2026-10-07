package list

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Members struct {
	reactea.BasicComponent

	api      domain.ListMemberAPI
	theme    ui.Theme
	list     north.List
	users    []north.User
	next     *string
	selected int
	top      int
	loading  bool
	more     bool
	acting   bool
	err      error
	notice   string
	removing bool
}

func NewMembers(api domain.ListMemberAPI, theme ui.Theme, list north.List) *Members {
	return &Members{api: api, theme: theme, list: list}
}

func (s *Members) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.load(ctx.Context(), false)
}

var _ reactea.Component = (*Members)(nil)
