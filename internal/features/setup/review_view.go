package setup

import (
	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/nth/internal/ui"
)

func (w *wizard) review(compact bool, width int) []string {
	title, detail := w.selectedAuthentication()
	lines := []string{
		w.theme.Heading.Render("Review and save"),
		w.theme.Dim.Render("SIGN-IN"),
		w.theme.Active.Render("✓ " + title),
		"  " + w.theme.Dim.Render(detail),
	}
	if w.envToken && w.method == auth.MethodOAuth {
		for _, line := range ui.WrappedLines("NORTH_API_KEY remains active; OAuth will be saved as a fallback.", max(1, width-2)) {
			lines = append(lines, w.theme.Warn.Render("  "+line))
		}
	}
	lines = append(lines,
		"",
		w.theme.Dim.Render("OPTIONAL"),
		w.settingLine("Browser cookies", w.browserState(), w.browserKey(), width),
		w.settingLine("Terminal images", imageState(w.images), "i toggle", width),
	)
	if !compact {
		lines = append(lines,
			"",
			w.theme.Dim.Render("Browser cookies are used only for features missing from the public API."),
			w.theme.Dim.Render("Press enter to save these settings."),
		)
	}

	return lines
}

func (w *wizard) selectedAuthentication() (string, string) {
	if w.method == auth.MethodOAuth {
		return "north account", "Signed in · token refreshes automatically"
	}
	if w.envToken {
		return "API token", "NORTH_API_KEY · active"
	}

	return "API token", "Will be stored in the system keyring"
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
