package listview

func Move(selected, count, delta int) int {
	if count <= 0 {
		return 0
	}

	return min(max(0, selected+delta), count-1)
}

func EnsureVisible(top, selected, count, room int, height func(int) int) int {
	if count <= 0 {
		return 0
	}
	selected = min(max(0, selected), count-1)
	top = min(max(0, top), selected)
	if selected < top {
		return selected
	}
	for top < selected {
		used := 0
		for index := top; index <= selected; index++ {
			used += max(0, height(index))
		}
		if used <= room {
			break
		}
		top++
	}

	return top
}

func ItemAt(top, count, row, room int, height func(int) int) (int, bool) {
	if row < 0 || row >= room {
		return 0, false
	}
	offset := 0
	for index := max(0, top); index < count; index++ {
		itemHeight := max(0, height(index))
		if row >= offset && row < offset+itemHeight {
			return index, true
		}
		offset += itemHeight
	}

	return 0, false
}
