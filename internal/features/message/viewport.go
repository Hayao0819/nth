package message

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	messagedomain "github.com/Hayao0819/nth/internal/domain"
)

type messageViewAnchor struct {
	active        bool
	followLatest  bool
	selectedID    string
	topID         string
	selectedIndex int
	topIndex      int
	offset        int
}

func (s *Screen) messageScrollLimit(width, room int) int {
	if s.current == nil || room <= 0 || s.messageIndex < 0 || s.messageIndex >= len(s.messages) {
		return 0
	}

	height := lipgloss.Height(s.renderMessage(s.messages[s.messageIndex], width, true))

	return max(0, height-room)
}

func (s *Screen) scrollMessage(by, width, room int) bool {
	limit := s.messageScrollLimit(width, room)
	before := s.messageOffset
	s.messageOffset = min(max(0, s.messageOffset+by), limit)

	return s.messageOffset != before
}

func (s *Screen) clampMessageOffset(width, room int) {
	s.messageOffset = min(max(0, s.messageOffset), s.messageScrollLimit(width, room))
}

func (s *Screen) visibleMessageCard(card string, room int, selected bool) string {
	if !selected || s.current == nil || room <= 0 {
		return card
	}
	lines := strings.Split(card, "\n")
	offset := min(max(0, s.messageOffset), max(0, len(lines)-room))
	end := min(len(lines), offset+room)

	return strings.Join(lines[offset:end], "\n")
}

func (s *Screen) atLatestMessage(width, room int) bool {
	if len(s.messages) == 0 {
		return true
	}

	return s.messageIndex == len(s.messages)-1 && s.messageOffset >= s.messageScrollLimit(width, room)
}

func (s *Screen) saveMessageView(width, room int) {
	anchor := messageViewAnchor{
		active:        true,
		followLatest:  s.atLatestMessage(width, room),
		selectedIndex: s.messageIndex,
		topIndex:      s.messageTop,
		offset:        s.messageOffset,
	}
	if s.messageIndex >= 0 && s.messageIndex < len(s.messages) {
		anchor.selectedID = s.messages[s.messageIndex].ID
	}
	if s.messageTop >= 0 && s.messageTop < len(s.messages) {
		anchor.topID = s.messages[s.messageTop].ID
	}
	s.messageAnchor = anchor
}

func (s *Screen) restoreMessageView() bool {
	anchor := s.messageAnchor
	s.messageAnchor = messageViewAnchor{}
	if !anchor.active {
		return false
	}
	if anchor.followLatest {
		s.messageIndex = max(0, len(s.messages)-1)
		s.messageTop, s.messageOffset = 0, 0

		return true
	}

	s.messageIndex = min(max(0, anchor.selectedIndex), max(0, len(s.messages)-1))
	if anchor.selectedID != "" {
		if index := slices.IndexFunc(s.messages, func(message messagedomain.Message) bool {
			return message.ID == anchor.selectedID
		}); index >= 0 {
			s.messageIndex = index
		}
	}
	s.messageTop = min(max(0, anchor.topIndex), s.messageIndex)
	if anchor.topID != "" {
		if index := slices.IndexFunc(s.messages, func(message messagedomain.Message) bool {
			return message.ID == anchor.topID
		}); index >= 0 && index <= s.messageIndex {
			s.messageTop = index
		}
	}
	s.messageOffset = anchor.offset

	return true
}
