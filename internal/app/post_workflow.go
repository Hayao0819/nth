package app

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	postpage "github.com/Hayao0819/nth/internal/pages/post"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (r *root) compose(ctx *reactea.Ctx, replyTo, quote *north.Post) tea.Cmd {
	if r.posting {
		r.notice = "A post is already being sent"

		return nil
	}

	key := composeDraftKey(postID(replyTo), postID(quote))

	return modal.PushAt(ctx, dialog.NewCompose(r.theme, replyTo, quote, r.drafts[key]), dialog.Placement(ctx, 72, 16))
}

func (r *root) edit(ctx *reactea.Ctx, post north.Post) tea.Cmd {
	if r.editing {
		r.notice = "A post is already being updated"

		return nil
	}
	target := post.DisplayPost()
	if target == nil {
		r.notice = "This post cannot be edited"

		return nil
	}

	return modal.PushAt(
		ctx,
		dialog.NewEdit(r.theme, *target, r.drafts["edit:"+target.ID]),
		dialog.Placement(ctx, 72, 16),
	)
}

func (r *root) submit(ctx context.Context, post dialog.Submission) tea.Cmd {
	if r.posting {
		return nil
	}
	r.posting = true
	r.notice = "Posting…"
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""
	draftKey := composeDraftKey(post.ReplyTo, post.QuoteID)

	req := north.CreatePostRequest{Text: post.Text, QuotePostID: post.QuoteID}
	if post.ReplyTo != "" {
		req.Reply = &north.CreatePostReply{InReplyToPostID: post.ReplyTo}
	}

	return func() tea.Msg {
		_, resp, err := r.api.CreatePost(ctx, req)

		return postCreatedMsg{target: r, draftKey: draftKey, resp: resp, err: err}
	}
}

func (r *root) handlePostAction(ctx *reactea.Ctx, action postcomponent.Action, post north.Post) tea.Cmd {
	if action == postcomponent.ViewPost {
		return r.openPost(ctx, post)
	}
	if action == postcomponent.ViewAuthor {
		target := post.DisplayPost()
		if target == nil {
			r.notice = "This account is unavailable"

			return nil
		}

		return r.openUser(ctx, target.Author)
	}
	if !postcomponent.CanInteract(&post) {
		r.notice = "This post cannot be acted on"

		return nil
	}
	target := post.DisplayPost()
	switch action {
	case postcomponent.Reply:
		return r.compose(ctx, target, nil)
	case postcomponent.Repost:
		return r.feed.ToggleRepostPost(ctx.Context(), &post)
	case postcomponent.Like:
		return r.feed.ToggleLikePost(ctx.Context(), &post)
	case postcomponent.Quote:
		return r.compose(ctx, nil, target)
	case postcomponent.Edit:
		if !r.ownsPost(post) {
			r.notice = "Only your own posts can be edited"

			return nil
		}
		if _, ok := r.api.(postpage.Editor); !ok {
			r.notice = "Editing is not available with the current authentication"

			return nil
		}

		return r.edit(ctx, post)
	case postcomponent.Delete:
		if !r.ownsPost(post) {
			r.notice = "Only your own posts can be deleted"

			return nil
		}

		return r.deletePost(ctx.Context(), target.ID)
	default:
		return nil
	}
}

func (r *root) editPost(ctx context.Context, id, text string, mediaIDs []string) tea.Cmd {
	if r.editing {
		return nil
	}
	editor, ok := r.api.(postpage.Editor)
	if !ok {
		r.notice = "Editing is not available with the current authentication"

		return nil
	}
	r.editing = true
	r.notice = "Updating post…"
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""

	return func() tea.Msg {
		response, err := editor.EditPost(ctx, id, text, mediaIDs)

		return postEditedMsg{target: r, postID: id, text: text, resp: response, err: err}
	}
}

func (r *root) deletePost(ctx context.Context, id string) tea.Cmd {
	if r.deleting {
		r.notice = "A post is already being deleted"

		return nil
	}
	r.deleting = true
	r.notice = "Deleting…"
	r.problem = nil
	r.postFailed = false
	r.failureJob = ""

	return func() tea.Msg {
		deleted, resp, err := r.api.DeletePost(ctx, id)

		return postDeletedMsg{target: r, postID: id, deleted: deleted, resp: resp, err: err}
	}
}

func (r *root) ownsPost(post north.Post) bool {
	if r.me == nil {
		return false
	}
	target := post.DisplayPost()
	if target == nil {
		return false
	}
	if r.me.ID != "" && target.Author.ID != "" {
		return r.me.ID == target.Author.ID
	}

	return strings.EqualFold(r.me.Handle, target.Author.Handle)
}

func (r *root) rememberDraft(key, text string) {
	if strings.TrimSpace(text) == "" {
		delete(r.drafts, key)

		return
	}
	r.drafts[key] = text
}

func composeDraftKey(replyTo, quoteID string) string {
	switch {
	case replyTo != "":
		return "reply:" + replyTo
	case quoteID != "":
		return "quote:" + quoteID
	default:
		return "post"
	}
}

func submissionDraftKey(submission dialog.Submission) string {
	if submission.EditID != "" {
		return "edit:" + submission.EditID
	}

	return composeDraftKey(submission.ReplyTo, submission.QuoteID)
}

func postID(post *north.Post) string {
	if post == nil {
		return ""
	}

	return post.ID
}
