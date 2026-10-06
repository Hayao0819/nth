package message

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/listview"
	"github.com/Hayao0819/nth/internal/components/pageheader"
)

func (s *Screen) back() tea.Cmd {
	if s.current != nil {
		s.current = nil
		s.messages = nil
		s.messageCursor = nil
		s.messageIndex, s.messageTop = 0, 0
		s.err = nil
		s.notice = ""

		return nil
	}

	return pageheader.Back()
}

func (s *Screen) loadNearEnd(ctx context.Context) tea.Cmd {
	if s.itemCount()-s.index() > 3 {
		return nil
	}
	if s.current == nil {
		return s.loadConversations(ctx, true)
	}

	return s.loadMessages(ctx, true)
}

func (s *Screen) move(by, width, room int) {
	if s.itemCount() == 0 {
		return
	}
	s.setIndex(listview.Move(s.index(), s.itemCount(), by))
	s.ensureVisible(width, room)
}

func (s *Screen) ensureVisible(width, room int) {
	s.setTop(listview.EnsureVisible(s.top(), s.index(), s.itemCount(), room, func(index int) int {
		return lipgloss.Height(s.renderItem(index, width, index == s.index()))
	}))
}

func (s *Screen) itemPositionAt(row, width, room int) (int, int, bool) {
	return listview.ItemPositionAt(s.top(), s.itemCount(), row, room, func(index int) int {
		return lipgloss.Height(s.renderItem(index, width, index == s.index()))
	})
}

func (s *Screen) itemCount() int {
	if s.current == nil {
		return len(s.conversations)
	}

	return len(s.messages)
}

func (s *Screen) index() int {
	if s.current == nil {
		return s.conversationIndex
	}

	return s.messageIndex
}

func (s *Screen) setIndex(index int) {
	if s.current == nil {
		s.conversationIndex = index
	} else {
		s.messageIndex = index
	}
}

func (s *Screen) top() int {
	if s.current == nil {
		return s.conversationTop
	}

	return s.messageTop
}

func (s *Screen) setTop(top int) {
	if s.current == nil {
		s.conversationTop = top
	} else {
		s.messageTop = top
	}
}

func (s *Screen) hasMore() bool {
	if s.current == nil {
		return s.conversationCursor != nil
	}

	return s.messageCursor != nil
}

func (s *Screen) selectedSender() (north.User, bool) {
	if s.messageIndex < 0 || s.messageIndex >= len(s.messages) {
		return north.User{}, false
	}
	user := s.messages[s.messageIndex].Sender

	return user, strings.TrimSpace(user.Handle) != ""
}

func (s *Screen) room(height int) int {
	return max(0, height-pageheader.Height-1)
}
