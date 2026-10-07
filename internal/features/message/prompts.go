package message

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type promptKind uint8

const (
	promptNone promptKind = iota
	promptConversation
	promptMessage
	promptReply
	promptEdit
	promptReaction
	promptRename
	promptMembers
)

type confirmKind uint8

const (
	confirmNone confirmKind = iota
	confirmDeleteMessage
	confirmDeleteMessageForEveryone
	confirmDeleteRequest
	confirmLeaveConversation
)

func (s *Screen) promptFor(ctx *reactea.Ctx, kind promptKind) tea.Cmd {
	if s.write == nil && kind <= promptReaction {
		return nil
	}
	s.promptID = ""
	title, placeholder, initial, button := "", "", "", "Save"
	multiline := false
	switch kind {
	case promptConversation:
		title, placeholder, button = "New message", "@alice, @bob", "Start"
	case promptMessage:
		title, placeholder, button, multiline = "New message", "Write a message", "Send", true
	case promptReply:
		message := s.selectedMessage()
		if message == nil {
			return nil
		}
		s.promptID = message.ID
		title, placeholder, button, multiline = "Reply", "Write a reply", "Send", true
	case promptEdit:
		message := s.selectedMessage()
		if message == nil || !s.ownsMessage(*message) {
			s.notice = "Only your messages can be edited"

			return nil
		}
		s.promptID = message.ID
		title, placeholder, initial, button, multiline = "Edit message", "Message", message.Text, "Save", true
	case promptReaction:
		message := s.selectedMessage()
		if message == nil {
			return nil
		}
		s.promptID = message.ID
		title, placeholder, button = "React", "Emoji", "React"
	case promptRename:
		if s.current == nil || s.group == nil || !s.current.Group {
			return nil
		}
		initial = conversationTitle(*s.current)
		title, placeholder, button = "Rename conversation", "Conversation name", "Save"
	case promptMembers:
		if s.current == nil || s.group == nil || !s.current.Group {
			return nil
		}
		title, placeholder, button = "Add people", "@alice, @bob", "Add"
	default:
		return nil
	}
	s.prompt = kind
	if multiline {
		return modal.PushAt(ctx, dialog.NewTextEditor(s.theme, title, placeholder, initial, button), dialog.Placement(ctx, 64, 14))
	}

	return modal.PushAt(ctx, dialog.NewTextPrompt(s.theme, title, placeholder, initial, button), dialog.Placement(ctx, 56, 8))
}

func (s *Screen) applyPrompt(ctx context.Context, text string) tea.Cmd {
	kind := s.prompt
	id := s.promptID
	s.prompt = promptNone
	s.promptID = ""
	switch kind {
	case promptConversation:
		return s.createConversation(ctx, parseHandles(text))
	case promptMessage:
		return s.sendMessage(ctx, text, "")
	case promptReply:
		return s.sendMessage(ctx, text, id)
	case promptEdit:
		return s.editMessage(ctx, id, text)
	case promptReaction:
		return s.setReaction(ctx, id, text)
	case promptRename:
		return s.renameConversation(ctx, text)
	case promptMembers:
		return s.addMembers(ctx, parseHandles(text))
	default:
		return nil
	}
}

func (s *Screen) confirmAction(ctx *reactea.Ctx, kind confirmKind) tea.Cmd {
	if s.current == nil {
		return nil
	}
	title, message, button := "", "", "Delete"
	switch kind {
	case confirmDeleteMessage:
		if s.selectedMessage() == nil || s.write == nil {
			return nil
		}
		title, message = "Delete message?", "The message will be removed from your conversation."
	case confirmDeleteMessageForEveryone:
		selected := s.selectedMessage()
		if selected == nil || s.write == nil || !s.ownsMessage(*selected) {
			s.notice = "Only your messages can be deleted for everyone"

			return nil
		}
		title, message = "Delete for everyone?", "The message will be removed for every participant."
	case confirmDeleteRequest:
		if !s.current.Request || s.request == nil {
			return nil
		}
		title, message = "Delete message request?", "The request and its message history will be removed."
	case confirmLeaveConversation:
		if !s.current.Group || s.group == nil {
			return nil
		}
		title, message, button = "Leave conversation?", "You will no longer receive messages from this group.", "Leave"
	default:
		return nil
	}
	s.confirm = kind

	return modal.PushAt(ctx, dialog.NewConfirm(s.theme, title, message, button), dialog.Placement(ctx, 56, 8))
}

func (s *Screen) applyConfirmation(ctx context.Context, accepted bool) tea.Cmd {
	kind := s.confirm
	s.confirm = confirmNone
	if !accepted {
		return nil
	}
	switch kind {
	case confirmDeleteMessage:
		return s.deleteMessage(ctx, false)
	case confirmDeleteMessageForEveryone:
		return s.deleteMessage(ctx, true)
	case confirmDeleteRequest:
		return s.deleteRequest(ctx)
	case confirmLeaveConversation:
		return s.leaveConversation(ctx)
	default:
		return nil
	}
}
