package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/termimage"
	listdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Detail struct {
	reactea.BasicComponent

	api      listdomain.ListAPI
	editor   listdomain.ListEditorAPI
	members  listdomain.ListMemberAPI
	theme    ui.Theme
	item     listdomain.ListItem
	feed     *feed.Feed
	loading  bool
	acting   bool
	err      error
	notice   string
	deleting bool
}

func NewDetail(api feed.API, lists listdomain.ListAPI, theme ui.Theme, item listdomain.ListItem, images *termimage.Renderer) *Detail {
	loader := func(ctx context.Context, cursor string) (north.PostPage, *north.Response, error) {
		return lists.ListTimeline(ctx, item.ID, cursor)
	}

	screen := &Detail{
		api:   lists,
		theme: theme,
		item:  item,
		feed:  feed.NewSourceWithImages(api, theme, "list", loader, images),
	}
	if supportsListEditing(lists) {
		screen.editor, _ = lists.(listdomain.ListEditorAPI)
	}
	if supportsListMembers(lists) {
		screen.members, _ = lists.(listdomain.ListMemberAPI)
	}

	return screen
}

func supportsListMembers(api any) bool {
	reporter, ok := api.(interface{ SupportsListMembers() bool })

	return !ok || reporter.SupportsListMembers()
}

func (s *Detail) Init(ctx *reactea.Ctx) tea.Cmd {
	commands := []tea.Cmd{s.feed.Init(s.feedCtx(ctx))}
	if s.item.Name == "" {
		commands = append(commands, s.load(ctx.Context()))
	}

	return tea.Batch(commands...)
}

var _ reactea.Component = (*Detail)(nil)
