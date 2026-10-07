package post

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
)

type Action uint8

const (
	Reply Action = iota
	Repost
	Like
	Quote
	Bookmark
	Edit
	Delete
	ViewAuthor
	ViewPost
)

type ActionMsg struct {
	Action Action
	Post   north.Post
}

type ReactionUpdate struct {
	PostID   string
	Action   Action
	Like     *north.LikeState
	Repost   *north.RepostState
	Bookmark *bool
	Err      error
}

func Request(action Action, post north.Post) tea.Cmd {
	return func() tea.Msg { return ActionMsg{Action: action, Post: post} }
}

func CanInteract(post *north.Post) bool {
	target := post.DisplayPost()

	return target != nil && !Concealed(target)
}

func CardActionAt(x, width int) Action {
	usable := max(1, width-4)
	switch ui.ColumnAt(x-4, usable, 5) {
	case 0:
		return Reply
	case 1:
		return Repost
	case 2:
		return Like
	case 3:
		return Quote
	default:
		return Bookmark
	}
}
