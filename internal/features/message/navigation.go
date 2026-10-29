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
	s.composing = false
	s.composer.Widget.Blur()
	s.composer.Widget.SetValue("")
	s.composerError = ""
	s.submittedText = ""
	release := s.inputCapture.ReleaseInput()
	if s.current != nil {
		s.current = nil
		s.messages = nil
		s.messageCursor = nil
		s.messageIndex, s.messageTop, s.messageOffset = 0, 0, 0
		s.messageAnchor = messageViewAnchor{}
		s.err = nil
		s.notice = ""

		return release
	}

	return tea.Batch(release, pageheader.Back())
}

func (s *Screen) toggleRequests(ctx context.Context) tea.Cmd {
	if s.request == nil || s.current != nil || s.loading {
		return nil
	}
	s.requests = !s.requests
	s.conversations = nil
	s.conversationCursor = nil
	s.conversationIndex, s.conversationTop = 0, 0
	s.notice = ""
	s.err = nil

	return s.loadConversations(ctx, false)
}

func (s *Screen) loadNearEdge(ctx context.Context) tea.Cmd {
	if s.current == nil {
		if s.itemCount()-s.index() > 3 {
			return nil
		}

		return s.loadConversations(ctx, true)
	}
	if s.index() > 2 {
		return nil
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
	s.clampMessageOffset(width, room)
}

func (s *Screen) itemPositionAt(row, width, room int) (int, int, bool) {
	row -= s.bottomPadding(width, room)
	if row < 0 || row >= room {
		return 0, 0, false
	}
	if s.current != nil && s.itemCount() > 0 && s.top() == s.index() {
		height := lipgloss.Height(s.renderItem(s.index(), width, true))
		if height > room {
			return s.index(), row + s.messageOffset, true
		}
	}

	return listview.ItemPositionAt(s.top(), s.itemCount(), row, room, func(index int) int {
		return lipgloss.Height(s.renderItem(index, width, index == s.index()))
	})
}

func (s *Screen) bottomPadding(width, room int) int {
	if s.current == nil || s.loadingMore || s.index() != s.itemCount()-1 {
		return 0
	}
	used := 0
	for index := s.top(); index < s.itemCount(); index++ {
		height := lipgloss.Height(s.renderItem(index, width, index == s.index()))
		if used+height > room {
			if used == 0 {
				return 0
			}
			break
		}
		used += height
	}

	return max(0, room-used)
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
		if s.messageIndex != index {
			s.messageOffset = 0
		}
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
	footerHeight := 1
	if s.composerVisible() {
		footerHeight = composerHeight
	}

	return max(0, height-pageheader.Height-footerHeight)
}
