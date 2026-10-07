package account

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type tab uint8

const (
	requestsTab tab = iota
	blockedTab
	mutedTab
	keywordsTab
)

type changeKind uint8

const (
	changeNone changeKind = iota
	changeAccept
	changeReject
	changeUnblock
	changeUnmute
	changeCreateKeyword
	changeDeleteKeyword
)

type Screen struct {
	reactea.BasicComponent

	api      domain.AccountSafetyAPI
	keywords domain.MutedKeywordAPI
	theme    ui.Theme
	tab      tab
	users    []north.User
	words    []north.MutedKeyword
	next     *string
	selected int
	top      int
	loading  bool
	more     bool
	acting   bool
	err      error
	notice   string
	confirm  changeKind
}

func New(api domain.AccountSafetyAPI, keywords domain.MutedKeywordAPI, theme ui.Theme) *Screen {
	return &Screen{api: api, keywords: keywords, theme: theme}
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.load(ctx.Context(), false)
}

var _ reactea.Component = (*Screen)(nil)
