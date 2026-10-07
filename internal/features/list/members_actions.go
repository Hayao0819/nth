package list

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type memberChangedMsg struct {
	target   *Members
	handle   string
	added    bool
	response *north.Response
	err      error
}

func (m memberChangedMsg) Response() *north.Response { return m.response }

func (s *Members) promptAdd(ctx *reactea.Ctx) tea.Cmd {
	if !s.list.OwnedByViewer || s.acting {
		return nil
	}
	return modal.PushAt(ctx, dialog.NewTextPrompt(s.theme, "Add member", "@handle", "", "Add"), dialog.Placement(ctx, 52, 8))
}

func (s *Members) confirmRemove(ctx *reactea.Ctx) tea.Cmd {
	if !s.list.OwnedByViewer || s.acting || s.selected < 0 || s.selected >= len(s.users) {
		return nil
	}
	s.removing = true
	user := s.users[s.selected]

	return modal.PushAt(ctx, dialog.NewConfirm(s.theme, "Remove member?", "@"+ui.SafeInline(user.Handle)+" will be removed from this list.", "Remove"), dialog.Placement(ctx, 52, 8))
}

func (s *Members) add(ctx context.Context, value string) tea.Cmd {
	handle := strings.TrimPrefix(strings.TrimSpace(value), "@")
	if handle == "" || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Adding member…"

	return func() tea.Msg {
		_, response, err := s.api.AddListMember(ctx, s.list.ID, handle)

		return memberChangedMsg{target: s, handle: handle, added: true, response: response, err: err}
	}
}

func (s *Members) remove(ctx context.Context) tea.Cmd {
	if s.acting || s.selected < 0 || s.selected >= len(s.users) {
		return nil
	}
	handle := s.users[s.selected].Handle
	s.acting = true
	s.notice = "Removing member…"

	return func() tea.Msg {
		_, response, err := s.api.RemoveListMember(ctx, s.list.ID, handle)

		return memberChangedMsg{target: s, handle: handle, response: response, err: err}
	}
}

func (s *Members) applyMemberChanged(ctx context.Context, msg memberChangedMsg) tea.Cmd {
	if msg.target != s {
		return nil
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return nil
	}
	if msg.added {
		s.notice = "Member added"

		return s.load(ctx, false)
	}
	s.notice = "Member removed"
	for index := range s.users {
		if strings.EqualFold(s.users[index].Handle, msg.handle) {
			s.users = append(s.users[:index], s.users[index+1:]...)
			break
		}
	}
	s.selected = min(s.selected, max(0, len(s.users)-1))

	return nil
}
