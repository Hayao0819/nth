package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/navigation"
)

func (s *Members) move(delta, room int) {
	if len(s.users) == 0 {
		return
	}
	s.selected = listview.Move(s.selected, len(s.users), delta)
	s.top = listview.EnsureVisible(s.top, s.selected, len(s.users), room, func(int) int { return itemHeight })
}

func (s *Members) open() tea.Cmd {
	if s.selected < 0 || s.selected >= len(s.users) {
		return nil
	}

	return navigation.OpenUser(s.users[s.selected])
}

func (s *Members) loadNearEnd(ctx context.Context) tea.Cmd {
	if len(s.users)-s.selected > 3 {
		return nil
	}

	return s.load(ctx, true)
}
