package message

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type messageAction uint8

const (
	actionCreateConversation messageAction = iota
	actionSendMessage
	actionEditMessage
	actionDeleteMessage
	actionSetReaction
	actionRemoveReaction
	actionAcceptRequest
	actionDeleteRequest
	actionRenameConversation
	actionAddMembers
	actionLeaveConversation
)

type actionMsg struct {
	target       *Screen
	action       messageAction
	conversation north.DMConversation
	message      north.DMMessage
	reactions    []north.DMReaction
	response     *north.Response
	err          error
}

func (m actionMsg) Response() *north.Response { return m.response }

func (s *Screen) applyAction(ctx *reactea.Ctx, msg actionMsg) tea.Cmd {
	if msg.target != s {
		return nil
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return nil
	}
	s.notice = actionNotice(msg.action)
	switch msg.action {
	case actionCreateConversation:
		s.current = &msg.conversation
		s.messages = nil
		s.messageCursor = nil

		return s.loadMessages(ctx.Context(), false)
	case actionDeleteMessage:
		s.messages = slices.DeleteFunc(s.messages, func(message north.DMMessage) bool {
			return message.ID == msg.message.ID
		})
		s.messageIndex = min(s.messageIndex, max(0, len(s.messages)-1))
		s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
	case actionSetReaction, actionRemoveReaction:
		if len(msg.reactions) == 0 {
			return s.loadMessages(ctx.Context(), false)
		}
		for index := range s.messages {
			if s.messages[index].ID == msg.message.ID {
				s.messages[index].Reactions = append([]north.DMReaction(nil), msg.reactions...)
				break
			}
		}
	case actionAcceptRequest, actionDeleteRequest, actionLeaveConversation:
		s.current = nil
		s.messages = nil
		s.messageCursor = nil

		return s.loadConversations(ctx.Context(), false)
	case actionRenameConversation:
		if s.current != nil {
			s.current.Name = msg.conversation.Name
		}
	case actionSendMessage, actionEditMessage, actionAddMembers:
		return s.loadMessages(ctx.Context(), false)
	}

	return nil
}

func (s *Screen) startAction(notice string) {
	s.acting = true
	s.notice = notice
}

func (s *Screen) selectedMessage() *north.DMMessage {
	if s.messageIndex < 0 || s.messageIndex >= len(s.messages) {
		return nil
	}

	return &s.messages[s.messageIndex]
}

func (s *Screen) ownsMessage(message north.DMMessage) bool {
	if s.viewer.ID != "" && message.Sender.ID != "" {
		return s.viewer.ID == message.Sender.ID
	}

	return s.viewer.Handle != "" && strings.EqualFold(s.viewer.Handle, message.Sender.Handle)
}

func viewerReacted(reactions []north.DMReaction) bool {
	return slices.ContainsFunc(reactions, func(reaction north.DMReaction) bool {
		return reaction.ReactedByViewer
	})
}

func parseHandles(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
	})
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		handle := strings.TrimPrefix(strings.TrimSpace(part), "@")
		key := strings.ToLower(handle)
		if handle == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, handle)
	}

	return result
}

func actionNotice(action messageAction) string {
	switch action {
	case actionCreateConversation:
		return "Conversation started"
	case actionSendMessage:
		return "Message sent"
	case actionEditMessage:
		return "Message updated"
	case actionDeleteMessage:
		return "Message deleted"
	case actionSetReaction:
		return "Reaction saved"
	case actionRemoveReaction:
		return "Reaction removed"
	case actionAcceptRequest:
		return "Request accepted"
	case actionDeleteRequest:
		return "Request deleted"
	case actionRenameConversation:
		return "Conversation renamed"
	case actionAddMembers:
		return "People added"
	case actionLeaveConversation:
		return "Conversation left"
	default:
		return ""
	}
}
