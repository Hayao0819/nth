package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type createdMsg struct {
	target   *Index
	item     north.List
	response *north.Response
	err      error
}

func (m createdMsg) Response() *north.Response { return m.response }

func (s *Index) openCreate(ctx *reactea.Ctx) tea.Cmd {
	if s.editor == nil || s.acting {
		return nil
	}

	return modal.PushAt(ctx, NewForm(s.theme, "Create list", "Create", nil), dialog.Placement(ctx, 62, 12))
}

func (s *Index) create(ctx context.Context, result dialog.FormResult) tea.Cmd {
	if s.editor == nil || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Creating list…"
	request := CreateRequest(result)

	return func() tea.Msg {
		item, response, err := s.editor.CreateList(ctx, request)

		return createdMsg{target: s, item: item, response: response, err: err}
	}
}

func (s *Index) applyCreated(msg createdMsg) {
	if msg.target != s {
		return
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return
	}
	s.notice = "List created"
	s.entries = append([]entry{{section: "Your lists", item: msg.item}}, s.entries...)
	s.selected, s.top = 0, 0
}
