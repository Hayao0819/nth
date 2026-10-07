package user

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Connections) move(delta, width, room int) {
	if len(s.items) == 0 {
		return
	}
	s.selected = listview.Move(s.selected, len(s.items), delta)
	s.ensureVisible(width, room)
}

func (s *Connections) ensureVisible(width, room int) {
	s.top = listview.EnsureVisible(s.top, s.selected, len(s.items), room, func(index int) int {
		return lipgloss.Height(s.renderUser(s.items[index], width, index == s.selected))
	})
}

func (s *Connections) open() tea.Cmd {
	if s.selected < 0 || s.selected >= len(s.items) {
		return nil
	}

	return navigation.OpenUser(s.items[s.selected])
}

func (s *Connections) click(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside || msg.Button != tea.MouseLeft {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	row := y - pageheader.Height
	index, _, ok := listview.ItemPositionAt(s.top, len(s.items), row, s.room(ctx.Height()), func(index int) int {
		return lipgloss.Height(s.renderUser(s.items[index], ctx.Width(), index == s.selected))
	})
	if !ok {
		return nil
	}
	s.selected = index
	s.ensureVisible(ctx.Width(), s.room(ctx.Height()))

	return s.open()
}
