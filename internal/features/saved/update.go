package saved

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (s *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case loadedMsg:
		s.applyLoaded(msg)

		return nil
	case actionMsg:
		return s.applyAction(ctx.Context(), msg)
	case modal.Result[dialog.TextResult]:
		if s.editor != editorDraft && s.editor != editorPending {
			return nil
		}
		if !msg.Ok() || msg.Value.Canceled {
			s.editor, s.editingID = editorNone, ""

			return nil
		}

		return s.saveText(ctx.Context(), msg.Value.Text)
	case modal.Result[dialog.FormResult]:
		if s.editor != editorScheduled {
			return nil
		}
		if !msg.Ok() || msg.Value.Canceled {
			s.editor, s.editingID = editorNone, ""

			return nil
		}

		return s.saveSchedule(ctx.Context(), msg.Value)
	case modal.Result[dialog.Confirmation]:
		if !s.confirm {
			return nil
		}
		s.confirm = false
		if !msg.Ok() || !msg.Value.Accepted {
			s.deleteID = ""

			return nil
		}

		return s.delete(ctx.Context())
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			s.move(1, s.room(ctx.Height()))
		} else if msg.Button == tea.MouseWheelUp {
			s.move(-1, s.room(ctx.Height()))
		}

		return nil
	case tea.MouseClickMsg:
		return s.click(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "tab", "right"):
		s.setTab((s.tab + 1) % 3)

		return s.load(ctx.Context())
	case reactea.Key(msg, "shift+tab"):
		s.setTab((s.tab + 2) % 3)

		return s.load(ctx.Context())
	case reactea.Key(msg, "j", "down"):
		s.move(1, s.room(ctx.Height()))
	case reactea.Key(msg, "k", "up"):
		s.move(-1, s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.selected, s.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if s.count() > 0 {
			s.selected = s.count() - 1
			s.move(0, s.room(ctx.Height()))
		}
	case reactea.Key(msg, "enter", "e"):
		return s.openEditor(ctx, false)
	case reactea.Key(msg, "c"):
		return s.openEditor(ctx, true)
	case reactea.Key(msg, "d"):
		return s.confirmDelete(ctx)
	case reactea.Key(msg, "s"):
		return s.pendingAction(ctx.Context(), false)
	case reactea.Key(msg, "u"):
		return s.pendingAction(ctx.Context(), true)
	case reactea.Key(msg, ".") && s.err != nil:
		return s.load(ctx.Context())
	}

	return nil
}

func (s *Screen) click(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside || msg.Button != tea.MouseLeft {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	if y < pageheader.Height && x >= ctx.Width()/2 {
		column := min(2, max(0, (x-ctx.Width()/2)*3/max(1, ctx.Width()-ctx.Width()/2)))
		s.setTab(tab(column))

		return s.load(ctx.Context())
	}
	index, ok := listview.ItemAt(s.top, s.count(), y-pageheader.Height, s.room(ctx.Height()), func(int) int { return savedItemHeight })
	if !ok {
		return nil
	}
	s.selected = index
	if s.tab == pendingTab {
		return nil
	}

	return s.openEditor(ctx, false)
}
