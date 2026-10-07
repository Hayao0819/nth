package list

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (s *Members) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case membersLoadedMsg:
		s.applyLoaded(msg)

		return nil
	case memberChangedMsg:
		return s.applyMemberChanged(ctx.Context(), msg)
	case modal.Result[dialog.TextResult]:
		if !msg.Ok() || msg.Value.Canceled {
			return nil
		}

		return s.add(ctx.Context(), msg.Value.Text)
	case modal.Result[dialog.Confirmation]:
		if !s.removing {
			return nil
		}
		s.removing = false
		if !msg.Ok() || !msg.Value.Accepted {
			return nil
		}

		return s.remove(ctx.Context())
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			s.move(1, s.room(ctx.Height()))

			return s.loadNearEnd(ctx.Context())
		}
		if msg.Button == tea.MouseWheelUp {
			s.move(-1, s.room(ctx.Height()))
		}

		return nil
	case tea.MouseClickMsg:
		x, y, inside := reactea.Mouse(ctx, msg)
		if !inside || msg.Button != tea.MouseLeft {
			return nil
		}
		if pageheader.BackAt(x, y) {
			return pageheader.Back()
		}
		index, ok := listview.ItemAt(s.top, len(s.users), y-pageheader.Height, s.room(ctx.Height()), func(int) int { return itemHeight })
		if ok {
			s.selected = index

			return s.open()
		}
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "j", "down"):
		s.move(1, s.room(ctx.Height()))

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "k", "up"):
		s.move(-1, s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.selected, s.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if len(s.users) > 0 {
			s.selected = len(s.users) - 1
			s.move(0, s.room(ctx.Height()))
		}

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "enter", "u"):
		return s.open()
	case reactea.Key(msg, "a"):
		return s.promptAdd(ctx)
	case reactea.Key(msg, "x"):
		return s.confirmRemove(ctx)
	case reactea.Key(msg, "L"):
		return s.load(ctx.Context(), true)
	case reactea.Key(msg, ".") && s.err != nil:
		return s.load(ctx.Context(), false)
	}

	return nil
}
