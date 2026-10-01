package feed

import (
	"time"

	"charm.land/lipgloss/v2"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
)

func (f *Feed) move(delta, width, height int) {
	if len(f.posts) == 0 {
		return
	}
	f.selected = min(max(f.selected+delta, 0), len(f.posts)-1)
	f.ensureVisible(width, height)
}

func (f *Feed) selectPost(index, width, height int) {
	f.selected = index
	f.ensureVisible(width, height)
	if index > 0 {
		f.refreshReady = true
	}
}

func (f *Feed) moveByHeight(direction, distance, width, height int) {
	if len(f.posts) == 0 || direction == 0 {
		return
	}
	direction = max(-1, min(1, direction))
	distance = max(1, distance)
	target := f.selected
	used := 0
	for used < distance {
		next := target + direction
		if next < 0 || next >= len(f.posts) {
			break
		}
		target = next
		used += lipgloss.Height(postcomponent.RenderCardWithImages(f.posts[target], width, target == f.selected, f.theme, time.Now(), f.images))
	}
	f.selected = target
	f.ensureVisible(width, height)
}

func (f *Feed) postAt(y, width, height int) (index, row int, ok bool) {
	if y < 0 {
		return 0, 0, false
	}
	offset := 0
	for index := f.top; index < len(f.posts); index++ {
		cardHeight := lipgloss.Height(postcomponent.RenderCardWithImages(f.posts[index], width, index == f.selected, f.theme, time.Now(), f.images))
		if y >= offset && y < offset+cardHeight {
			return index, y - offset, true
		}
		offset += cardHeight
		if offset > height {
			return 0, 0, false
		}
	}

	return 0, 0, false
}

func (f *Feed) clampSelection() {
	if len(f.posts) == 0 {
		f.selected, f.top = 0, 0

		return
	}
	f.selected = min(max(f.selected, 0), len(f.posts)-1)
	f.top = min(max(f.top, 0), f.selected)
}

func (f *Feed) ensureVisible(width, height int) {
	f.clampSelection()
	if len(f.posts) == 0 || width <= 0 || height <= 0 {
		return
	}
	if f.selected < f.top {
		f.top = f.selected
	}

	for f.top < f.selected && !f.selectionFits(f.top, width, height) {
		f.top++
	}
	for f.top > 0 && f.selectionFits(f.top-1, width, height) {
		f.top--
	}
}

func (f *Feed) selectionFits(top, width, height int) bool {
	row := f.heightBetween(top, f.selected-1, width) + postcomponent.CardCursorRow(f.posts[f.selected])
	if row > height/2 {
		return false
	}

	return top == f.selected || f.heightBetween(top, f.selected, width) <= height
}

func (f *Feed) heightBetween(first, last, width int) int {
	height := 0
	for index := first; index <= last && index < len(f.posts); index++ {
		height += lipgloss.Height(postcomponent.RenderCardWithImages(f.posts[index], width, index == f.selected, f.theme, time.Now(), f.images))
	}

	return height
}

func (f *Feed) renderedHeight(width int) int {
	return f.heightBetween(f.top, len(f.posts)-1, width)
}
