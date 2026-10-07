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
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type FolderDetail struct {
	reactea.BasicComponent

	theme  ui.Theme
	folder bookmarkdomain.BookmarkFolder
	feed   *feed.Feed
	api    bookmarkdomain.BookmarkFolderAPI
	busy   bool
	notice string
}

type removedMsg struct {
	target   *FolderDetail
	postID   string
	removed  bool
	response *north.Response
	err      error
}

func (m removedMsg) Response() *north.Response { return m.response }

func NewFolderDetail(api feed.API, folders bookmarkdomain.BookmarkFolderAPI, theme ui.Theme, folder bookmarkdomain.BookmarkFolder, images *termimage.Renderer) *FolderDetail {
	loader := func(ctx context.Context, cursor string) (north.PostPage, *north.Response, error) {
		page, response, err := folders.BookmarkFolderPosts(ctx, folder.ID, cursor)

		return north.PostPage{Items: page.Items, NextCursor: page.NextCursor}, response, err
	}

	return &FolderDetail{
		theme:  theme,
		folder: folder,
		feed:   feed.NewSourceWithImages(api, theme, "bookmark folder", loader, images),
		api:    folders,
	}
}

func (s *FolderDetail) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.feed.Init(s.feedCtx(ctx))
}

func (s *FolderDetail) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if result, ok := msg.(removedMsg); ok {
		if result.target != s {
			return nil
		}
		s.busy = false
		if result.err != nil {
			s.notice = ui.FriendlyError(result.err)
		} else if !result.removed {
			s.notice = "Post was not removed"
		} else {
			s.feed.RemovePost(s.feedCtx(ctx), result.postID)
			s.notice = "Removed from folder"
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
			return navigation.OpenPost(*post)
		}
	case reactea.Key(msg, "u"):
		if post := s.feed.SelectedPost(); post != nil && post.DisplayPost() != nil {
			return navigation.OpenUser(post.DisplayPost().Author)
		}
	case reactea.Key(msg, "r", "R"):
		return s.action(postcomponent.Reply)
	case reactea.Key(msg, "Q"):
		return s.action(postcomponent.Quote)
	case reactea.Key(msg, "b"):
		return s.action(postcomponent.Bookmark)
	case reactea.Key(msg, "x"):
		return s.remove(ctx.Context())
	}

	return s.updateFeed(ctx, msg)
}

func (s *FolderDetail) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	title := ui.SafeInline(s.folder.Name)
	if title == "" {
		title = "Bookmark folder"
	}
	right := ""
	if s.busy {
		right = s.theme.Dim.Render("Removing…")
	} else if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-len(title)-8)))
	}
	header := pageheader.Render(s.theme, title, right, width)
	body := s.feed.Render(s.feedCtx(ctx))
	footer := "j/k move · Enter open · u profile · b unbookmark · x remove from folder · Esc/← back"
	if progress := s.feed.Progress(); progress != "" {
		footer = ui.Sides(footer, progress, width)
	}

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *FolderDetail) action(action postcomponent.Action) tea.Cmd {
	if post := s.feed.SelectedPost(); post != nil {
		return postcomponent.Request(action, *post)
	}

	return nil
}

func (s *FolderDetail) remove(ctx context.Context) tea.Cmd {
	if s.api == nil || s.busy {
		return nil
	}
	post := s.feed.SelectedPost()
	if post == nil || post.DisplayPost() == nil || post.DisplayPost().ID == "" {
		return nil
	}
	postID := post.DisplayPost().ID
	s.busy = true
	s.notice = ""

	return func() tea.Msg {
		removed, response, err := s.api.RemoveBookmarkFromFolder(ctx, s.folder.ID, postID)

		return removedMsg{target: s, postID: postID, removed: removed, response: response, err: err}
	}
}

func (s *FolderDetail) feedCtx(ctx *reactea.Ctx) *reactea.Ctx {
	return ctx.Inset(0, pageheader.Height, ctx.Width(), max(0, ctx.Height()-pageheader.Height-1))
}

func (s *FolderDetail) updateFeed(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	feedCtx := s.feedCtx(ctx)
	if reactea.IsMouse(msg) {
		parentX, parentY := ctx.Origin()
		feedX, feedY := feedCtx.Origin()
		msg = reactea.TranslateMouse(msg, feedX-parentX, feedY-parentY)
	}

	return s.feed.Update(feedCtx, msg)
}

var _ reactea.Component = (*FolderDetail)(nil)
