package account

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/navigation"
)

func (s *Screen) move(delta, room int) {
	if s.count() == 0 {
		return
	}
	s.selected = listview.Move(s.selected, s.count(), delta)
	s.top = listview.EnsureVisible(s.top, s.selected, s.count(), room, func(int) int { return accountItemHeight })
}

func (s *Screen) openUser() tea.Cmd {
	if s.tab == keywordsTab || s.selected < 0 || s.selected >= len(s.users) {
		return nil
	}

	return navigation.OpenUser(s.users[s.selected])
}

func (s *Screen) setTab(next tab) {
	maximum := mutedTab
	if s.keywords != nil {
		maximum = keywordsTab
	}
	if next > maximum {
		next = requestsTab
	}
	s.tab = next
	s.users, s.words, s.next = nil, nil, nil
	s.selected, s.top = 0, 0
	s.loading = false
	s.err = nil
	s.notice = ""
}

func (s *Screen) loadNearEnd(ctx context.Context) tea.Cmd {
	if s.count()-s.selected > 3 {
		return nil
	}

	return s.load(ctx, true)
}
