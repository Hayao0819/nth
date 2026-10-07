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
			"nth setup", "", fmt.Sprintf("terminal is %d×%d", width, height),
			"resize to at least 42×14", "", "ctrl+c exit",
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
		w.theme.Dim.Render(fmt.Sprintf("Step %d of %d", w.step+1, stepReview+1)),
		contentWidth,
	)
	lines := []string{header, w.progress(contentWidth), ""}
	lines = append(lines, body...)
	lines = append(lines, w.theme.Dim.Render(w.footer(contentWidth)))
	for index, line := range lines {
		lines[index] = ui.Clip(line, contentWidth)
	}

	panel := panelStyle.Width(panelWidth).Height(panelHeight).Render(strings.Join(lines, "\n"))

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
	labels := []string{"1  Method", "2  Credentials", "3  Review"}
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

func (w *wizard) footer(width int) string {
	footer := "↑/↓ choose   enter continue   ctrl+c exit"
	switch w.step {
	case stepMethod:
		if w.method == auth.MethodOAuth && w.oauth.NeedsAuthorization() {
			footer = "enter authorize   ctrl+c exit"
		} else if w.method == auth.MethodOAuth && w.oauth.Valid() {
			footer = "enter review   r sign in again   ctrl+c exit"
		}
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
		footer = "i images   enter save   esc back"
		if len(w.profiles) > 0 {
			footer = "b browser   " + footer
		}
	}
	if width < 58 {
		switch w.step {
		case stepReview:
			footer = "i images   enter save"
			if len(w.profiles) > 0 {
				footer = "b browser   " + footer
			}
		case stepCredential:
			if w.method == auth.MethodOAuth {
				if w.oauthBusy {
					footer = "esc cancel"
				} else {
					footer = "enter retry   esc back"
				}
			} else {
				footer = "enter continue   esc back"
			}
		case stepMethod:
			if w.method == auth.MethodOAuth && w.oauth.NeedsAuthorization() {
				footer = "enter authorize"
			} else if w.method == auth.MethodOAuth && w.oauth.Valid() {
				footer = "enter review   r sign in again"
			} else {
				footer = "↑/↓ choose   enter continue"
			}
		default:
			footer = "↑/↓ choose   enter continue"
		}
	}

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, footer)
}
