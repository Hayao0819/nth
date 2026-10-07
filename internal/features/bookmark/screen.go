package bookmark

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type Screen struct {
	reactea.BasicComponent

	theme    ui.Theme
	feed     *feed.Feed
	folders  bookmarkdomain.BookmarkFolderAPI
	clearing bool
	notice   string
}

func (s *Screen) SetFolderAPI(api bookmarkdomain.BookmarkFolderAPI) {
	s.folders = api
}

func New(api feed.API, bookmarks bookmarkdomain.BookmarkAPI, theme ui.Theme, images *termimage.Renderer) *Screen {
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
	if result, ok := msg.(modal.Result[dialog.Confirmation]); ok && s.clearing {
		s.clearing = false
		if result.Ok() && result.Value.Accepted {
			return s.clear(ctx.Context())
		}

		return nil
	}
	if cleared, ok := msg.(clearedMsg); ok {
		if cleared.target != s {
			return nil
		}
		s.clearing = false
		if cleared.err != nil {
			s.notice = ui.FriendlyError(cleared.err)
		} else {
			s.feed.Clear()
			s.notice = "Bookmarks cleared"
		}

		return nil
	}
	if update, ok := msg.(postcomponent.ReactionUpdate); ok && update.Bookmark != nil && update.Err == nil && !*update.Bookmark {
		s.feed.RemovePost(s.feedCtx(ctx), update.PostID)
	}
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
	case reactea.Key(msg, "b"):
		return s.action(postcomponent.Bookmark)
	case s.folders != nil && reactea.Key(msg, "f"):
		post := s.feed.SelectedPost()
		if post == nil || post.DisplayPost() == nil {
			return nil
		}

		return modal.PushAt(
			ctx,
			newFolderPicker(s.folders, s.theme, post.DisplayPost().ID),
			dialog.Placement(ctx, 56, 16),
		)
	case s.folders != nil && reactea.Key(msg, "F"):
		return navigation.OpenBookmarkFolders()
	case s.folders != nil && reactea.Key(msg, "x"):
		s.clearing = true

		return modal.PushAt(
			ctx,
			dialog.NewConfirm(s.theme, "Clear bookmarks?", "Every bookmark will be removed. This cannot be undone.", "Clear"),
			dialog.Placement(ctx, 52, 8),
		)
	}

	return s.updateFeed(ctx, msg)
}

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	right := ""
	if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-12)))
	}
	header := pageheader.Render(s.theme, "Bookmarks", right, width)
	body := s.feed.Render(s.feedCtx(ctx))
	footer := "j/k move · Enter open · u profile · Esc/← back"
	if s.folders != nil {
		footer += " · f add to folder · F folders · x clear"
	}
	if progress := s.feed.Progress(); progress != "" {
		footer = ui.Sides(footer, progress, width)
	}

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

type clearedMsg struct {
	target   *Screen
	response *north.Response
	err      error
}

func (m clearedMsg) Response() *north.Response { return m.response }

func (s *Screen) clear(ctx context.Context) tea.Cmd {
	if s.folders == nil {
		return nil
	}
	s.clearing = true
	s.notice = "Clearing bookmarks…"

	return func() tea.Msg {
		_, response, err := s.folders.ClearBookmarks(ctx)

		return clearedMsg{target: s, response: response, err: err}
	}
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

func (s *Screen) updateFeed(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	feedCtx := s.feedCtx(ctx)
	if reactea.IsMouse(msg) {
		parentX, parentY := ctx.Origin()
		feedX, feedY := feedCtx.Origin()
		msg = reactea.TranslateMouse(msg, feedX-parentX, feedY-parentY)
	}

	return s.feed.Update(feedCtx, msg)
}

var _ reactea.Component = (*Screen)(nil)
