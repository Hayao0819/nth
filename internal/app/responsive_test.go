package app

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestRootFitsTerminalBoundaries(t *testing.T) {
	t.Parallel()

	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(1, 1))
	program.Start()

	widths := []int{1, 2, 5, 30, 47, 48, 87, 88, 151, 152, 220}
	heights := []int{1, 2, 5, 10, 15, 19, 20, 30}
	for _, width := range widths {
		for _, height := range heights {
			view := testkit.RenderAt(program, width, height)
			gotWidth, gotHeight := lipgloss.Size(view)
			if gotWidth != width || gotHeight != height {
				t.Errorf("%dx%d terminal rendered as %dx%d", width, height, gotWidth, gotHeight)
			}
		}
	}
}
