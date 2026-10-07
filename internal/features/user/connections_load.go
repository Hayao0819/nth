package user

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/support/collection"
)

type loadedMsg struct {
	target   *Connections
	page     north.UserPage
	response *north.Response
	more     bool
	err      error
}

func (m loadedMsg) Response() *north.Response { return m.response }

func (s *Connections) load(ctx context.Context, more bool) tea.Cmd {
	if s.api == nil || s.loading || more && s.next == nil {
		return nil
	}
	cursor := ""
	if more {
		cursor = *s.next
	}
	handle := strings.TrimPrefix(strings.TrimSpace(s.user.Handle), "@")
	s.loading = true
	s.more = more
	s.err = nil

	return func() tea.Msg {
		var page north.UserPage
		var response *north.Response
		var err error
		if s.following {
			page, response, err = s.api.Following(ctx, handle, cursor)
		} else {
			page, response, err = s.api.Followers(ctx, handle, cursor)
		}

		return loadedMsg{target: s, page: page, response: response, more: more, err: err}
	}
}

func (s *Connections) applyLoaded(msg loadedMsg, width, room int) {
	if msg.target != s {
		return
	}
	s.loading = false
	s.more = false
	s.err = msg.err
	if msg.err != nil {
		return
	}
	if msg.more {
		s.items = collection.AppendUniqueBy(s.items, msg.page.Items, userKey)
	} else {
		s.items = append([]north.User(nil), msg.page.Items...)
		s.selected, s.top = 0, 0
	}
	s.next = msg.page.NextCursor
	s.ensureVisible(width, room)
}

func (s *Connections) loadNearEnd(ctx context.Context) tea.Cmd {
	if len(s.items)-s.selected > 3 {
		return nil
	}

	return s.load(ctx, true)
}

func userKey(user north.User) string {
	return user.ID + "\x00" + user.Handle
}
