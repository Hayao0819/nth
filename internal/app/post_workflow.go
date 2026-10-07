package app

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	postfeature "github.com/Hayao0819/nth/internal/features/post"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (r *root) compose(ctx *reactea.Ctx, replyTo, quote *north.Post) tea.Cmd {
	if r.posting {
		r.notice = "A post is already being sent"

		return nil
	}

	key := composeDraftKey(postID(replyTo), postID(quote))
	draft := r.drafts[key]
	composer := dialog.NewCompose(r.theme, replyTo, quote, draft.Text).
		SetMediaAPI(r.media).
		SetThreadEnabled(r.threads != nil).
		RestoreMediaIDs(draft.MediaIDs).
		RestorePoll(draft.Poll).
		RestoreThread(draft.ThreadItems)

	return modal.PushAt(ctx, composer, dialog.Placement(ctx, 72, 16))
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

	key := "edit:" + target.ID
	draft := r.drafts[key]
	composer := dialog.NewEdit(r.theme, *target, draft.Text).
		SetMediaAPI(r.media).
		RestoreMediaIDs(draft.MediaIDs)

	return modal.PushAt(ctx, composer, dialog.Placement(ctx, 72, 16))
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

	if len(post.ThreadItems) > 0 {
		if r.threads == nil {
			r.posting = false
			r.notice = "Thread creation is not available with the current authentication"

			return nil
		}

		return func() tea.Msg {
			_, resp, err := r.threads.CreateThread(ctx, post.ThreadItems)

			return postCreatedMsg{target: r, draftKey: draftKey, resp: resp, err: err}
		}
	}

	req := north.CreatePostRequest{Text: post.Text, QuotePostID: post.QuoteID, Poll: clonePoll(post.Poll)}
	if len(post.MediaIDs) > 0 {
		req.Media = &north.CreatePostMedia{MediaIDs: append([]string(nil), post.MediaIDs...)}
	}
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
	case postcomponent.Bookmark:
		return r.toggleBookmark(ctx.Context(), target)
	case postcomponent.Edit:
		if !r.ownsPost(post) {
			r.notice = "Only your own posts can be edited"

			return nil
		}
		if _, ok := r.api.(postfeature.Editor); !ok {
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

func (r *root) toggleBookmark(ctx context.Context, post *north.Post) tea.Cmd {
	if r.bookmarks == nil || post == nil || strings.TrimSpace(post.ID) == "" {
		r.notice = "Bookmarks are not available with the current authentication"

		return nil
	}
	if _, busy := r.bookmarking[post.ID]; busy {
		return nil
	}
	r.bookmarking[post.ID] = struct{}{}
	r.notice = "Updating bookmark…"
	r.problem = nil
	id := post.ID
	bookmarked := !post.Bookmarked

	return func() tea.Msg {
		var response *north.Response
		var err error
		if bookmarked {
			_, response, err = r.bookmarks.Bookmark(ctx, id)
		} else {
			_, response, err = r.bookmarks.Unbookmark(ctx, id)
		}

		return bookmarkChangedMsg{target: r, postID: id, bookmarked: bookmarked, response: response, err: err}
	}
}

func (r *root) editPost(ctx context.Context, id, text string, mediaIDs []string) tea.Cmd {
	if r.editing {
		return nil
	}
	editor, ok := r.api.(postfeature.Editor)
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

func (r *root) rememberDraft(key string, draft dialog.Submission) {
	if strings.TrimSpace(draft.Text) == "" && len(draft.MediaIDs) == 0 && draft.Poll == nil && len(draft.ThreadItems) == 0 {
		delete(r.drafts, key)

		return
	}
	draft.MediaIDs = append([]string(nil), draft.MediaIDs...)
	draft.Poll = clonePoll(draft.Poll)
	draft.ThreadItems = cloneThreadItems(draft.ThreadItems)
	r.drafts[key] = draft
}

func clonePoll(poll *north.CreatePoll) *north.CreatePoll {
	if poll == nil {
		return nil
	}

	return &north.CreatePoll{Options: append([]string(nil), poll.Options...), DurationMinutes: poll.DurationMinutes}
}

func cloneThreadItems(items []north.ThreadItem) []north.ThreadItem {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]north.ThreadItem, len(items))
	for index, item := range items {
		cloned[index] = item
		cloned[index].MediaIDs = append([]string(nil), item.MediaIDs...)
		cloned[index].Poll = clonePoll(item.Poll)
	}

	return cloned
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
