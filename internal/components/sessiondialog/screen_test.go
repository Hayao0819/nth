package sessiondialog

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestScreenRetriesUntilBrowserSessionIsAvailable(t *testing.T) {
	t.Parallel()

	calls := 0
	request := domain.NewRefreshRequest("Firefox · default", func(context.Context) error {
		calls++
		if calls == 1 {
			return errors.New("cookies are still expired")
		}

		return nil
	})
	stack := modal.New(reactea.Text("base"))
	program := reactea.New(stack, reactea.WithSize(80, 20))
	program.Start()
	program.Send(stack.Push(New(ui.NewTheme(), request))())

	if plain := testkit.Plain(program); !strings.Contains(plain, "Browser session expired") || !strings.Contains(plain, "Firefox · default") {
		t.Fatalf("refresh screen is incomplete:\n%s", plain)
	}
	testkit.SendKeys(program, "enter")
	if plain := testkit.Plain(program); !strings.Contains(plain, "cookies are still expired") {
		t.Fatalf("refresh failure is missing:\n%s", plain)
	}
	testkit.SendKeys(program, "enter")
	if plain := testkit.Plain(program); !strings.Contains(plain, "base") || strings.Contains(plain, "Browser session expired") {
		t.Fatalf("refresh screen did not close:\n%s", plain)
	}
	if err := request.Wait(context.Background()); err != nil {
		t.Fatalf("refresh result = %v", err)
	}
}
