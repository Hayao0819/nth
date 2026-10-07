package bookmark

import (
	tea "charm.land/bubbletea/v2"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type FolderIndex struct {
	reactea.BasicComponent

	api        bookmarkdomain.BookmarkFolderAPI
	theme      ui.Theme
	folders    []bookmarkdomain.BookmarkFolder
	selected   int
	top        int
	loading    bool
	acting     bool
	editingID  string
	deletingID string
	err        error
	notice     string
}

func NewFolderIndex(api bookmarkdomain.BookmarkFolderAPI, theme ui.Theme) *FolderIndex {
	return &FolderIndex{api: api, theme: theme}
}

func (s *FolderIndex) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.load(ctx.Context())
}

var _ reactea.Component = (*FolderIndex)(nil)
