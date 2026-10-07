package saved

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
)

type loadedMsg struct {
	target    *Screen
	tab       tab
	drafts    []north.Draft
	scheduled []north.ScheduledPost
	pending   []north.PendingPost
	response  *north.Response
	err       error
}

func (m loadedMsg) Response() *north.Response { return m.response }

func (s *Screen) load(ctx context.Context) tea.Cmd {
	if s.api == nil || s.loading {
		return nil
	}
	s.loading = true
	s.err = nil
	current := s.tab

	return func() tea.Msg {
		message := loadedMsg{target: s, tab: current}
		switch current {
		case scheduledTab:
			message.scheduled, message.response, message.err = s.api.ScheduledPosts(ctx, "")
		case pendingTab:
			message.pending, message.response, message.err = s.api.PendingPosts(ctx, "")
		default:
			message.drafts, message.response, message.err = s.api.Drafts(ctx, "")
		}

		return message
	}
}

func (s *Screen) applyLoaded(msg loadedMsg) {
	if msg.target != s || msg.tab != s.tab {
		return
	}
	s.loading = false
	s.err = msg.err
	if msg.err != nil {
		return
	}
	s.drafts = append([]north.Draft(nil), msg.drafts...)
	s.scheduled = append([]north.ScheduledPost(nil), msg.scheduled...)
	s.pending = append([]north.PendingPost(nil), msg.pending...)
	s.selected, s.top = 0, 0
}
