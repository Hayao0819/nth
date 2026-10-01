package notification

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
)

func (d *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case PageLoadedMsg:
		return d.applyPageLoaded(ctx, msg)
	case ReadMsg:
		d.applyRead(msg)

		return nil
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			d.move(1, d.innerWidth(ctx.Width()), d.room(ctx.Height()))

			return d.loadNearEnd(ctx.Context())
		}
		if msg.Button == tea.MouseWheelUp {
			d.move(-1, d.innerWidth(ctx.Width()), d.room(ctx.Height()))
		}

		return nil
	case tea.WindowSizeMsg:
		return d.loadImages(ctx.Context(), d.items)
	case tea.MouseClickMsg:
		return d.handleClick(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "j", "down"):
		d.move(1, d.innerWidth(ctx.Width()), d.room(ctx.Height()))

		return d.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "k", "up"):
		d.move(-1, d.innerWidth(ctx.Width()), d.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		d.selected, d.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if len(d.items) > 0 {
			d.selected = len(d.items) - 1
			d.ensureVisible(d.innerWidth(ctx.Width()), d.room(ctx.Height()))
		}

		return d.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "enter"):
		return d.open()
	case reactea.Key(msg, "u"):
		return d.openUser()
	case reactea.Key(msg, "L"):
		return d.load(ctx.Context(), true)
	case reactea.Key(msg, "."):
		return d.load(ctx.Context(), false)
	}

	return nil
}

func (d *Screen) handleClick(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside || msg.Button != tea.MouseLeft {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	row := y - d.bodyTop()
	if index, ok := d.itemAt(row, d.innerWidth(ctx.Width()), d.room(ctx.Height())); ok {
		d.selected = index
		d.ensureVisible(d.innerWidth(ctx.Width()), d.room(ctx.Height()))

		return d.open()
	}

	return nil
}
