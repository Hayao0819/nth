package post

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Activity) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case loadedMsg:
		s.applyLoaded(msg)

		return s.loadImages(ctx.Context(), ctx.Width(), msg.posts)
	case postcomponent.ReactionUpdate:
		if msg.Err == nil {
			s.applyReaction(msg)
		}

		return nil
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			s.move(1, ctx.Width(), s.room(ctx.Height()))

			return s.loadNearEnd(ctx.Context())
		}
		if msg.Button == tea.MouseWheelUp {
			s.move(-1, ctx.Width(), s.room(ctx.Height()))
		}

		return nil
	case tea.MouseClickMsg:
		return s.click(ctx, msg)
	case tea.WindowSizeMsg:
		return s.loadImages(ctx.Context(), ctx.Width(), s.posts)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "j", "down"):
		s.move(1, ctx.Width(), s.room(ctx.Height()))

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "k", "up"):
		s.move(-1, ctx.Width(), s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.selected, s.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if s.itemCount() > 0 {
			s.selected = s.itemCount() - 1
			s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
		}

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "enter"):
		return s.open()
	case reactea.Key(msg, "u"):
		return s.openUser()
	case reactea.Key(msg, "r"):
		return s.action(postcomponent.Reply)
	case reactea.Key(msg, "t"):
		return s.action(postcomponent.Repost)
	case reactea.Key(msg, "l", "f"):
		return s.action(postcomponent.Like)
	case reactea.Key(msg, "Q"):
		return s.action(postcomponent.Quote)
	case reactea.Key(msg, "b"):
		return s.action(postcomponent.Bookmark)
	case reactea.Key(msg, "L"):
		return s.load(ctx.Context(), true)
	case reactea.Key(msg, ".") && s.err != nil:
		return s.load(ctx.Context(), false)
	}

	return nil
}

func (s *Activity) click(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside || msg.Button != tea.MouseLeft {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	row := y - pageheader.Height
	index, itemRow, ok := listview.ItemPositionAt(s.top, s.itemCount(), row, s.room(ctx.Height()), func(index int) int {
		return lipgloss.Height(s.renderItem(index, ctx.Width(), index == s.selected))
	})
	if !ok {
		return nil
	}
	s.selected = index
	s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
	if s.kind == navigation.PostQuotes {
		post := s.posts[index]
		height := lipgloss.Height(s.renderItem(index, ctx.Width(), true))
		if itemRow == height-2 {
			return postcomponent.Request(postcomponent.CardActionAt(x, ctx.Width()), post)
		}
		if user, hit := postcomponent.CardUserAtWithImages(post, x, itemRow, ctx.Width(), s.images); hit {
			return navigation.OpenUser(user)
		}
		if quoted, hit := postcomponent.CardQuotedPostAtWithImages(post, x, itemRow, ctx.Width(), s.images); hit {
			return navigation.OpenPost(quoted)
		}
	}

	return s.open()
}

func (s *Activity) applyReaction(update postcomponent.ReactionUpdate) {
	for index := range s.posts {
		target := s.posts[index].DisplayPost()
		if target == nil || target.ID != update.PostID {
			continue
		}
		if update.Like != nil {
			target.Liked, target.LikeCount = update.Like.Liked, update.Like.LikeCount
		}
		if update.Repost != nil {
			target.Reposted, target.RepostCount = update.Repost.Reposted, update.Repost.RepostCount
		}
		if update.Bookmark != nil {
			target.Bookmarked = *update.Bookmark
		}
	}
}
