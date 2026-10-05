package setup

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (w *wizard) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	if width <= 0 || height <= 0 {
		return ""
	}
	if width < 42 || height < 14 {
		message := strings.Join([]string{
			"nth setup",
			"",
			fmt.Sprintf("terminal is %d×%d", width, height),
			"resize to at least 42×14",
			"",
			"ctrl+c exit",
		}, "\n")

		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
	}

	panelStyle := w.theme.Box.Padding(1, 2)
	panelWidth := min(width-4, 74)
	panelHeight := min(height-2, 23)
	contentWidth := panelWidth - panelStyle.GetHorizontalFrameSize()
	contentHeight := panelHeight - panelStyle.GetVerticalFrameSize()
	compact := contentWidth < 58 || contentHeight < 17
	w.token.SetWidth(max(1, contentWidth-2))

	body := w.screen(compact, contentWidth)
	if w.problem != "" {
		body = append(body, "", w.theme.Bad.Render("! "+ui.SafeInline(w.problem)))
	}
	bodyRoom := max(1, contentHeight-4)
	if len(body) > bodyRoom {
		body = body[:bodyRoom]
	}
	for len(body) < bodyRoom {
		body = append(body, "")
	}

	header := ui.Sides(
		w.theme.Brand.Render("nth setup"),
		w.theme.Dim.Render(fmt.Sprintf("%d / %d", w.step+1, stepReview+1)),
		contentWidth,
	)
	lines := []string{header, w.progress(contentWidth), ""}
	lines = append(lines, body...)
	lines = append(lines, w.theme.Dim.Render(w.footer(contentWidth)))
	for index, line := range lines {
		lines[index] = ui.Clip(line, contentWidth)
	}

	panel := panelStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, panel)
}

func (w *wizard) screen(compact bool, width int) []string {
	switch w.step {
	case stepMethod:
		return w.methodScreen(compact, width)
	case stepCredential:
		if w.method == auth.MethodOAuth {
			return w.oauthScreen(compact, width)
		}

		return w.tokenScreen(compact)
	case stepReview:
		return w.review(compact, width)
	default:
		return nil
	}
}

func (w *wizard) progress(width int) string {
	labels := []string{"1  Account", "2  Verify", "3  Finish"}
	for index, label := range labels {
		name := strings.TrimPrefix(label, fmt.Sprintf("%d  ", index+1))
		switch {
		case index < w.step:
			labels[index] = w.theme.Active.Render("✓  " + name)
		case index == w.step:
			labels[index] = w.theme.Heading.Render("●  " + name)
		default:
			labels[index] = w.theme.Dim.Render("○  " + name)
		}
	}

	return ui.Columns(labels, width)
}

func (w *wizard) methodScreen(compact bool, width int) []string {
	lines := []string{
		w.theme.Heading.Render("Authentication"),
		w.currentAuthentication(),
		"",
	}
	if !compact {
		lines = append(lines, w.theme.Dim.Render("Choose the credential used for public API requests."), "")
	}
	lines = append(lines, w.authChoice(
		w.method == auth.MethodOAuth,
		auth.MethodOAuth,
		"Sign in with north",
		"Browser code · refreshes automatically",
		compact,
		width,
	)...)
	lines = append(lines, "")
	lines = append(lines, w.authChoice(
		w.method == auth.MethodAPIToken,
		auth.MethodAPIToken,
		"API token",
		"Paste a token · advanced",
		compact,
		width,
	)...)
	if !compact {
		lines = append(lines, "", w.theme.Dim.Render("Browser features are optional and configured after sign-in."))
	}

	return lines
}

func (w *wizard) currentAuthentication() string {
	label := "Not configured"
	switch {
	case w.envToken:
		label = "API token · NORTH_API_KEY"
	case w.savedMethod == auth.MethodOAuth && w.oauth.Valid():
		label = "OAuth · north account"
	case w.savedMethod == auth.MethodAPIToken && strings.TrimSpace(w.token.Value()) != "":
		label = "API token · keyring"
	case w.savedMethod == auth.MethodBrowser && w.hasBrowser:
		label = "Browser session · legacy"
	}

	return w.theme.Dim.Render("CURRENT  ") + w.theme.Active.Render(label)
}

func (w *wizard) authChoice(
	selected bool,
	method auth.Method,
	title string,
	detail string,
	compact bool,
	width int,
) []string {
	marker := "  "
	style := w.theme.Heading
	if selected {
		marker = w.theme.Active.Render("▌ ")
		style = w.theme.Active
	}
	badge := ""
	if w.currentMethod() == method {
		badge = w.theme.Dim.Render("CURRENT")
	}
	lines := []string{ui.Sides(marker+style.Render(title), badge, width)}
	if !compact {
		lines = append(lines, "  "+w.theme.Dim.Render(detail))
	}

	return lines
}

func (w *wizard) currentMethod() auth.Method {
	if w.envToken {
		return auth.MethodAPIToken
	}

	return w.savedMethod
}

func (w *wizard) tokenScreen(compact bool) []string {
	lines := []string{
		w.theme.Heading.Render("API token"),
		w.theme.Dim.Render("Stored in the system keyring."),
		"",
	}
	if !compact {
		lines = append(lines, w.theme.Dim.Render("Create a read/write token in north Settings › Developer."), "")
	}

	return append(lines, w.theme.Accent.Render("› ")+w.token.View())
}

func (w *wizard) oauthScreen(compact bool, width int) []string {
	lines := []string{w.theme.Heading.Render("Link your north account")}
	if w.prompt.VerificationURI == "" {
		if w.oauthBusy {
			return append(lines, "", w.theme.Dim.Render("Requesting a browser code…"))
		}

		return append(lines, "", w.theme.Dim.Render("Press enter to try again."))
	}

	address := w.prompt.VerificationURIComplete
	if address == "" {
		address = w.prompt.VerificationURI
	}
	lines = append(lines,
		w.theme.Dim.Render("1  Open this address"),
		"   "+w.theme.Accent.Render(ui.Clip(ui.SafeInline(address), max(1, width-3))),
		"",
		w.theme.Dim.Render("2  Enter this code"),
		lipgloss.PlaceHorizontal(width, lipgloss.Center, w.theme.Brand.Render(ui.SafeInline(w.prompt.UserCode))),
	)
	if !compact {
		lines = append(lines,
			"",
			w.theme.Dim.Render("Waiting for north to confirm this device…"),
			w.theme.Dim.Render("nth never receives your password."),
		)
	}

	return lines
}

func (w *wizard) review(compact bool, width int) []string {
	title, detail := w.selectedAuthentication()
	lines := []string{
		w.theme.Heading.Render("Ready to start"),
		w.theme.Dim.Render("ACTIVE AUTHENTICATION"),
		w.theme.Active.Render("▌ " + title),
		"  " + w.theme.Dim.Render(detail),
		"",
		w.theme.Dim.Render("OPTIONAL FEATURES"),
		w.settingLine("Browser features", w.browserState(), w.browserKey(), width),
		w.settingLine("Terminal images", imageState(w.images), "i toggle", width),
	}
	if !compact {
		lines = append(lines,
			"",
			w.theme.Dim.Render("Browser cookies are only used when the public API cannot provide a feature."),
		)
	}

	return lines
}

func (w *wizard) selectedAuthentication() (string, string) {
	if w.envToken {
		return "API token", "NORTH_API_KEY · always preferred"
	}
	if w.method == auth.MethodOAuth {
		return "OAuth", "Device authorization · refreshable"
	}

	return "API token", "System keyring"
}

func (w *wizard) browserState() string {
	if w.hasBrowser && len(w.profiles) > 0 {
		return ui.SafeInline(w.profiles[w.selected].Label())
	}
	if len(w.profiles) == 0 {
		return "Unavailable"
	}

	return "Not configured"
}

func (w *wizard) browserKey() string {
	if len(w.profiles) == 0 {
		return ""
	}

	return "b change"
}

func imageState(enabled bool) string {
	if enabled {
		return "Enabled"
	}

	return "Disabled"
}

func (w *wizard) settingLine(label, value, key string, width int) string {
	right := value
	if key != "" {
		right += "  " + key
	}

	return ui.Sides(w.theme.Heading.Render(label), w.theme.Dim.Render(right), width)
}

func (w *wizard) footer(width int) string {
	footer := "↑/↓ choose   enter continue   ctrl+c exit"
	switch w.step {
	case stepCredential:
		if w.method == auth.MethodOAuth {
			footer = "esc cancel   ctrl+c exit"
			if !w.oauthBusy {
				footer = "enter retry   esc back   ctrl+c exit"
			}
		} else {
			footer = "enter continue   esc back   ctrl+c exit"
		}
	case stepReview:
		footer = "b browser   i images   enter save and start   esc back"
	}
	if width < 58 {
		switch w.step {
		case stepReview:
			footer = "b browser   i images   enter save"
		case stepCredential:
			footer = "enter continue   esc back"
		default:
			footer = "↑/↓ choose   enter continue"
		}
	}

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, footer)
}
