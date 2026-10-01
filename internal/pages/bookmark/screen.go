package bookmark

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain/bookmark"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Screen struct {
	reactea.BasicComponent

	theme ui.Theme
	feed  *feed.Feed
}

func New(api feed.API, bookmarks bookmarkdomain.API, theme ui.Theme, images *termimage.Renderer) *Screen {
	loader := func(ctx context.Context, cursor string) (north.PostPage, *north.Response, error) {
		return bookmarks.Bookmarks(ctx, cursor)
	}

	return &Screen{
		theme: theme,
		feed:  feed.NewSourceWithImages(api, theme, "bookmarks", loader, images),
	}
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.feed.Init(s.feedCtx(ctx))
}

func (s *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if !inside {
			return nil
		}
		if pageheader.BackAt(x, y) {
			return pageheader.Back()
		}
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "enter"):
		if post := s.feed.SelectedPost(); post != nil {
			selected := *post

			return navigation.OpenPost(selected)
		}
	case reactea.Key(msg, "u"):
		if post := s.feed.SelectedPost(); post != nil && post.DisplayPost() != nil {
			user := post.DisplayPost().Author

			return navigation.OpenUser(user)
		}
	case reactea.Key(msg, "r", "R"):
		return s.action(postcomponent.Reply)
	case reactea.Key(msg, "Q"):
		return s.action(postcomponent.Quote)
	}

	return s.feed.Update(s.feedCtx(ctx), msg)
}

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	header := pageheader.Render(s.theme, "Bookmarks", "", width)
	body := s.feed.Render(s.feedCtx(ctx))
	footer := "j/k move · Enter open · u profile · Esc/← back"
	if progress := s.feed.Progress(); progress != "" {
		footer = ui.Sides(footer, progress, width)
	}

	return ui.Fit(header+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Screen) action(action postcomponent.Action) tea.Cmd {
	post := s.feed.SelectedPost()
	if post == nil {
		return nil
	}
	selected := *post

	return postcomponent.Request(action, selected)
}

func (s *Screen) feedCtx(ctx *reactea.Ctx) *reactea.Ctx {
	return ctx.Inset(0, pageheader.Height, ctx.Width(), max(0, ctx.Height()-pageheader.Height-1))
}

var _ reactea.Component = (*Screen)(nil)
