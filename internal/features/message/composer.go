package message

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const composerHeight = 5

func (s *Screen) composerVisible() bool {
	return s.write != nil && s.current != nil && !s.current.Request
}

func (s *Screen) openComposer(ctx *reactea.Ctx) tea.Cmd {
	if !s.composerVisible() {
		return nil
	}
	s.composing = true
	s.composerError = ""
	s.ensureVisible(ctx.Width(), s.room(ctx.Height()))

	return tea.Batch(s.composer.Widget.Focus(), s.inputCapture.CaptureInput(ctx))
}

func (s *Screen) closeComposer() tea.Cmd {
	s.composing = false
	s.composerError = ""
	s.composer.Widget.Blur()

	return s.inputCapture.ReleaseInput()
}

func (s *Screen) updateComposer(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch {
	case reactea.Key(msg, "esc"):
		return s.closeComposer()
	case reactea.Key(msg, "shift+enter"):
		if !s.acting {
			s.composerError = ""
			s.composer.Widget.InsertString("\n")
		}

		return nil
	case reactea.Key(msg, "enter", "ctrl+enter", "ctrl+j", "ctrl+m"):
		return s.submitComposer(ctx)
	}
	if s.acting || !reactea.IsInput(msg) {
		return nil
	}
	s.composerError = ""

	return s.composer.Update(s.composerCtx(ctx), msg)
}

func (s *Screen) submitComposer(ctx *reactea.Ctx) tea.Cmd {
	if s.acting {
		return nil
	}
	text := strings.TrimSpace(s.composer.Widget.Value())
	if text == "" {
		s.composerError = "Enter a message"

		return nil
	}
	s.composerError = ""
	s.submittedText = text

	return s.sendMessage(ctx.Context(), text, "")
}

func (s *Screen) finishComposerSubmission(success bool) {
	if !success {
		s.submittedText = ""

		return
	}
	if strings.TrimSpace(s.composer.Widget.Value()) == s.submittedText {
		s.composer.Widget.SetValue("")
	}
	s.submittedText = ""
	s.composerError = ""
}

func (s *Screen) renderComposer(ctx *reactea.Ctx) string {
	width := ctx.Width()
	border := s.theme.Dim.GetForeground()
	if s.composing {
		border = s.theme.Active.GetForeground()
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width).
		MaxWidth(width).
		Render(s.composer.Render(s.composerCtx(ctx)))
	left := "Enter write · m new DM · j/k move · Esc back · r reply · e edit · +/- react · d/D delete"
	if !s.composing && s.messageScrollLimit(width, s.room(ctx.Height())) > 0 {
		left = "Ctrl+D/U scroll message · " + left
	}
	leftStyle := s.theme.Dim
	if s.composing {
		left = "Esc  Done · Shift+Enter  New line"
	}
	if s.composerError != "" {
		left = s.composerError
		leftStyle = s.theme.Bad
	} else if s.acting {
		left = "Sending…"
	}
	left = leftStyle.Render(ui.Clip(left, max(1, width-lipgloss.Width(s.composerButton())-1)))

	return box + "\n" + ui.Sides(left, s.composerButton(), width)
}

func (s *Screen) composerButton() string {
	label := "Enter  Write"
	if s.composing {
		label = "Enter  Send"
	}
	if s.acting {
		label = "Sending…"
	}

	return s.theme.Button.Padding(0, 1).Render(label)
}

func (s *Screen) composerCtx(ctx *reactea.Ctx) *reactea.Ctx {
	return ctx.Inset(2, max(0, ctx.Height()-composerHeight+1), max(1, ctx.Width()-4), 2)
}
