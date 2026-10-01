package listview

import "testing"

func TestEnsureVisibleUsesVariableItemHeights(t *testing.T) {
	t.Parallel()

	heights := []int{2, 3, 2, 4}
	height := func(index int) int { return heights[index] }
	if got := EnsureVisible(0, 2, len(heights), 5, height); got != 1 {
		t.Fatalf("EnsureVisible() = %d, want 1", got)
	}
	if got := EnsureVisible(2, 1, len(heights), 5, height); got != 1 {
		t.Fatalf("EnsureVisible() above viewport = %d, want 1", got)
	}
}

func TestItemAtUsesRenderedRows(t *testing.T) {
	t.Parallel()

	heights := []int{2, 3, 2}
	height := func(index int) int { return heights[index] }
	for row, want := range []int{1, 1, 1, 2, 2} {
		got, ok := ItemAt(1, len(heights), row, 5, height)
		if !ok || got != want {
			t.Fatalf("ItemAt(row %d) = %d, %v; want %d, true", row, got, ok, want)
		}
	}
	if _, ok := ItemAt(1, len(heights), 5, 5, height); ok {
		t.Fatal("ItemAt accepted a row outside the viewport")
	}
}

func TestMoveClampsSelection(t *testing.T) {
	t.Parallel()

	if got := Move(1, 3, 8); got != 2 {
		t.Fatalf("Move down = %d, want 2", got)
	}
	if got := Move(1, 3, -8); got != 0 {
		t.Fatalf("Move up = %d, want 0", got)
	}
}
