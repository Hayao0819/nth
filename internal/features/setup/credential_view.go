package setup

import (
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/ui"
)

func (w *wizard) tokenScreen(compact bool) []string {
	lines := []string{
		w.theme.Heading.Render("Enter an API token"),
		w.theme.Dim.Render("The token will be stored in the system keyring."),
		"",
	}
	if !compact {
		lines = append(lines, w.theme.Dim.Render("Create a read/write token in north Settings › Developer."), "")
	}

	return append(lines, w.theme.Dim.Render("TOKEN"), w.theme.Accent.Render("› ")+w.token.View())
}

func (w *wizard) oauthScreen(compact bool, width int) []string {
	lines := []string{w.theme.Heading.Render("Sign in with north")}
	if w.prompt.VerificationURI == "" {
		if w.oauthBusy {
			return append(lines, "", w.theme.Dim.Render("Requesting an authorization code…"))
		}

		return append(lines, "", w.theme.Dim.Render("Sign-in stopped. Press enter to try again."))
	}

	address := w.prompt.VerificationURIComplete
	if address == "" {
		address = w.prompt.VerificationURI
	}
	lines = append(lines,
		w.theme.Active.Render("● Waiting for approval"),
		"",
		w.theme.Dim.Render("1  Open this URL in your browser"),
		"   "+w.theme.Accent.Render(ui.Clip(ui.SafeInline(address), max(1, width-3))),
	)
	if w.prompt.UserCode != "" {
		lines = append(lines,
			"",
			w.theme.Dim.Render("2  Enter this code if north asks for it"),
			lipgloss.PlaceHorizontal(width, lipgloss.Center, w.theme.Brand.Render(ui.SafeInline(w.prompt.UserCode))),
		)
	}
	if !compact {
		lines = append(lines,
			"",
			w.theme.Dim.Render("Approve nth in the browser. Setup continues automatically."),
			w.theme.Dim.Render("Your password is entered only on north."),
		)
	}

	return lines
}
