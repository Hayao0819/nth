package message

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
)

func (s *Screen) createConversation(ctx context.Context, handles []string) tea.Cmd {
	if s.write == nil || len(handles) == 0 || s.acting {
		if len(handles) == 0 {
			s.notice = "Enter at least one account"
		}

		return nil
	}
	s.startAction("Starting conversation…")

	return func() tea.Msg {
		conversation, response, err := s.write.CreateDMConversation(ctx, handles...)

		return actionMsg{target: s, action: actionCreateConversation, conversation: conversation, response: response, err: err}
	}
}

func (s *Screen) acceptRequest(ctx context.Context) tea.Cmd {
	if s.request == nil || s.current == nil || !s.current.Request || s.acting {
		return nil
	}
	conversationID := s.current.ID
	s.startAction("Accepting request…")

	return func() tea.Msg {
		_, response, err := s.request.AcceptDMRequest(ctx, conversationID)

		return actionMsg{target: s, action: actionAcceptRequest, response: response, err: err}
	}
}

func (s *Screen) deleteRequest(ctx context.Context) tea.Cmd {
	if s.request == nil || s.current == nil || !s.current.Request || s.acting {
		return nil
	}
	conversationID := s.current.ID
	s.startAction("Deleting request…")

	return func() tea.Msg {
		_, response, err := s.request.DeleteDMRequest(ctx, conversationID)

		return actionMsg{target: s, action: actionDeleteRequest, response: response, err: err}
	}
}

func (s *Screen) renameConversation(ctx context.Context, name string) tea.Cmd {
	if s.group == nil || s.current == nil || s.acting {
		return nil
	}
	conversationID := s.current.ID
	s.startAction("Renaming conversation…")

	return func() tea.Msg {
		_, response, err := s.group.RenameDMConversation(ctx, conversationID, name)
		nameCopy := name

		return actionMsg{target: s, action: actionRenameConversation, conversation: north.DMConversation{ID: conversationID, Name: &nameCopy}, response: response, err: err}
	}
}

func (s *Screen) addMembers(ctx context.Context, handles []string) tea.Cmd {
	if s.group == nil || s.current == nil || len(handles) == 0 || s.acting {
		if len(handles) == 0 {
			s.notice = "Enter at least one account"
		}

		return nil
	}
	conversationID := s.current.ID
	s.startAction("Adding people…")

	return func() tea.Msg {
		_, response, err := s.group.AddDMConversationMembers(ctx, conversationID, handles...)

		return actionMsg{target: s, action: actionAddMembers, response: response, err: err}
	}
}

func (s *Screen) leaveConversation(ctx context.Context) tea.Cmd {
	if s.group == nil || s.current == nil || s.acting {
		return nil
	}
	conversationID := s.current.ID
	s.startAction("Leaving conversation…")

	return func() tea.Msg {
		_, response, err := s.group.LeaveDMConversation(ctx, conversationID)

		return actionMsg{target: s, action: actionLeaveConversation, response: response, err: err}
	}
}
