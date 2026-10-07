package post

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/navigation"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
)

func (s *Activity) itemCount() int {
	switch s.kind {
	case navigation.PostQuotes:
		return len(s.posts)
	case navigation.PostHistory:
		return len(s.versions)
	default:
		return len(s.users)
	}
}

func (s *Activity) move(delta, width, room int) {
	if s.itemCount() == 0 {
		return
	}
	s.selected = listview.Move(s.selected, s.itemCount(), delta)
	s.ensureVisible(width, room)
}

func (s *Activity) ensureVisible(width, room int) {
	s.top = listview.EnsureVisible(s.top, s.selected, s.itemCount(), room, func(index int) int {
		return lipgloss.Height(s.renderItem(index, width, index == s.selected))
	})
}

func (s *Activity) loadNearEnd(ctx context.Context) tea.Cmd {
	if s.itemCount()-s.selected > 3 {
		return nil
	}

	return s.load(ctx, true)
}

func (s *Activity) open() tea.Cmd {
	if s.selected < 0 || s.selected >= s.itemCount() {
		return nil
	}
	switch s.kind {
	case navigation.PostQuotes:
		return navigation.OpenPost(s.posts[s.selected])
	case navigation.PostHistory:
		return nil
	default:
		return navigation.OpenUser(s.users[s.selected])
	}
}

func (s *Activity) openUser() tea.Cmd {
	if s.selected < 0 || s.selected >= s.itemCount() {
		return nil
	}
	if s.kind == navigation.PostQuotes {
		target := s.posts[s.selected].DisplayPost()
		if target == nil {
			return nil
		}

		return navigation.OpenUser(target.Author)
	}
	if s.kind == navigation.PostHistory {
		return nil
	}

	return navigation.OpenUser(s.users[s.selected])
}

func (s *Activity) action(action postcomponent.Action) tea.Cmd {
	if s.kind != navigation.PostQuotes || s.selected < 0 || s.selected >= len(s.posts) {
		return nil
	}

	return postcomponent.Request(action, s.posts[s.selected])
}
