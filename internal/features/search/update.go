package search

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if delayed, ok := msg.(debounceMsg); ok {
		return s.applyDebounce(ctx, delayed)
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		return s.handleClick(ctx, click)
	}
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		if _, y, inside := reactea.Mouse(ctx, msg); inside && y >= pageheader.Height {
			return s.updateFeed(ctx, wheel)
		}

		return nil
	}

	switch {
	case reactea.Key(msg, "esc"):
		return pageheader.Back()
	case reactea.Key(msg, "enter"):
		if post := s.feed.SelectedPost(); post != nil {
			return navigation.OpenPost(*post)
		}
		query := strings.TrimSpace(s.input.Widget.Value())
		if query != "" {
			s.seq++
			s.pending = false

			return s.feed.SetMode(s.feedCtx(ctx), feed.Search, query)
		}
	case reactea.Key(msg, "down", "ctrl+n", "ctrl+j"):
		return s.feed.Update(s.feedCtx(ctx), tea.KeyPressMsg{Code: tea.KeyDown})
	case reactea.Key(msg, "up", "ctrl+p", "ctrl+k"):
		return s.feed.Update(s.feedCtx(ctx), tea.KeyPressMsg{Code: tea.KeyUp})
	case reactea.Key(msg, "pgdown", "ctrl+d", "pgup", "ctrl+u"):
		return s.feed.Update(s.feedCtx(ctx), msg)
	}

	if !reactea.IsInput(msg) {
		return s.feed.Update(s.feedCtx(ctx), msg)
	}

	return s.updateInput(ctx, msg)
}

func (s *Screen) applyDebounce(ctx *reactea.Ctx, msg debounceMsg) tea.Cmd {
	if msg.target != s || msg.seq != s.seq {
		return nil
	}
	query := strings.TrimSpace(msg.query)
	s.pending = false
	if query == "" {
		s.feed.ClearSearch("")

		return nil
	}

	return s.feed.SetMode(s.feedCtx(ctx), feed.Search, query)
}

func (s *Screen) handleClick(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	if y < pageheader.Height {
		return s.input.Widget.Focus()
	}

	return s.updateFeed(ctx, msg)
}

func (s *Screen) updateFeed(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	feedCtx := s.feedCtx(ctx)
	if reactea.IsMouse(msg) {
		parentX, parentY := ctx.Origin()
		feedX, feedY := feedCtx.Origin()
		msg = reactea.TranslateMouse(msg, feedX-parentX, feedY-parentY)
	}

	return s.feed.Update(feedCtx, msg)
}

func (s *Screen) updateInput(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	before := s.input.Widget.Value()
	command := s.input.Update(s.inputCtx(ctx), msg)
	after := s.input.Widget.Value()
	if before == after {
		return command
	}

	s.seq++
	query := strings.TrimSpace(after)
	s.feed.ClearSearch(query)
	if query == "" {
		s.pending = false

		return command
	}
	s.pending = true

	return tea.Batch(command, s.debounce(query, debounceDelay))
}
