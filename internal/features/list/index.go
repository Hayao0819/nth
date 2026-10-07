package list

import (
	tea "charm.land/bubbletea/v2"
	listdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type entry struct {
	section string
	item    listdomain.ListItem
}

type Index struct {
	reactea.BasicComponent

	api      listdomain.ListAPI
	editor   listdomain.ListEditorAPI
	theme    ui.Theme
	entries  []entry
	selected int
	top      int
	loading  bool
	err      error
	notice   string
	acting   bool
}

func NewIndex(api listdomain.ListAPI, theme ui.Theme) *Index {
	screen := &Index{api: api, theme: theme}
	if supportsListEditing(api) {
		screen.editor, _ = api.(listdomain.ListEditorAPI)
	}

	return screen
}

func supportsListEditing(api any) bool {
	reporter, ok := api.(interface{ SupportsListEditing() bool })

	return !ok || reporter.SupportsListEditing()
}

func (s *Index) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.load(ctx.Context())
}

var _ reactea.Component = (*Index)(nil)
