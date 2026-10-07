package saved

import "github.com/Hayao0819/nth/internal/components/listview"

func (s *Screen) move(delta, room int) {
	if s.count() == 0 {
		return
	}
	s.selected = listview.Move(s.selected, s.count(), delta)
	s.top = listview.EnsureVisible(s.top, s.selected, s.count(), room, func(int) int { return savedItemHeight })
}

func (s *Screen) setTab(next tab) {
	if next == s.tab {
		return
	}
	s.tab = next
	s.selected, s.top = 0, 0
	s.loading = false
	s.err = nil
	s.notice = ""
}
