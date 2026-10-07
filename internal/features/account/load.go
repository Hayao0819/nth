package account

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/support/collection"
)

type loadedMsg struct {
	target   *Screen
	tab      tab
	users    north.UserPage
	words    []north.MutedKeyword
	response *north.Response
	more     bool
	err      error
}

func (m loadedMsg) Response() *north.Response { return m.response }

func (s *Screen) load(ctx context.Context, more bool) tea.Cmd {
	if s.api == nil || s.loading || (more && s.next == nil) || (s.tab == keywordsTab && s.keywords == nil) {
		return nil
	}
	cursor := ""
	if more {
		cursor = *s.next
	}
	s.loading, s.more, s.err = true, more, nil
	current := s.tab

	return func() tea.Msg {
		message := loadedMsg{target: s, tab: current, more: more}
		switch current {
		case blockedTab:
			message.users, message.response, message.err = s.api.BlockedUsers(ctx, cursor)
		case mutedTab:
			message.users, message.response, message.err = s.api.MutedUsers(ctx, cursor)
		case keywordsTab:
			message.words, message.response, message.err = s.keywords.MutedKeywords(ctx, cursor)
		default:
			message.users, message.response, message.err = s.api.FollowRequests(ctx, cursor)
		}

		return message
	}
}

func (s *Screen) applyLoaded(msg loadedMsg) {
	if msg.target != s || msg.tab != s.tab {
		return
	}
	s.loading, s.more, s.err = false, false, msg.err
	if msg.err != nil {
		return
	}
	if msg.tab == keywordsTab {
		s.words = append([]north.MutedKeyword(nil), msg.words...)
		s.next = nil
	} else if msg.more {
		s.users = collection.AppendUniqueBy(s.users, msg.users.Items, userKey)
		s.next = msg.users.NextCursor
	} else {
		s.users = append([]north.User(nil), msg.users.Items...)
		s.next = msg.users.NextCursor
	}
	if !msg.more {
		s.selected, s.top = 0, 0
	}
}

func userKey(user north.User) string {
	if user.ID != "" {
		return user.ID
	}

	return user.Handle
}
