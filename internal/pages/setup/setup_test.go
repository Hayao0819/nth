package setup

import (
	"errors"
	"strings"
	"testing"

	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func setupProgram(wizard *wizard, width, height int) *reactea.App {
	program := reactea.New(wizard, reactea.WithSize(width, height))
	program.Start()

	return program
}

func TestWizardCollectsBrowser(t *testing.T) {
	t.Parallel()

	profiles := []auth.Profile{
		{Browser: "chrome", Name: "Default", Default: true},
		{Browser: "firefox", Name: "default-release", Default: true},
	}
	var saved auth.Settings
	wizard := newWizard(auth.Settings{}, profiles, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 80, 24)

	initial := testkit.Plain(program)
	if !strings.Contains(initial, "Authentication") || !strings.Contains(initial, "Browser session") || !strings.Contains(initial, "API token") {
		t.Fatalf("method step is incomplete:\n%s", initial)
	}
	testkit.SendKeys(program, "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "Google Chrome · Default") || !strings.Contains(view, "Firefox · default-release") {
		t.Fatalf("browser choices are incomplete:\n%s", view)
	}
	testkit.SendKeys(program, "down", "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "Ready to start") || !strings.Contains(view, "Firefox · default-release") {
		t.Fatalf("review is incomplete:\n%s", view)
	}
	testkit.SendKeys(program, "i")
	testkit.SendKeys(program, "enter")

	if !wizard.complete {
		t.Fatal("setup did not complete")
	}
	if saved.Method != auth.MethodBrowser || saved.Browser.Browser != "firefox" || !saved.Images {
		t.Fatalf("saved settings = %#v", saved)
	}
}

func TestWizardCollectsAPIToken(t *testing.T) {
	t.Parallel()

	var saved auth.Settings
	wizard := newWizard(auth.Settings{}, nil, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 72, 20)
	testkit.SendKeys(program, "down", "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "North API token") {
		t.Fatalf("token step is incomplete:\n%s", view)
	}
	testkit.SendKeys(program, "s", "e", "c", "r", "e", "t")
	if view := testkit.Plain(program); strings.Contains(view, "secret") {
		t.Fatalf("API token is visible:\n%s", view)
	}
	testkit.SendKeys(program, "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "Ready to start") || !strings.Contains(view, "Configured · preferred") {
		t.Fatalf("review is incomplete:\n%s", view)
	}
	testkit.SendKeys(program, "enter")
	if !wizard.complete || saved.Method != auth.MethodAPIToken || saved.Token != "secret" {
		t.Fatalf("saved settings = %#v", saved)
	}
}

func TestWizardKeepsReviewOpenAfterKeyringFailure(t *testing.T) {
	t.Parallel()

	wizard := newWizard(
		auth.Settings{},
		[]auth.Profile{{Browser: "firefox"}},
		func(auth.Settings) error { return errors.New("keyring is locked") },
	)
	program := setupProgram(wizard, 72, 20)
	testkit.SendKeys(program, "enter", "enter", "enter")

	if wizard.complete {
		t.Fatal("setup completed despite a keyring error")
	}
	if view := testkit.Plain(program); !strings.Contains(view, "keyring is locked") {
		t.Fatalf("save error is missing:\n%s", view)
	}
}

func TestWizardFollowsAMovedCookieStore(t *testing.T) {
	t.Parallel()

	initial := auth.Settings{
		Method:  auth.MethodBrowser,
		Browser: auth.Profile{Browser: "chrome", Name: "Default", Path: "/old/Cookies"},
	}
	current := auth.Profile{Browser: "chrome", Name: "Default", Path: "/new/Network/Cookies"}
	wizard := newWizard(initial, []auth.Profile{current}, nil)
	if len(wizard.profiles) != 1 || !wizard.profiles[wizard.selected].Same(current) {
		t.Fatalf("selected profile = %#v", wizard.profiles[wizard.selected])
	}
}

func TestWizardExplainsSmallTerminal(t *testing.T) {
	t.Parallel()

	wizard := newWizard(auth.Settings{}, nil, nil)
	program := setupProgram(wizard, 30, 8)
	if view := testkit.Plain(program); !strings.Contains(view, "resize to at least 38×10") {
		t.Fatalf("small-terminal message is missing:\n%s", view)
	}
}

func TestWizardCanBeCompletedWithTheMouse(t *testing.T) {
	t.Parallel()

	var saved auth.Settings
	wizard := newWizard(auth.Settings{}, nil, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 72, 20)

	clickSetupText(t, program, "API token")
	clickSetupText(t, program, "enter continue")
	testkit.SendKeys(program, "s", "e", "c", "r", "e", "t")
	clickSetupText(t, program, "enter continue")
	clickSetupText(t, program, "enter save and start")

	if !wizard.complete || saved.Method != auth.MethodAPIToken || saved.Token != "secret" {
		t.Fatalf("saved settings = %#v", saved)
	}
}

func clickSetupText(t *testing.T, program *reactea.App, label string) {
	t.Helper()
	for y, line := range testkit.Lines(program) {
		if x := strings.Index(line, label); x >= 0 {
			testkit.Click(program, x, y)

			return
		}
	}
	t.Fatalf("could not find %q in setup:\n%s", label, testkit.Plain(program))
}
