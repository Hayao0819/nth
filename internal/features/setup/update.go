package setup

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/reactea/v2"
)

func (w *wizard) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if reactea.Key(msg, "ctrl+c") {
		w.cancelOAuth()

		return tea.Quit
	}
	if command, handled := w.handleOAuthResult(msg); handled {
		return command
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		if command, handled := w.handleClick(ctx, msg); handled {
			return command
		}
	}
	w.problem = ""

	switch w.step {
	case stepMethod:
		return w.updateMethod(ctx, msg)
	case stepCredential:
		return w.updateCredential(ctx, msg)
	case stepReview:
		return w.updateReview(ctx, msg)
	default:
		return nil
	}
}

func (w *wizard) handleOAuthResult(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case oauthStartedMsg:
		if msg.attempt != w.attempt {
			return nil, true
		}
		if msg.err != nil {
			w.stopOAuth()
			w.problem = msg.err.Error()

			return nil, true
		}
		if msg.session == nil {
			w.stopOAuth()
			w.problem = "North sign-in did not return a session"

			return nil, true
		}
		w.prompt = msg.session.Prompt()

		return w.waitOAuth(msg.attempt, msg.session), true
	case oauthCompletedMsg:
		if msg.attempt != w.attempt {
			return nil, true
		}
		w.stopOAuth()
		if msg.err != nil {
			w.problem = msg.err.Error()

			return nil, true
		}
		if !msg.credential.Valid() {
			w.prompt = auth.OAuthPrompt{}
			w.problem = "North returned an invalid sign-in"

			return nil, true
		}
		if missing := msg.credential.MissingScopes(); len(missing) > 0 {
			w.prompt = auth.OAuthPrompt{}
			w.problem = "North did not grant: " + strings.Join(missing, ", ")

			return nil, true
		}
		w.oauth = msg.credential
		w.step = stepReview

		return nil, true
	default:
		return nil, false
	}
}

func (w *wizard) updateMethod(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
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
			if w.oauthReady() {
				w.step = stepReview

				return nil
			}
			w.step = stepCredential

			return w.beginOAuth(ctx.Context())
		}
	}

	return nil
}

func (w *wizard) updateCredential(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
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

	return nil
}

func (w *wizard) updateReview(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
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
		if w.method == auth.MethodOAuth && !w.oauthReady() {
			w.step = stepCredential

			return w.beginOAuth(ctx.Context())
		}
		return w.saveSettings()
	case reactea.Key(msg, "i"):
		w.images = !w.images
	case reactea.Key(msg, "b"):
		w.cycleBrowser()
	}

	return nil
}

func (w *wizard) saveSettings() tea.Cmd {
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
}
