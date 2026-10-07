package setup

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/oauth2"
)

func setupProgram(wizard *wizard, width, height int) *reactea.App {
	program := reactea.New(wizard, reactea.WithSize(width, height))
	program.Start()

	return program
}

func TestWizardConfiguresBrowserFeatures(t *testing.T) {
	t.Parallel()

	profiles := []auth.Profile{
		{Browser: "chrome", Name: "Default", Default: true},
		{Browser: "firefox", Name: "default-release", Default: true},
	}
	var saved auth.Settings
	wizard := newWizard(auth.Settings{Method: auth.MethodAPIToken, Token: "saved-token"}, profiles, nil, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 80, 24)

	initial := testkit.Plain(program)
	if !strings.Contains(initial, "Choose how to sign in") || !strings.Contains(initial, "Sign in with north") || !strings.Contains(initial, "API token") || strings.Contains(initial, "Browser session") {
		t.Fatalf("method step is incomplete:\n%s", initial)
	}
	testkit.SendKeys(program, "enter", "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "Review and save") || !strings.Contains(view, "Browser cookies") || !strings.Contains(view, "Not configured") {
		t.Fatalf("review is incomplete:\n%s", view)
	}
	testkit.SendKeys(program, "b", "b")
	if view := testkit.Plain(program); !strings.Contains(view, "Firefox · default-release") {
		t.Fatalf("selected browser is missing:\n%s", view)
	}
	testkit.SendKeys(program, "i")
	testkit.SendKeys(program, "enter")

	if !wizard.complete {
		t.Fatal("setup did not complete")
	}
	if saved.Method != auth.MethodAPIToken || saved.Token != "saved-token" || saved.Browser.Browser != "firefox" || !saved.Images {
		t.Fatalf("saved settings = %#v", saved)
	}
}

func TestWizardSignsInWithOAuth(t *testing.T) {
	t.Parallel()

	var saved auth.Settings
	credential := auth.OAuthCredential{
		ClientID: "client-id",
		Token:    oauth2.Token{AccessToken: "access-token", RefreshToken: "refresh-token"},
		Scopes:   auth.RequiredOAuthScopes(),
	}
	start := func(context.Context) (auth.OAuthSession, error) {
		return staticOAuthSession{credential: credential}, nil
	}
	wizard := newWizard(auth.Settings{}, nil, start, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 72, 20)
	testkit.SendKeys(program, "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "Review and save") || !strings.Contains(view, "SIGN-IN") || !strings.Contains(view, "north account") || !strings.Contains(view, "Signed in · token refreshes automatically") {
		t.Fatalf("OAuth review is incomplete:\n%s", view)
	}
	testkit.SendKeys(program, "enter")
	if !wizard.complete || saved.Method != auth.MethodOAuth || !saved.OAuth.Valid() || saved.OAuth.Token.RefreshToken != "refresh-token" {
		t.Fatalf("saved settings = %#v", saved)
	}
}

func TestWizardReusesCompletedOAuthAfterGoingBack(t *testing.T) {
	t.Parallel()

	credential := auth.OAuthCredential{
		ClientID: "client-id",
		Token:    oauth2.Token{AccessToken: "access-token", RefreshToken: "refresh-token"},
		Scopes:   auth.RequiredOAuthScopes(),
	}
	starts := 0
	start := func(context.Context) (auth.OAuthSession, error) {
		starts++

		return staticOAuthSession{credential: credential}, nil
	}
	wizard := newWizard(auth.Settings{}, nil, start, func(auth.Settings) error { return nil })
	program := setupProgram(wizard, 72, 20)

	testkit.SendKeys(program, "enter")
	if starts != 1 || !wizard.oauth.Valid() || wizard.step != stepReview {
		t.Fatalf("initial sign-in: starts=%d step=%d credential=%#v", starts, wizard.step, wizard.oauth)
	}
	testkit.SendKeys(program, "esc")
	if view := testkit.Plain(program); wizard.step != stepMethod || !strings.Contains(view, "READY") || !strings.Contains(view, "r sign in again") {
		t.Fatalf("completed sign-in was not preserved:\n%s", view)
	}
	testkit.SendKeys(program, "enter")
	if starts != 1 || wizard.step != stepReview || !wizard.oauth.Valid() {
		t.Fatalf("sign-in restarted: starts=%d step=%d credential=%#v", starts, wizard.step, wizard.oauth)
	}
}

func TestWizardExplainsAndRenewsOutdatedOAuth(t *testing.T) {
	t.Parallel()

	required := auth.RequiredOAuthScopes()
	old := auth.OAuthCredential{
		ClientID: "client-id",
		Token:    oauth2.Token{AccessToken: "old-access", RefreshToken: "old-refresh"},
		Scopes:   append([]string(nil), required[:len(required)-1]...),
	}
	current := auth.OAuthCredential{
		ClientID: "client-id",
		Token:    oauth2.Token{AccessToken: "new-access", RefreshToken: "new-refresh"},
		Scopes:   required,
	}
	starts := 0
	start := func(context.Context) (auth.OAuthSession, error) {
		starts++

		return staticOAuthSession{credential: current}, nil
	}
	var saved auth.Settings
	wizard := newWizard(auth.Settings{Method: auth.MethodOAuth, OAuth: old}, nil, start, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 72, 24)

	view := testkit.Plain(program)
	missing := required[len(required)-1]
	for _, want := range []string{
		"Update north authorization",
		"Additional permissions are required",
		missing,
		"enter authorize",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("authorization update is missing %q:\n%s", want, view)
		}
	}
	testkit.SendKeys(program, "enter")
	if starts != 1 || wizard.step != stepReview || !wizard.oauthReady() {
		t.Fatalf("authorization was not renewed: starts=%d step=%d oauth=%#v", starts, wizard.step, wizard.oauth)
	}
	testkit.SendKeys(program, "enter")
	if !wizard.complete || saved.OAuth.NeedsAuthorization() || saved.OAuth.Token.AccessToken != "new-access" {
		t.Fatalf("saved OAuth = %#v", saved.OAuth)
	}
}

func TestWizardCollectsAPIToken(t *testing.T) {
	t.Parallel()

	var saved auth.Settings
	wizard := newWizard(auth.Settings{}, nil, nil, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 72, 20)
	testkit.SendKeys(program, "down", "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "API token") {
		t.Fatalf("token step is incomplete:\n%s", view)
	}
	testkit.SendKeys(program, "s", "e", "c", "r", "e", "t")
	if view := testkit.Plain(program); strings.Contains(view, "secret") {
		t.Fatalf("API token is visible:\n%s", view)
	}
	testkit.SendKeys(program, "enter")
	if view := testkit.Plain(program); !strings.Contains(view, "Review and save") || !strings.Contains(view, "SIGN-IN") || !strings.Contains(view, "system keyring") {
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
		auth.Settings{Method: auth.MethodAPIToken, Token: "saved-token"},
		[]auth.Profile{{Browser: "firefox"}},
		nil,
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
	wizard := newWizard(initial, []auth.Profile{current}, nil, nil)
	if len(wizard.profiles) != 1 || !wizard.profiles[wizard.selected].Same(current) {
		t.Fatalf("selected profile = %#v", wizard.profiles[wizard.selected])
	}
}

func TestWizardExplainsSmallTerminal(t *testing.T) {
	t.Parallel()

	wizard := newWizard(auth.Settings{}, nil, nil, nil)
	program := setupProgram(wizard, 30, 8)
	if view := testkit.Plain(program); !strings.Contains(view, "resize to at least 42×14") {
		t.Fatalf("small-terminal message is missing:\n%s", view)
	}
}

func TestWizardKeepsPanelAlignedAcrossSteps(t *testing.T) {
	t.Parallel()

	wizard := newWizard(auth.Settings{}, nil, nil, nil)
	program := setupProgram(wizard, 88, 26)
	methodView := testkit.Lines(program)
	methodX := lineTextX(methodView, "Choose how to sign in")

	testkit.SendKeys(program, "down", "enter")
	tokenView := testkit.Lines(program)
	tokenX := lineTextX(tokenView, "Enter an API token")
	if methodX < 0 || tokenX != methodX {
		t.Fatalf("content columns differ: method=%d token=%d\n%s", methodX, tokenX, testkit.Plain(program))
	}
	if borderX(methodView) != borderX(tokenView) {
		t.Fatalf("panel moved between steps: method=%d token=%d", borderX(methodView), borderX(tokenView))
	}
}

func TestWizardCanBeCompletedWithTheMouse(t *testing.T) {
	t.Parallel()

	var saved auth.Settings
	wizard := newWizard(auth.Settings{}, nil, nil, func(settings auth.Settings) error {
		saved = settings

		return nil
	})
	program := setupProgram(wizard, 72, 20)

	clickSetupText(t, program, "API token")
	clickSetupText(t, program, "enter continue")
	testkit.SendKeys(program, "s", "e", "c", "r", "e", "t")
	clickSetupText(t, program, "enter continue")
	clickSetupText(t, program, "enter save")

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

func lineTextX(lines []string, text string) int {
	for _, line := range lines {
		if x := strings.Index(ansi.Strip(line), text); x >= 0 {
			return x
		}
	}

	return -1
}

func borderX(lines []string) int {
	return lineTextX(lines, "╭")
}

type staticOAuthSession struct {
	credential auth.OAuthCredential
}

func (staticOAuthSession) Prompt() auth.OAuthPrompt {
	return auth.OAuthPrompt{VerificationURI: "https://north.rip/device", UserCode: "ABCD-EFGH"}
}

func (s staticOAuthSession) Wait(context.Context) (auth.OAuthCredential, error) {
	return s.credential, nil
}
