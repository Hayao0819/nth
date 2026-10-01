package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestQuitGuardHandlesControlCAboveAModal(t *testing.T) {
	t.Parallel()

	guard := &quitGuard{Wrapper: reactea.Wrap(modal.New(dialog.NewHelp(ui.NewTheme(), true, true, true)))}
	program := reactea.New(guard, reactea.WithSize(60, 16))
	command := guard.Update(program.Ctx(), testkit.Key("ctrl+c"))
	if command == nil {
		t.Fatal("ctrl+c did not return a command")
	}
	if _, ok := command().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c did not quit while a modal was active")
	}
}
