package saved

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type tab uint8

const (
	draftsTab tab = iota
	scheduledTab
	pendingTab
)

type editorKind uint8

const (
	editorNone editorKind = iota
	editorDraft
	editorScheduled
	editorPending
)

type Screen struct {
	reactea.BasicComponent

	api       domain.SavedPostAPI
	theme     ui.Theme
	tab       tab
	drafts    []north.Draft
	scheduled []north.ScheduledPost
	pending   []north.PendingPost
	selected  int
	top       int
	loading   bool
	acting    bool
	err       error
	notice    string
	editor    editorKind
	editingID string
	confirm   bool
	deleteID  string
	deleteTab tab
}

func New(api domain.SavedPostAPI, theme ui.Theme) *Screen {
	return &Screen{api: api, theme: theme}
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.load(ctx.Context())
}

var _ reactea.Component = (*Screen)(nil)
