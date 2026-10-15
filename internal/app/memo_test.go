package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/layout"
	"github.com/Hayao0819/reactea/v2/modal"
)

type renderCounter struct {
	reactea.BasicComponent
	renders int
}

func (c *renderCounter) Render(*reactea.Ctx) string {
	c.renders++

	return "background"
}

func TestRootMemoizesBackgroundWhileModalCapturesInput(t *testing.T) {
	root := newRoot(&fakeAPI{})
	stack := modal.New(root)
	program := reactea.New(stack, reactea.WithSize(80, 24))
	program.Start()

	counter := &renderCounter{}
	root.body.SetItems(layout.Grow(1, counter).Focusable())
	program.View()
	program.Send(stack.Push(reactea.Text("dialog"))())
	program.View()
	coveredRenders := counter.renders

	program.Send(tea.KeyPressMsg{Code: tea.KeyBackspace})
	program.View()
	program.View()
	if counter.renders != coveredRenders {
		t.Fatalf("covered background rendered %d more time(s)", counter.renders-coveredRenders)
	}

	program.Send(struct{}{})
	program.View()
	if counter.renders != coveredRenders+1 {
		t.Fatalf("background renders after update = %d, want %d", counter.renders, coveredRenders+1)
	}
}
