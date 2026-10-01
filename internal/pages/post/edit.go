package postpage

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
)

type Editor interface {
	EditablePost(context.Context, string) (north.Post, bool, *north.Response, error)
	EditPost(context.Context, string, string, []string) (*north.Response, error)
}

type ManagementUpdate struct {
	Account north.User
}

type editableLoadedMsg struct {
	target   *Screen
	post     north.Post
	eligible bool
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
		post, eligible, response, err := d.editor.EditablePost(ctx, id)
		return editableLoadedMsg{target: d, post: post, eligible: eligible, resp: response, err: err}
	}
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
