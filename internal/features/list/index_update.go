package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type stateChangedMsg struct {
	target   *Index
	id       string
	kind     stateKind
	enabled  bool
	response *north.Response
	err      error
}

func (m stateChangedMsg) Response() *north.Response { return m.response }

type stateKind uint8

const (
	stateFollow stateKind = iota
	statePin
)

func (s *Index) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case modal.Result[dialog.FormResult]:
		if !msg.Ok() || msg.Value.Canceled {
			return nil
		}

		return s.create(ctx.Context(), msg.Value)
	case indexLoadedMsg:
		if msg.target != s {
			return nil
		}
		s.loading = false
		s.err = msg.err
		if msg.err == nil {
			s.entries = flatten(msg.collection)
			s.selected, s.top = 0, 0
		}

		return nil
	case stateChangedMsg:
		return s.applyState(msg)
	case createdMsg:
		s.applyCreated(msg)

		return nil
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
	case reactea.Key(msg, "j", "down"):
		s.move(1, s.room(ctx.Height()))
	case reactea.Key(msg, "k", "up"):
		s.move(-1, s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.selected, s.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if len(s.entries) > 0 {
			s.selected = len(s.entries) - 1
			s.ensureVisible(s.room(ctx.Height()))
		}
	case reactea.Key(msg, "enter"):
		return s.open()
	case reactea.Key(msg, "u"):
		return s.openOwner()
	case reactea.Key(msg, "f"):
		return s.toggle(ctx.Context(), stateFollow)
	case reactea.Key(msg, "p"):
		return s.toggle(ctx.Context(), statePin)
	case reactea.Key(msg, "c"):
		return s.openCreate(ctx)
	case reactea.Key(msg, "."):
		return s.load(ctx.Context())
	}

	return nil
}

func (s *Index) click(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	if msg.Button != tea.MouseLeft {
		return nil
	}
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	index, ok := s.entryAt(y - pageheader.Height)
	if !ok {
		return nil
	}
	s.selected = index
	s.ensureVisible(s.room(ctx.Height()))

	return s.open()
}

func (s *Index) open() tea.Cmd {
	if s.selected < 0 || s.selected >= len(s.entries) {
		return nil
	}

	return navigation.OpenList(s.entries[s.selected].item)
}

func (s *Index) openOwner() tea.Cmd {
	if s.selected < 0 || s.selected >= len(s.entries) {
		return nil
	}

	return navigation.OpenUser(s.entries[s.selected].item.Owner)
}

func (s *Index) toggle(ctx context.Context, kind stateKind) tea.Cmd {
	if s.acting || s.selected < 0 || s.selected >= len(s.entries) {
		return nil
	}
	item := s.entries[s.selected].item
	if kind == stateFollow && item.OwnedByViewer {
		s.notice = "You own this list"

		return nil
	}
	s.acting = true
	s.notice = ""

	return func() tea.Msg {
		var (
			enabled  bool
			response *north.Response
			err      error
		)
		switch kind {
		case stateFollow:
			enabled = !item.FollowedByViewer
			if enabled {
				_, response, err = s.api.FollowList(ctx, item.ID)
			} else {
				_, response, err = s.api.UnfollowList(ctx, item.ID)
			}
		case statePin:
			enabled = !item.PinnedByViewer
			if enabled {
				_, response, err = s.api.PinList(ctx, item.ID)
			} else {
				_, response, err = s.api.UnpinList(ctx, item.ID)
			}
		}

		return stateChangedMsg{target: s, id: item.ID, kind: kind, enabled: enabled, response: response, err: err}
	}
}

func (s *Index) applyState(msg stateChangedMsg) tea.Cmd {
	if msg.target != s {
		return nil
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return nil
	}
	for index := range s.entries {
		if s.entries[index].item.ID != msg.id {
			continue
		}
		switch msg.kind {
		case stateFollow:
			s.entries[index].item.FollowedByViewer = msg.enabled
		case statePin:
			s.entries[index].item.PinnedByViewer = msg.enabled
		}
	}

	return nil
}
