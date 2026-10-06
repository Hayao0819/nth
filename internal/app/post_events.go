package app

import (
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type postCreatedMsg struct {
	target   *root
	draftKey string
	resp     *north.Response
	err      error
}

type postDeletedMsg struct {
	target  *root
	postID  string
	deleted bool
	resp    *north.Response
	err     error
}

type postEditedMsg struct {
	target *root
	postID string
	text   string
	resp   *north.Response
	err    error
}

type postUpdateTarget interface {
	UpdatePost(string, string, time.Time)
}

func (r *root) handlePostCreated(ctx *reactea.Ctx, msg postCreatedMsg) tea.Cmd {
	if msg.target != r {
		return nil
	}
	r.posting = false
	r.setResponse(msg.resp)
	if msg.err != nil {
		r.problem, r.notice = msg.err, ""
		r.postFailed = true
		r.failureJob = "Post"

		return nil
	}
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""
	r.notice = "Post sent"
	delete(r.drafts, msg.draftKey)

	return r.feed.Refresh(ctx)
}

func (r *root) handlePostDeleted(ctx *reactea.Ctx, msg postDeletedMsg) tea.Cmd {
	if msg.target != r {
		return nil
	}
	r.deleting = false
	r.setResponse(msg.resp)
	if msg.err != nil {
		r.problem, r.notice = msg.err, ""

		return nil
	}
	if !msg.deleted {
		r.problem = errors.New("north did not delete the post")
		r.notice = ""

		return nil
	}
	r.problem = nil
	r.notice = "Post deleted"
	r.feed.RemovePost(ctx, msg.postID)
	delete(r.posts, msg.postID)
	if r.page.kind == postPage && r.page.key == msg.postID {
		return r.backPage(ctx)
	}

	return nil
}

func (r *root) handlePostEdited(ctx *reactea.Ctx, msg postEditedMsg) tea.Cmd {
	if msg.target != r {
		return nil
	}
	r.editing = false
	r.setResponse(msg.resp)
	if msg.err != nil {
		r.problem, r.notice = msg.err, ""
		r.postFailed = true
		r.failureJob = "Update"

		return nil
	}
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""
	r.notice = "Post updated"
	delete(r.drafts, "edit:"+msg.postID)
	editedAt := time.Now()
	r.feed.UpdatePost(ctx, msg.postID, msg.text, editedAt)
	if cached, ok := r.posts[msg.postID]; ok {
		if target := cached.DisplayPost(); target != nil {
			target.Text = msg.text
			target.EditedAt = &editedAt
			r.posts[msg.postID] = cached
		}
	}
	if detail, ok := r.currentPage().(postUpdateTarget); ok {
		detail.UpdatePost(msg.postID, msg.text, editedAt)
	}

	return nil
}

func (r *root) handleSubmission(ctx *reactea.Ctx, result modal.Result[dialog.Submission]) tea.Cmd {
	if !result.Ok() {
		return nil
	}

	submission := result.Value
	key := submissionDraftKey(submission)
	r.rememberDraft(key, submission.Text)
	if submission.Canceled {
		if strings.TrimSpace(submission.Text) != "" {
			r.notice = "Draft saved"
		}

		return nil
	}
	if submission.EditID != "" {
		return r.editPost(ctx.Context(), submission.EditID, submission.Text, submission.MediaIDs)
	}

	return r.submit(ctx.Context(), submission)
}
