package message

import (
	"context"
	"slices"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	messagedomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/support/collection"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type conversationsLoadedMsg struct {
	target   *Screen
	page     messagedomain.MessageConversationPage
	response *north.Response
	err      error
	more     bool
}

func (m conversationsLoadedMsg) Response() *north.Response { return m.response }

type messagesLoadedMsg struct {
	target   *Screen
	page     messagedomain.MessagePage
	response *north.Response
	err      error
	more     bool
}

func (m messagesLoadedMsg) Response() *north.Response { return m.response }

type readMsg struct {
	target         *Screen
	conversationID string
	response       *north.Response
	err            error
}

func (m readMsg) Response() *north.Response { return m.response }

func (s *Screen) loadConversations(ctx context.Context, more bool) tea.Cmd {
	if s.api == nil || s.loading || more && s.conversationCursor == nil {
		return nil
	}
	s.loading = true
	s.loadingMore = more
	s.err = nil
	cursor := ""
	if more {
		cursor = *s.conversationCursor
	}

	return func() tea.Msg {
		page, response, err := s.api.DMConversations(ctx, cursor, s.requests)

		return conversationsLoadedMsg{target: s, page: page, response: response, err: err, more: more}
	}
}

func (s *Screen) loadMessages(ctx context.Context, more bool) tea.Cmd {
	if s.api == nil || s.current == nil || s.loading || more && s.messageCursor == nil {
		return nil
	}
	s.loading = true
	s.loadingMore = more
	s.err = nil
	cursor := ""
	if more {
		cursor = *s.messageCursor
	}
	conversationID := s.current.ID

	return func() tea.Msg {
		page, response, err := s.api.DMMessages(ctx, conversationID, cursor)

		return messagesLoadedMsg{target: s, page: page, response: response, err: err, more: more}
	}
}

func (s *Screen) markRead(ctx context.Context, conversationID string) tea.Cmd {
	return func() tea.Msg {
		response, err := s.api.MarkDMRead(ctx, conversationID)

		return readMsg{target: s, conversationID: conversationID, response: response, err: err}
	}
}

func (s *Screen) openConversation(ctx context.Context) tea.Cmd {
	if s.conversationIndex < 0 || s.conversationIndex >= len(s.conversations) {
		return nil
	}
	conversation := s.conversations[s.conversationIndex]
	s.current = &conversation
	s.messages = nil
	s.messageCursor = nil
	s.messageIndex, s.messageTop, s.messageOffset = 0, 0, 0
	s.messageAnchor = messageViewAnchor{}
	s.composer.Widget.SetValue("")
	s.composer.Widget.Blur()
	s.composing = false
	s.composerError = ""
	s.submittedText = ""
	s.err = nil
	s.notice = ""

	return tea.Batch(s.loadMessages(ctx, false), s.markRead(ctx, conversation.ID))
}

func (s *Screen) applyConversationsLoaded(ctx *reactea.Ctx, msg conversationsLoadedMsg) {
	if msg.target != s {
		return
	}
	s.loading = false
	s.loadingMore = false
	s.err = msg.err
	if msg.err != nil {
		return
	}
	if msg.more {
		s.conversations = collection.AppendUniqueBy(s.conversations, msg.page.Items, func(conversation messagedomain.MessageConversation) string {
			return conversation.ID
		})
	} else {
		s.conversations = append([]messagedomain.MessageConversation(nil), msg.page.Items...)
		s.conversationIndex, s.conversationTop = 0, 0
	}
	s.conversationCursor = msg.page.NextCursor
	s.requestCount = msg.page.RequestCount
	s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
}

func (s *Screen) applyMessagesLoaded(ctx *reactea.Ctx, msg messagesLoadedMsg) {
	if msg.target != s {
		return
	}
	s.loading = false
	s.loadingMore = false
	s.err = msg.err
	if msg.err != nil {
		if !msg.more {
			s.messageAnchor = messageViewAnchor{}
		}
		return
	}
	if msg.page.Conversation.ID != "" {
		conversation := msg.page.Conversation
		s.current = &conversation
	}
	incoming := append([]messagedomain.Message(nil), msg.page.Items...)
	slices.Reverse(incoming)
	if msg.more {
		selectedID := ""
		if selected := s.selectedMessage(); selected != nil {
			selectedID = selected.ID
		}
		s.messages = collection.AppendUniqueBy(incoming, s.messages, func(message messagedomain.Message) string {
			return message.ID
		})
		if selectedID != "" {
			s.messageIndex = max(0, slices.IndexFunc(s.messages, func(message messagedomain.Message) bool {
				return message.ID == selectedID
			}))
		}
	} else {
		if s.messageAnchor.active {
			s.messages = mergeRefreshedMessages(s.messages, incoming)
		} else {
			s.messages = incoming
		}
		if !s.restoreMessageView() {
			s.messageIndex = max(0, len(s.messages)-1)
			s.messageTop, s.messageOffset = 0, 0
		}
	}
	s.messageCursor = msg.page.NextCursor
	s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
}

func mergeRefreshedMessages(current, refreshed []messagedomain.Message) []messagedomain.Message {
	replacements := make(map[string]messagedomain.Message, len(refreshed))
	for _, message := range refreshed {
		if message.ID != "" {
			replacements[message.ID] = message
		}
	}
	result := make([]messagedomain.Message, 0, len(current)+len(refreshed))
	for _, message := range current {
		if replacement, ok := replacements[message.ID]; ok && message.ID != "" {
			result = append(result, replacement)
			delete(replacements, message.ID)
		} else {
			result = append(result, message)
		}
	}
	for _, message := range refreshed {
		if message.ID == "" {
			result = append(result, message)
			continue
		}
		if _, ok := replacements[message.ID]; !ok {
			continue
		}
		result = append(result, message)
		delete(replacements, message.ID)
	}

	return result
}

func (s *Screen) applyRead(msg readMsg) {
	if msg.target != s {
		return
	}
	if msg.err != nil {
		s.notice = "Could not mark conversation as read: " + ui.FriendlyError(msg.err)

		return
	}
	for index := range s.conversations {
		if s.conversations[index].ID == msg.conversationID {
			s.conversations[index].UnreadCount = 0
		}
	}
}
