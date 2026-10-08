package post

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
)

type Editor interface {
	EditablePost(context.Context, string) (north.Post, bool, string, *north.Response, error)
	EditPost(context.Context, domain.PostEdit) (north.Post, *north.Response, error)
}

type EditRequestMsg struct {
	Post north.Post
	ETag string
}

type ManagementUpdate struct {
	Account north.User
}

type editableLoadedMsg struct {
	target   *Screen
	post     north.Post
	eligible bool
	etag     string
	resp     *north.Response
	err      error
}

func (m editableLoadedMsg) Response() *north.Response { return m.resp }

func (d *Screen) loadEditable(ctx context.Context) tea.Cmd {
	if d.editor == nil {
		return nil
	}
	target := d.post.DisplayPost()
	if target == nil {
		return nil
	}
	id := target.ID

	return func() tea.Msg {
		post, eligible, etag, response, err := d.editor.EditablePost(ctx, id)
		return editableLoadedMsg{target: d, post: post, eligible: eligible, etag: etag, resp: response, err: err}
	}
}

func (d *Screen) HandlePostEdit(ctx context.Context, id, text string, editedAt time.Time, err error) tea.Cmd {
	target := d.post.DisplayPost()
	if target == nil || target.ID != id {
		return nil
	}
	if err != nil {
		d.notice = "Update failed · " + ui.FriendlyError(err)
	} else {
		target.Text = text
		target.EditedAt = &editedAt
		d.notice = "Post updated"
	}
	if d.editor == nil {
		return nil
	}
	d.editable = false
	d.editableChecked = false
	d.checkingEditable = true
	d.editETag = ""

	return d.loadEditable(ctx)
}

func ownedBy(post north.Post, account north.User) bool {
	target := post.DisplayPost()
	if target == nil {
		return false
	}
	if account.ID != "" && target.Author.ID != "" {
		return account.ID == target.Author.ID
	}

	return account.Handle != "" && strings.EqualFold(account.Handle, target.Author.Handle)
}
