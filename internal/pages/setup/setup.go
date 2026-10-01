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

	theme      ui.Theme
	method     auth.Method
	token      textinput.Model
	profiles   []auth.Profile
	selected   int
	hasBrowser bool
	images     bool
	step       int
	problem    string
	complete   bool
	save       func(auth.Settings) error
}

func newWizard(initial auth.Settings, profiles []auth.Profile, save func(auth.Settings) error) *wizard {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Paste your north API token"
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'
	styles := input.Styles()
	styles.Cursor.Blink = false
	input.SetStyles(styles)
	input.SetValue(initial.Token)

	method := initial.Method
	if !method.Valid() {
		method = auth.MethodBrowser
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
		theme:      ui.NewTheme(),
		method:     method,
		token:      input,
		profiles:   profiles,
		selected:   selected,
		hasBrowser: initial.Browser.Valid(),
		images:     initial.Images,
		save:       save,
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
		return tea.Quit
	}
	if wheel, ok := msg.(tea.MouseWheelMsg); ok && w.step == stepCredential && w.method == auth.MethodBrowser {
		if _, _, inside := reactea.Mouse(ctx, msg); inside {
			if wheel.Button == tea.MouseWheelUp {
				w.moveProfile(-1)
			} else if wheel.Button == tea.MouseWheelDown {
				w.moveProfile(1)
			}

			return nil
		}
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
		case reactea.Key(msg, "up", "k", "left", "h", "down", "j", "right", "l", "space"):
			w.toggleMethod()
		case reactea.Key(msg, "enter"):
			w.step = stepCredential
			if w.method == auth.MethodAPIToken {
				return w.token.Focus()
			}
		}

	case stepCredential:
		if w.method == auth.MethodAPIToken {
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
		}

		switch {
		case reactea.Key(msg, "esc"):
			w.step = stepMethod
		case reactea.Key(msg, "up", "k", "left", "h"):
			w.moveProfile(-1)
		case reactea.Key(msg, "down", "j", "right", "l", "space"):
			w.moveProfile(1)
		case reactea.Key(msg, "enter"):
			if len(w.profiles) == 0 {
				w.problem = "No supported browser profile was found"

				return nil
			}
			w.hasBrowser = true
			w.step = stepReview
		}

	case stepReview:
		switch {
		case reactea.Key(msg, "esc"):
			w.step = stepCredential
			if w.method == auth.MethodAPIToken {
				return w.token.Focus()
			}
		case reactea.Key(msg, "enter"):
			if w.save == nil {
				w.problem = "Settings cannot be saved"

				return nil
			}
			settings := auth.Settings{Method: w.method, Token: w.token.Value(), Images: w.images}
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
		if ui.TextAt(line, x, "Browser session") {
			w.method = auth.MethodBrowser

			return nil, true
		}
		if ui.TextAt(line, x, "API token") {
			w.method = auth.MethodAPIToken

			return nil, true
		}
	case stepCredential:
		if w.method == auth.MethodBrowser {
			for index, profile := range w.profiles {
				if ui.TextAt(line, x, ui.SafeInline(profile.Label())) {
					w.selected = index

					return nil, true
				}
			}
		} else if strings.Contains(line, "› ") {
			return w.token.Focus(), true
		}
	case stepReview:
		if ui.TextAt(line, x, "Terminal images") {
			w.images = !w.images

			return nil, true
		}
	}

	if ui.TextAt(line, x, "enter continue") {
		return w.Update(ctx, tea.KeyPressMsg{Code: tea.KeyEnter}), true
	}
	if ui.TextAt(line, x, "enter save and start") || ui.TextAt(line, x, "enter save") {
		return w.Update(ctx, tea.KeyPressMsg{Code: tea.KeyEnter}), true
	}
	if ui.TextAt(line, x, "esc back") {
		return w.Update(ctx, tea.KeyPressMsg{Code: tea.KeyEscape}), true
	}

	return nil, false
}

func (w *wizard) toggleMethod() {
	if w.method == auth.MethodBrowser {
		w.method = auth.MethodAPIToken
	} else {
		w.method = auth.MethodBrowser
	}
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
	save func(auth.Settings) error,
) (bool, error) {
	wizard := newWizard(initial, profiles, save)
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
