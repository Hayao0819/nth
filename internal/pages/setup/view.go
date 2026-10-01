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
	if width < 38 || height < 10 {
		message := strings.Join([]string{
			"nth setup",
			"",
			fmt.Sprintf("terminal is %d×%d", width, height),
			"resize to at least 38×10",
			"",
			"ctrl+c exit",
		}, "\n")

		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
	}

	inner := min(width-4, 72)
	compact := height < 16
	w.token.SetWidth(max(1, inner-2))
	header := ui.Sides(
		w.theme.Brand.Render("nth setup"),
		w.theme.Dim.Render(fmt.Sprintf("%d / %d", w.step+1, stepReview+1)),
		inner,
	)
	lines := []string{header}
	if !compact {
		lines = append(lines, "")
	}

	switch w.step {
	case stepMethod:
		lines = append(lines, w.methodScreen(compact)...)
	case stepCredential:
		if w.method == auth.MethodBrowser {
			lines = append(lines, w.browserScreen(compact, height-len(lines)-4)...)
		} else {
			lines = append(lines, w.tokenScreen(compact)...)
		}
	case stepReview:
		lines = append(lines, w.review(compact)...)
	}

	if w.problem != "" {
		if !compact {
			lines = append(lines, "")
		}
		lines = append(lines, w.theme.Bad.Render("! "+ui.SafeInline(w.problem)))
	}
	if !compact {
		lines = append(lines, "")
	}
	lines = append(lines, w.theme.Dim.Render(w.footer(inner)))

	for index, line := range lines {
		lines[index] = ui.Clip(line, inner)
	}
	if len(lines) > height {
		lines = append(lines[:height-1], lines[len(lines)-1])
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}

func (w *wizard) methodScreen(compact bool) []string {
	lines := []string{w.theme.Heading.Render("Authentication")}
	if !compact {
		lines = append(lines, w.theme.Dim.Render("Choose a credential to add or update."), "")
	}
	lines = append(lines,
		w.choice(w.method == auth.MethodBrowser, "Browser session"),
		w.choice(w.method == auth.MethodAPIToken, "API token"),
	)
	if !compact {
		description := "Enables messages, bookmarks, and other browser-only features."
		if w.method == auth.MethodAPIToken {
			description = "Preferred for every operation supported by the public API."
		}
		lines = append(lines, "", w.theme.Dim.Render(description))
	}

	return lines
}

func (w *wizard) choice(active bool, label string) string {
	marker := w.theme.Dim.Render("  ○ ")
	style := w.theme.Heading
	if active {
		marker = w.theme.Active.Render("  ● ")
		style = w.theme.Active
	}

	return marker + style.Render(label)
}

func (w *wizard) browserScreen(compact bool, room int) []string {
	lines := []string{w.theme.Heading.Render("Browser profile")}
	if !compact {
		lines = append(lines,
			w.theme.Dim.Render("Choose the profile that is signed in to north.rip."),
			w.theme.Dim.Render("nth reads its cookies again whenever it starts."),
			"",
		)
	}

	return append(lines, w.browserChoices(max(1, room-len(lines)))...)
}

func (w *wizard) tokenScreen(compact bool) []string {
	lines := []string{w.theme.Heading.Render("North API token")}
	if !compact {
		lines = append(lines,
			w.theme.Dim.Render("Create a read/write token in north Settings › Developer."),
			w.theme.Dim.Render("The token will be stored in your system keyring."),
			"",
		)
	}

	return append(lines, w.theme.Accent.Render("› ")+w.token.View())
}

func (w *wizard) footer(width int) string {
	footer := "↑/↓ choose   enter continue   ctrl+c exit"
	switch w.step {
	case stepCredential:
		if w.method == auth.MethodAPIToken {
			footer = "enter continue   esc back   ctrl+c exit"
		} else {
			footer = "↑/↓ choose   enter continue   esc back   ctrl+c exit"
		}
	case stepReview:
		footer = "i toggle images   enter save and start   esc back   ctrl+c exit"
	}
	if width < 58 {
		switch w.step {
		case stepReview:
			return "i images   enter save   esc back"
		case stepCredential:
			return "enter continue   esc back   ctrl+c exit"
		default:
			return "↑/↓ choose   enter continue"
		}
	}

	return footer
}

func (w *wizard) browserChoices(room int) []string {
	if len(w.profiles) == 0 {
		return []string{w.theme.Bad.Render("No supported browser profile was found")}
	}

	room = min(room, len(w.profiles))
	start := min(max(0, w.selected-room/2), len(w.profiles)-room)
	lines := make([]string, 0, room)
	for index := start; index < start+room; index++ {
		profile := w.profiles[index]
		lines = append(lines, w.choice(index == w.selected, ui.SafeInline(profile.Label())))
	}

	return lines
}

func (w *wizard) review(compact bool) []string {
	value := func(label, value string) string {
		return w.theme.Dim.Render(fmt.Sprintf("%-15s", label)) + w.theme.Heading.Render(ui.SafeInline(value))
	}
	lines := []string{w.theme.Heading.Render("Ready to start")}
	if strings.TrimSpace(w.token.Value()) != "" {
		lines = append(lines, value("API token", "Configured · preferred"))
	}
	if w.method == auth.MethodBrowser || w.hasBrowser {
		profile := "No browser selected"
		if len(w.profiles) > 0 {
			profile = w.profiles[w.selected].Label()
		}
		lines = append(lines, value("Browser", profile))
	}
	imageState := "Disabled"
	if w.images {
		imageState = "Enabled · automatic"
	}
	lines = append(lines, value("Terminal images", imageState))
	if compact {
		return lines
	}
	if w.method == auth.MethodBrowser || w.hasBrowser {
		return append(lines,
			"",
			w.theme.Dim.Render("Browser cookies are refreshed at startup and cached in the keyring."),
			w.theme.Dim.Render("Public API operations still prefer the API token when one is configured."),
			w.theme.Dim.Render("Press i to toggle terminal image previews."),
		)
	}

	return append(lines, "",
		w.theme.Dim.Render("Press i to toggle terminal image previews."),
		w.theme.Dim.Render("Run nth setup again to add a browser session or replace the token."),
	)
}
