package postpage

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (d *Screen) actions() []postcomponent.Action {
	actions := []postcomponent.Action{
		postcomponent.Reply,
		postcomponent.Repost,
		postcomponent.Like,
		postcomponent.Quote,
	}
	if d.manage {
		if d.editable {
			actions = append(actions, postcomponent.Edit)
		}
		actions = append(actions, postcomponent.Delete)
	}

	return actions
}

func (d *Screen) request(ctx *reactea.Ctx, action postcomponent.Action) tea.Cmd {
	if action == postcomponent.ViewAuthor {
		target := d.post.DisplayPost()
		if target == nil {
			return nil
		}

		return d.requestUser(ctx, target.Author)
	}
	if !postcomponent.CanInteract(&d.post) {
		return nil
	}
	if action == postcomponent.Edit {
		if !d.manage || !d.editable {
			return nil
		}

		return postcomponent.Request(postcomponent.Edit, d.post)
	}
	if action == postcomponent.Delete {
		if !d.manage || d.confirmingDelete {
			return nil
		}
		d.confirmingDelete = true

		return modal.PushAt(
			ctx,
			dialog.NewConfirm(d.theme, "Delete post?", "This cannot be undone.", "Delete"),
			dialog.Placement(ctx, 48, 7),
		)
	}

	if action == postcomponent.Like || action == postcomponent.Repost {
		if d.acting[action] {
			return nil
		}
		d.acting[action] = true
		d.notice = "Updating…"
	}

	return postcomponent.Request(action, d.post)
}

func (d *Screen) requestSelected(ctx *reactea.Ctx, action postcomponent.Action) tea.Cmd {
	if d.replyFocused {
		if reply := d.selectedReply(); reply != nil {
			return postcomponent.Request(action, *reply)
		}
	}

	return d.request(ctx, action)
}

func (d *Screen) requestUser(_ *reactea.Ctx, user north.User) tea.Cmd {
	return postcomponent.Request(postcomponent.ViewAuthor, north.Post{Author: user})
}

func (d *Screen) requestNextUser(ctx *reactea.Ctx) tea.Cmd {
	users := postcomponent.LinkedUsers(d.post, d.ancestors...)
	if len(users) == 0 {
		return nil
	}
	if len(users) > 1 {
		d.linkedUser = (d.linkedUser + 1) % len(users)
	} else {
		d.linkedUser = 0
	}

	return d.requestUser(ctx, users[d.linkedUser])
}

func (d *Screen) UpdatePost(id, text string, editedAt time.Time) {
	target := d.post.DisplayPost()
	if target == nil || target.ID != id {
		return
	}
	target.Text = text
	target.EditedAt = &editedAt
}
