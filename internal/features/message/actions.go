package message

import (
	"slices"
	"strings"
	"time"

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
	submittedText := ""
	followLatest := false
	if msg.action == actionSendMessage {
		submittedText = s.submittedText
		followLatest = s.atLatestMessage(ctx.Width(), s.room(ctx.Height()))
		s.finishComposerSubmission(msg.err == nil)
	}
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
		s.messageAnchor = messageViewAnchor{}

		return s.loadMessages(ctx.Context(), false)
	case actionSendMessage:
		if index, ok := s.addSentMessage(msg.message, submittedText); ok {
			if followLatest && index >= 0 {
				s.setIndex(index)
				s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
			}

			return nil
		}
		s.saveMessageView(ctx.Width(), s.room(ctx.Height()))

		return s.loadMessages(ctx.Context(), false)
	case actionDeleteMessage:
		s.messages = slices.DeleteFunc(s.messages, func(message north.DMMessage) bool {
			return message.ID == msg.message.ID
		})
		s.messageIndex = min(s.messageIndex, max(0, len(s.messages)-1))
		s.messageOffset = 0
		s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
	case actionSetReaction, actionRemoveReaction:
		if len(msg.reactions) == 0 {
			s.saveMessageView(ctx.Width(), s.room(ctx.Height()))

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
		s.messageAnchor = messageViewAnchor{}

		return s.loadConversations(ctx.Context(), false)
	case actionRenameConversation:
		if s.current != nil {
			s.current.Name = msg.conversation.Name
		}
	case actionEditMessage, actionAddMembers:
		s.saveMessageView(ctx.Width(), s.room(ctx.Height()))

		return s.loadMessages(ctx.Context(), false)
	}

	return nil
}

func (s *Screen) addSentMessage(message north.DMMessage, fallbackText string) (int, bool) {
	if strings.TrimSpace(message.ID) == "" {
		return 0, false
	}
	if message.Text == "" {
		message.Text = fallbackText
	}
	if message.Sender.ID == "" && message.Sender.Handle == "" && message.Sender.Name == "" {
		message.Sender = s.viewer
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}
	conversationID := message.ConversationID
	if conversationID == "" && s.current != nil {
		conversationID = s.current.ID
		message.ConversationID = conversationID
	}
	if s.current == nil || s.current.ID != conversationID {
		s.updateConversationPreview(conversationID, message)

		return -1, true
	}

	index := slices.IndexFunc(s.messages, func(current north.DMMessage) bool {
		return current.ID == message.ID
	})
	if index >= 0 {
		s.messages[index] = message
	} else {
		s.messages = append(s.messages, message)
		index = len(s.messages) - 1
	}
	if s.current != nil {
		copy := message
		s.current.LastMessage = &copy
		s.current.UpdatedAt = message.CreatedAt
	}
	s.updateConversationPreview(conversationID, message)

	return index, true
}

func (s *Screen) updateConversationPreview(conversationID string, message north.DMMessage) {
	for conversationIndex := range s.conversations {
		if s.conversations[conversationIndex].ID != conversationID {
			continue
		}
		copy := message
		s.conversations[conversationIndex].LastMessage = &copy
		s.conversations[conversationIndex].UpdatedAt = message.CreatedAt
	}
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
