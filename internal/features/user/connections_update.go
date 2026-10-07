package user

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Connections) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case loadedMsg:
		s.applyLoaded(msg, ctx.Width(), s.room(ctx.Height()))

		return nil
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			s.move(1, ctx.Width(), s.room(ctx.Height()))

			return s.loadNearEnd(ctx.Context())
		}
		if msg.Button == tea.MouseWheelUp {
			s.move(-1, ctx.Width(), s.room(ctx.Height()))
		}

		return nil
	case tea.MouseClickMsg:
		return s.click(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "j", "down"):
		s.move(1, ctx.Width(), s.room(ctx.Height()))

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "k", "up"):
		s.move(-1, ctx.Width(), s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.selected, s.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if len(s.items) > 0 {
			s.selected = len(s.items) - 1
			s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
		}

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "enter", "u"):
		return s.open()
	case reactea.Key(msg, "L"):
		return s.load(ctx.Context(), true)
	case reactea.Key(msg, ".") && s.err != nil:
		return s.load(ctx.Context(), false)
	}

	return nil
}
