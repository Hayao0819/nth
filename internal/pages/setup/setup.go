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
	"github.com/charmbracelet/x/ansi"
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

func (w *wizard) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if reactea.Key(msg, "ctrl+c") {
		w.cancelOAuth()

		return tea.Quit
	}
	switch msg := msg.(type) {
	case oauthStartedMsg:
		if msg.attempt != w.attempt {
			return nil
		}
		if msg.err != nil {
			w.stopOAuth()
			w.problem = msg.err.Error()

			return nil
		}
		if msg.session == nil {
			w.stopOAuth()
			w.problem = "North sign-in did not return a session"

			return nil
		}
		w.prompt = msg.session.Prompt()

		return w.waitOAuth(msg.attempt, msg.session)
	case oauthCompletedMsg:
		if msg.attempt != w.attempt {
			return nil
		}
		w.stopOAuth()
		if msg.err != nil {
			w.problem = msg.err.Error()

			return nil
		}
		w.oauth = msg.credential
		w.step = stepReview

		return nil
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		if command, handled := w.handleClick(ctx, msg); handled {
			return command
		}
	}
	w.problem = ""

	switch w.step {
	case stepMethod:
		switch {
		case reactea.Key(msg, "up", "k", "left", "h"):
			w.moveMethod(-1)
		case reactea.Key(msg, "down", "j", "right", "l", "space"):
			w.moveMethod(1)
		case reactea.Key(msg, "r") && w.method == auth.MethodOAuth && w.oauth.Valid():
			w.step = stepCredential
			return w.beginOAuth(ctx.Context())
		case reactea.Key(msg, "enter"):
			switch w.method {
			case auth.MethodAPIToken:
				if w.envToken {
					w.step = stepReview

					return nil
				}
				w.step = stepCredential
				return w.token.Focus()
			case auth.MethodOAuth:
				if w.oauth.Valid() {
					w.step = stepReview

					return nil
				}
				w.step = stepCredential
				return w.beginOAuth(ctx.Context())
			}
		}

	case stepCredential:
		switch w.method {
		case auth.MethodAPIToken:
			switch {
			case reactea.Key(msg, "esc"):
				w.token.Blur()
				w.step = stepMethod
				return nil
			case reactea.Key(msg, "enter"):
				if strings.TrimSpace(w.token.Value()) == "" {
					w.problem = "Enter a north API token"

					return nil
				}
				w.token.Blur()
				w.step = stepReview
				return nil
			}
			var command tea.Cmd
			w.token, command = w.token.Update(msg)

			return command
		case auth.MethodOAuth:
			switch {
			case reactea.Key(msg, "esc"):
				w.cancelOAuth()
				w.step = stepMethod
			case reactea.Key(msg, "enter") && !w.oauthBusy:
				return w.beginOAuth(ctx.Context())
			}
		}

	case stepReview:
		switch {
		case reactea.Key(msg, "esc"):
			if w.method == auth.MethodOAuth || w.envToken {
				w.step = stepMethod
			} else {
				w.step = stepCredential
			}
			if w.method == auth.MethodAPIToken {
				return w.token.Focus()
			}
		case reactea.Key(msg, "enter"):
			if w.save == nil {
				w.problem = "Settings cannot be saved"

				return nil
			}
			settings := auth.Settings{
				Method: w.method,
				Token:  w.token.Value(),
				OAuth:  w.oauth,
				Images: w.images,
			}
			if len(w.profiles) > 0 && (w.method == auth.MethodBrowser || w.hasBrowser) {
				settings.Browser = w.profiles[w.selected]
			}
			if w.method == auth.MethodBrowser && !settings.Browser.Valid() {
				w.problem = "No supported browser profile was found"

				return nil
			}
			if err := w.save(settings); err != nil {
				w.problem = err.Error()

				return nil
			}
			w.complete = true

			return tea.Quit
		case reactea.Key(msg, "i"):
			w.images = !w.images
		case reactea.Key(msg, "b"):
			w.cycleBrowser()
		}
	}

	return nil
}

func (w *wizard) handleClick(ctx *reactea.Ctx, msg tea.Msg) (tea.Cmd, bool) {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil, false
	}
	lines := strings.Split(ansi.Strip(w.Render(ctx)), "\n")
	if y < 0 || y >= len(lines) {
		return nil, false
	}
	line := lines[y]
	w.problem = ""

	switch w.step {
	case stepMethod:
		if ui.TextAt(line, x, "Sign in with north") {
			w.method = auth.MethodOAuth

			return nil, true
		}
		if ui.TextAt(line, x, "API token") {
			w.method = auth.MethodAPIToken

			return nil, true
		}
	case stepCredential:
		if w.method == auth.MethodAPIToken && strings.Contains(line, "› ") {
			return w.token.Focus(), true
		}
	case stepReview:
		if ui.TextAt(line, x, "Browser cookies") {
			w.cycleBrowser()

			return nil, true
		}
		if ui.TextAt(line, x, "Terminal images") {
			w.images = !w.images

			return nil, true
		}
	}

	if ui.TextAt(line, x, "enter continue") || ui.TextAt(line, x, "enter review") {
		return w.Update(ctx, tea.KeyPressMsg{Code: tea.KeyEnter}), true
	}
	if ui.TextAt(line, x, "enter save") {
		return w.Update(ctx, tea.KeyPressMsg{Code: tea.KeyEnter}), true
	}
	if ui.TextAt(line, x, "r sign in again") {
		return w.Update(ctx, tea.KeyPressMsg{Code: 'r'}), true
	}
	if ui.TextAt(line, x, "b browser") {
		return w.Update(ctx, tea.KeyPressMsg{Code: 'b'}), true
	}
	if ui.TextAt(line, x, "i images") {
		return w.Update(ctx, tea.KeyPressMsg{Code: 'i'}), true
	}
	if ui.TextAt(line, x, "esc back") || ui.TextAt(line, x, "esc cancel") {
		return w.Update(ctx, tea.KeyPressMsg{Code: tea.KeyEscape}), true
	}

	return nil, false
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
