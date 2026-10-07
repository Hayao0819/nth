package setup

import (
	"strings"

	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/nth/internal/ui"
)

func (w *wizard) methodScreen(compact bool, width int) []string {
	needsAuthorization := w.method == auth.MethodOAuth && w.oauth.NeedsAuthorization()
	title := "Choose how to sign in"
	if needsAuthorization {
		title = "Update north authorization"
	}
	lines := []string{w.theme.Heading.Render(title), w.currentAuthentication(), ""}
	if needsAuthorization {
		lines = append(lines, w.theme.Warn.Render("Additional permissions are required."))
		if !compact {
			message := "This version of nth needs permissions that are not in the saved sign-in."
			if len(w.oauth.Scopes) == 0 {
				message = "The saved sign-in predates permission tracking and must be renewed once."
			}
			lines = append(lines, w.theme.Dim.Render(message))
			if len(w.oauth.Scopes) > 0 {
				missing := strings.Join(w.oauth.MissingScopes(), ", ")
				lines = append(lines, w.theme.Dim.Render("NEW  ")+w.theme.Warn.Render(missing))
			}
		}
		lines = append(lines, "")
	} else if !compact {
		instruction := "Select a sign-in method. Nothing changes until you save."
		if w.envToken {
			instruction = "NORTH_API_KEY is active and always takes priority."
		}
		lines = append(lines, w.theme.Dim.Render(instruction), "")
	}
	oauthDetail := "Authorize nth in your browser · recommended"
	if w.oauth.NeedsAuthorization() {
		oauthDetail = "Review the additional permissions in north"
	} else if w.oauth.Valid() {
		oauthDetail = "Signed in · token refreshes automatically"
	}
	lines = append(lines, w.authChoice(w.method == auth.MethodOAuth, auth.MethodOAuth, "Sign in with north", oauthDetail, compact, width)...)
	lines = append(lines, "")
	tokenDetail := "Paste a developer token · advanced"
	if w.envToken {
		tokenDetail = "NORTH_API_KEY · active"
	} else if strings.TrimSpace(w.token.Value()) != "" {
		tokenDetail = "Saved in the system keyring"
	}
	lines = append(lines, w.authChoice(w.method == auth.MethodAPIToken, auth.MethodAPIToken, "API token", tokenDetail, compact, width)...)
	if !compact {
		lines = append(lines, "", w.theme.Dim.Render("Optional browser cookies are configured on the review step."))
	}

	return lines
}

func (w *wizard) currentAuthentication() string {
	label := "Not configured"
	style := w.theme.Active
	switch {
	case w.envToken:
		label = "API token · NORTH_API_KEY"
	case w.savedMethod == auth.MethodOAuth && w.oauth.Valid():
		label = "OAuth · north account"
		if w.oauth.NeedsAuthorization() {
			label = "OAuth · authorization update required"
			style = w.theme.Warn
		}
	case w.savedMethod == auth.MethodAPIToken && strings.TrimSpace(w.token.Value()) != "":
		label = "API token · keyring"
	case w.savedMethod == auth.MethodBrowser && w.hasBrowser:
		label = "Browser session · legacy"
	}

	return w.theme.Dim.Render("CURRENT  ") + style.Render(label)
}

func (w *wizard) authChoice(selected bool, method auth.Method, title, detail string, compact bool, width int) []string {
	marker := "  "
	style := w.theme.Heading
	if selected {
		marker = w.theme.Active.Render("▌ ")
		style = w.theme.Active
	}
	badge := w.authBadge(method)
	if badge != "" {
		if method == auth.MethodOAuth && w.oauth.NeedsAuthorization() {
			badge = w.theme.Warn.Render(badge)
		} else {
			badge = w.theme.Active.Render(badge)
		}
	}
	lines := []string{ui.Sides(marker+style.Render(title), badge, width)}
	if !compact {
		lines = append(lines, "  "+w.theme.Dim.Render(detail))
	}

	return lines
}

func (w *wizard) authBadge(method auth.Method) string {
	switch method {
	case auth.MethodOAuth:
		if w.oauth.NeedsAuthorization() {
			return "UPDATE"
		}
		if w.oauth.Valid() {
			if w.savedMethod == method {
				return "SAVED"
			}

			return "READY"
		}
	case auth.MethodAPIToken:
		if w.envToken {
			return "ACTIVE"
		}
		if strings.TrimSpace(w.token.Value()) != "" {
			if w.savedMethod == method {
				return "SAVED"
			}

			return "READY"
		}
	}

	return ""
}
