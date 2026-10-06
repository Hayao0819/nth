package notification

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/navigation"
)

func (d *Screen) move(by, width, room int) {
	if len(d.items) == 0 {
		return
	}
	d.selected = listview.Move(d.selected, len(d.items), by)
	d.ensureVisible(width, room)
}

func (d *Screen) ensureVisible(width, room int) {
	d.top = listview.EnsureVisible(d.top, d.selected, len(d.items), room, func(index int) int {
		return lipgloss.Height(d.renderItem(d.items[index], width, index == d.selected))
	})
}

func (d *Screen) itemPositionAt(row, width, room int) (int, int, bool) {
	return listview.ItemPositionAt(d.top, len(d.items), row, room, func(index int) int {
		return lipgloss.Height(d.renderItem(d.items[index], width, index == d.selected))
	})
}

func (d *Screen) open() tea.Cmd {
	if d.selected < 0 || d.selected >= len(d.items) {
		return nil
	}
	item := d.items[d.selected]
	if item.Post != nil {
		post := *item.Post
		return navigation.OpenPost(post)
	}

	return d.openUser()
}

func (d *Screen) openUser() tea.Cmd {
	if d.selected < 0 || d.selected >= len(d.items) || len(d.items[d.selected].Actors) == 0 {
		return nil
	}
	user := d.items[d.selected].Actors[0]
	return navigation.OpenUser(user)
}
