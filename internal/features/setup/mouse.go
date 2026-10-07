package setup

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/charmbracelet/x/ansi"
)

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
