package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type editedMsg struct {
	target   *Detail
	item     north.List
	response *north.Response
	err      error
}

func (m editedMsg) Response() *north.Response { return m.response }

type deletedMsg struct {
	target   *Detail
	response *north.Response
	err      error
}

func (m deletedMsg) Response() *north.Response { return m.response }

func (s *Detail) openEdit(ctx *reactea.Ctx) tea.Cmd {
	if s.editor == nil || !s.item.OwnedByViewer || s.acting {
		return nil
	}

	return modal.PushAt(ctx, NewForm(s.theme, "Edit list", "Save", &s.item), dialog.Placement(ctx, 62, 12))
}

func (s *Detail) updateList(ctx context.Context, result dialog.FormResult) tea.Cmd {
	if s.editor == nil || !s.item.OwnedByViewer || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Saving list…"
	id := s.item.ID
	request := UpdateRequest(result)

	return func() tea.Msg {
		item, response, err := s.editor.UpdateList(ctx, id, request)

		return editedMsg{target: s, item: item, response: response, err: err}
	}
}

func (s *Detail) confirmDelete(ctx *reactea.Ctx) tea.Cmd {
	if s.editor == nil || !s.item.OwnedByViewer || s.acting || s.deleting {
		return nil
	}
	s.deleting = true

	return modal.PushAt(ctx, dialog.NewConfirm(s.theme, "Delete list?", "The list will be permanently deleted.", "Delete"), dialog.Placement(ctx, 52, 8))
}

func (s *Detail) deleteList(ctx context.Context) tea.Cmd {
	if s.editor == nil || !s.item.OwnedByViewer || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Deleting list…"
	id := s.item.ID

	return func() tea.Msg {
		_, response, err := s.editor.DeleteList(ctx, id)

		return deletedMsg{target: s, response: response, err: err}
	}
}

func (s *Detail) applyEdited(msg editedMsg) {
	if msg.target != s {
		return
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return
	}
	s.item = msg.item
	s.notice = "List updated"
}

func (s *Detail) applyDeleted(msg deletedMsg) tea.Cmd {
	if msg.target != s {
		return nil
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return nil
	}

	return pageheader.Back()
}
