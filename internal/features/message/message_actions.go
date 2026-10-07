package message

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
)

func (s *Screen) sendMessage(ctx context.Context, text, replyID string) tea.Cmd {
	if s.write == nil || s.current == nil || s.acting {
		return nil
	}
	request := north.SendDMRequest{Text: text}
	if replyID != "" {
		request.ReplyToID = &replyID
	}
	conversationID := s.current.ID
	s.startAction("Sending message…")

	return func() tea.Msg {
		message, response, err := s.write.SendDM(ctx, conversationID, request)

		return actionMsg{target: s, action: actionSendMessage, message: message, response: response, err: err}
	}
}

func (s *Screen) editMessage(ctx context.Context, id, text string) tea.Cmd {
	if s.write == nil || s.current == nil || id == "" || s.acting {
		return nil
	}
	conversationID := s.current.ID
	s.startAction("Saving message…")

	return func() tea.Msg {
		message, response, err := s.write.EditDM(ctx, conversationID, id, text)

		return actionMsg{target: s, action: actionEditMessage, message: message, response: response, err: err}
	}
}

func (s *Screen) deleteMessage(ctx context.Context, everyone bool) tea.Cmd {
	message := s.selectedMessage()
	if s.write == nil || s.current == nil || message == nil || s.acting {
		return nil
	}
	conversationID, messageID := s.current.ID, message.ID
	s.startAction("Deleting message…")

	return func() tea.Msg {
		_, response, err := s.write.DeleteDM(ctx, conversationID, messageID, everyone)

		return actionMsg{target: s, action: actionDeleteMessage, message: north.DMMessage{ID: messageID}, response: response, err: err}
	}
}

func (s *Screen) setReaction(ctx context.Context, id, emoji string) tea.Cmd {
	if s.write == nil || s.current == nil || id == "" || s.acting {
		return nil
	}
	conversationID := s.current.ID
	s.startAction("Saving reaction…")

	return func() tea.Msg {
		reactions, response, err := s.write.SetDMReaction(ctx, conversationID, id, emoji)

		return actionMsg{target: s, action: actionSetReaction, message: north.DMMessage{ID: id}, reactions: reactions, response: response, err: err}
	}
}

func (s *Screen) removeReaction(ctx context.Context) tea.Cmd {
	message := s.selectedMessage()
	if s.write == nil || s.current == nil || message == nil || s.acting || !viewerReacted(message.Reactions) {
		return nil
	}
	conversationID, messageID := s.current.ID, message.ID
	s.startAction("Removing reaction…")

	return func() tea.Msg {
		reactions, response, err := s.write.RemoveDMReaction(ctx, conversationID, messageID)

		return actionMsg{target: s, action: actionRemoveReaction, message: north.DMMessage{ID: messageID}, reactions: reactions, response: response, err: err}
	}
}
