package post

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/support/collection"
)

type loadedMsg struct {
	target   *Activity
	posts    []north.Post
	users    []north.User
	versions []north.PostVersion
	next     *string
	response *north.Response
	more     bool
	err      error
}

func (m loadedMsg) Response() *north.Response { return m.response }

func (s *Activity) load(ctx context.Context, more bool) tea.Cmd {
	if s.api == nil || s.loading || more && s.next == nil {
		return nil
	}
	id := postID(s.post)
	if id == "" {
		return nil
	}
	cursor := ""
	if more {
		cursor = *s.next
	}
	s.loading = true
	s.more = more
	s.err = nil

	return func() tea.Msg {
		result := loadedMsg{target: s, more: more}
		switch s.kind {
		case navigation.PostQuotes:
			page, response, err := s.api.Quotes(ctx, id, cursor)
			result.posts, result.next, result.response, result.err = page.Items, page.NextCursor, response, err
		case navigation.PostLikes:
			page, response, err := s.api.PostLikes(ctx, id, cursor)
			result.users, result.next, result.response, result.err = page.Items, page.NextCursor, response, err
		case navigation.PostReposts:
			page, response, err := s.api.PostReposts(ctx, id, cursor)
			result.users, result.next, result.response, result.err = page.Items, page.NextCursor, response, err
		case navigation.PostHistory:
			history, response, err := s.api.PostEditHistory(ctx, id)
			result.versions, result.response, result.err = history.Versions, response, err
		}

		return result
	}
}

func (s *Activity) applyLoaded(msg loadedMsg) {
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
		s.posts = collection.AppendUniqueBy(s.posts, msg.posts, postID)
		s.users = collection.AppendUniqueBy(s.users, msg.users, userID)
	} else {
		s.posts = append([]north.Post(nil), msg.posts...)
		s.users = append([]north.User(nil), msg.users...)
		s.versions = append([]north.PostVersion(nil), msg.versions...)
		s.selected, s.top = 0, 0
	}
	s.next = msg.next
}

func postID(post north.Post) string {
	if target := post.DisplayPost(); target != nil {
		return target.ID
	}

	return post.ID
}

func userID(user north.User) string {
	if user.ID != "" {
		return user.ID
	}

	return user.Handle
}
