package account

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
	case changedMsg:
		return s.applyChanged(ctx.Context(), msg)
	case modal.Result[dialog.FormResult]:
		if !msg.Ok() || msg.Value.Canceled || s.tab != keywordsTab {
			return nil
		}

		return s.createKeyword(ctx.Context(), msg.Value)
	case modal.Result[dialog.Confirmation]:
		kind := s.confirm
		s.confirm = changeNone
		if kind == changeNone || !msg.Ok() || !msg.Value.Accepted {
			return nil
		}

		return s.change(ctx.Context(), kind)
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
		return s.click(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "tab", "right"):
		s.setTab(s.tab + 1)

		return s.load(ctx.Context(), false)
	case reactea.Key(msg, "shift+tab"):
		maximum := keywordsTab
		if s.keywords == nil {
			maximum = mutedTab
		}
		if s.tab == requestsTab {
			s.setTab(maximum)
		} else {
			s.setTab(s.tab - 1)
		}

		return s.load(ctx.Context(), false)
	case reactea.Key(msg, "j", "down"):
		s.move(1, s.room(ctx.Height()))

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "k", "up"):
		s.move(-1, s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.selected, s.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if s.count() > 0 {
			s.selected = s.count() - 1
			s.move(0, s.room(ctx.Height()))
		}

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "enter", "u"):
		return s.openUser()
	case reactea.Key(msg, "y") && s.tab == requestsTab:
		return s.change(ctx.Context(), changeAccept)
	case reactea.Key(msg, "x") && s.tab == requestsTab:
		return s.confirmChange(ctx, changeReject)
	case reactea.Key(msg, "x") && s.tab == blockedTab:
		return s.confirmChange(ctx, changeUnblock)
	case reactea.Key(msg, "x") && s.tab == mutedTab:
		return s.confirmChange(ctx, changeUnmute)
	case reactea.Key(msg, "c") && s.tab == keywordsTab:
		return s.openKeywordEditor(ctx)
	case reactea.Key(msg, "x") && s.tab == keywordsTab:
		return s.confirmChange(ctx, changeDeleteKeyword)
	case reactea.Key(msg, "L"):
		return s.load(ctx.Context(), true)
	case reactea.Key(msg, ".") && s.err != nil:
		return s.load(ctx.Context(), false)
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
		columns := 3
		if s.keywords != nil {
			columns = 4
		}
		column := min(columns-1, max(0, (x-ctx.Width()/2)*columns/max(1, ctx.Width()-ctx.Width()/2)))
		s.setTab(tab(column))

		return s.load(ctx.Context(), false)
	}
	index, ok := listview.ItemAt(s.top, s.count(), y-pageheader.Height, s.room(ctx.Height()), func(int) int { return accountItemHeight })
	if !ok {
		return nil
	}
	s.selected = index

	return s.openUser()
}
