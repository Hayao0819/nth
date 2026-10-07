package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/support/collection"
)

type membersLoadedMsg struct {
	target   *Members
	page     north.UserPage
	response *north.Response
	more     bool
	err      error
}

func (m membersLoadedMsg) Response() *north.Response { return m.response }

func (s *Members) load(ctx context.Context, more bool) tea.Cmd {
	if s.api == nil || s.loading || s.list.ID == "" || (more && s.next == nil) {
		return nil
	}
	cursor := ""
	if more {
		cursor = *s.next
	}
	s.loading, s.more, s.err = true, more, nil

	return func() tea.Msg {
		page, response, err := s.api.ListMembers(ctx, s.list.ID, cursor)

		return membersLoadedMsg{target: s, page: page, response: response, more: more, err: err}
	}
}

func (s *Members) applyLoaded(msg membersLoadedMsg) {
	if msg.target != s {
		return
	}
	s.loading, s.more, s.err = false, false, msg.err
	if msg.err != nil {
		return
	}
	if msg.more {
		s.users = collection.AppendUniqueBy(s.users, msg.page.Items, userKey)
	} else {
		s.users = append([]north.User(nil), msg.page.Items...)
		s.selected, s.top = 0, 0
	}
	s.next = msg.page.NextCursor
}

func userKey(user north.User) string {
	if user.ID != "" {
		return user.ID
	}

	return user.Handle
}
