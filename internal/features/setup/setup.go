// Package setup implements nth's first-run setup.
package setup

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const (
	stepMethod = iota
	stepCredential
	stepReview
)

type wizard struct {
	reactea.BasicComponent

	theme       ui.Theme
	method      auth.Method
	savedMethod auth.Method
	token       textinput.Model
	oauth       auth.OAuthCredential
	prompt      auth.OAuthPrompt
	profiles    []auth.Profile
	selected    int
	hasBrowser  bool
	envToken    bool
	images      bool
	step        int
	problem     string
	complete    bool
	oauthBusy   bool
	attempt     uint64
	oauthCtx    context.Context
	cancel      context.CancelFunc
	startOAuth  auth.OAuthStartFunc
	save        func(auth.Settings) error
}

func newWizard(
	initial auth.Settings,
	profiles []auth.Profile,
	startOAuth auth.OAuthStartFunc,
	save func(auth.Settings) error,
) *wizard {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Paste your north API token"
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'
	styles := input.Styles()
	styles.Cursor.Blink = false
	input.SetStyles(styles)
	if !initial.HasEnvironmentToken() {
		input.SetValue(initial.Token)
	}

	method := initial.Method
	if initial.HasEnvironmentToken() {
		method = auth.MethodAPIToken
	} else if method == auth.MethodBrowser || !method.Valid() {
		switch {
		case initial.HasOAuth():
			method = auth.MethodOAuth
		case initial.HasAPIToken():
			method = auth.MethodAPIToken
		default:
			method = auth.MethodOAuth
		}
	}
	if method != auth.MethodOAuth && method != auth.MethodAPIToken {
		method = auth.MethodOAuth
	}
	profiles = append([]auth.Profile(nil), profiles...)
	selected := selectedProfile(profiles, initial.Browser)
	if selected < 0 && initial.Browser.Valid() {
		profiles = append(profiles, initial.Browser)
		selected = len(profiles) - 1
	}
	if selected < 0 {
		selected = 0
	}

	return &wizard{
		theme:       ui.NewTheme(),
		method:      method,
		savedMethod: initial.Method,
		token:       input,
		oauth:       initial.OAuth,
		profiles:    profiles,
		selected:    selected,
		hasBrowser:  initial.Browser.Valid(),
		envToken:    initial.HasEnvironmentToken(),
		images:      initial.Images,
		startOAuth:  startOAuth,
		save:        save,
	}
}

func selectedProfile(profiles []auth.Profile, selected auth.Profile) int {
	if !selected.Valid() {
		return -1
	}
	for index, profile := range profiles {
		if profile.Same(selected) {
			return index
		}
	}
	for index, profile := range profiles {
		if strings.EqualFold(profile.Browser, selected.Browser) && profile.Name == selected.Name {
			return index
		}
	}

	return -1
}

func (w *wizard) Init(*reactea.Ctx) tea.Cmd {
	return reactea.SetMouseMode(tea.MouseModeCellMotion)
}

func (w *wizard) moveMethod(by int) {
	methods := [...]auth.Method{auth.MethodOAuth, auth.MethodAPIToken}
	index := 0
	for candidate, method := range methods {
		if method == w.method {
			index = candidate
			break
		}
	}
	w.method = methods[(index+by+len(methods))%len(methods)]
}

func (w *wizard) oauthReady() bool {
	return w.oauth.Valid() && !w.oauth.NeedsAuthorization()
}

func (w *wizard) cycleBrowser() {
	if len(w.profiles) == 0 {
		return
	}
	if !w.hasBrowser {
		w.hasBrowser = true

		return
	}
	w.moveProfile(1)
}

func (w *wizard) moveProfile(by int) {
	if len(w.profiles) == 0 {
		return
	}
	w.selected = (w.selected + by + len(w.profiles)) % len(w.profiles)
}

// Run displays the setup wizard and reports whether its settings were saved.
func Run(
	ctx context.Context,
	initial auth.Settings,
	profiles []auth.Profile,
	startOAuth auth.OAuthStartFunc,
	save func(auth.Settings) error,
) (bool, error) {
	wizard := newWizard(initial, profiles, startOAuth, save)
	defer wizard.cancelOAuth()
	program := reactea.New(
		wizard,
		reactea.WithAltScreen(),
		reactea.WithWindowTitle("nth setup"),
	)
	if err := program.Run(tea.WithContext(ctx)); err != nil {
		return false, fmt.Errorf("run initial setup: %w", err)
	}

	return wizard.complete, nil
}
