package feed

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
)

type reactionKind uint8

const (
	reactionLike reactionKind = iota
	reactionRepost
)

type reactionMsg struct {
	target *Feed
	postID string
	kind   reactionKind
	like   north.LikeState
	repost north.RepostState
	resp   *north.Response
	err    error
}

type reactionKey struct {
	postID string
	kind   reactionKind
}

func (f *Feed) ToggleLike(ctx context.Context) tea.Cmd {
	return f.ToggleLikePost(ctx, f.SelectedPost())
}

func (f *Feed) ToggleLikePost(ctx context.Context, post *north.Post) tea.Cmd {
	return f.toggleReaction(ctx, post, reactionLike)
}

func (f *Feed) ToggleRepost(ctx context.Context) tea.Cmd {
	return f.ToggleRepostPost(ctx, f.SelectedPost())
}

func (f *Feed) ToggleRepostPost(ctx context.Context, post *north.Post) tea.Cmd {
	return f.toggleReaction(ctx, post, reactionRepost)
}

func (f *Feed) toggleReaction(ctx context.Context, post *north.Post, kind reactionKind) tea.Cmd {
	target, ok := f.beginReaction(post, kind)
	if !ok {
		return nil
	}

	return func() tea.Msg {
		result := reactionMsg{target: f, postID: target.ID, kind: kind}
		update := postcomponent.ReactionUpdate{PostID: target.ID}

		switch kind {
		case reactionLike:
			if target.Liked {
				result.like, result.resp, result.err = f.api.Unlike(ctx, target.ID)
			} else {
				result.like, result.resp, result.err = f.api.Like(ctx, target.ID)
			}
			update.Action = postcomponent.Like
			update.Like = &result.like
		case reactionRepost:
			if target.Reposted {
				result.repost, result.resp, result.err = f.api.UndoRepost(ctx, target.ID)
			} else {
				result.repost, result.resp, result.err = f.api.Repost(ctx, target.ID)
			}
			update.Action = postcomponent.Repost
			update.Repost = &result.repost
		}
		update.Err = result.err

		return tea.BatchMsg{
			func() tea.Msg { return result },
			func() tea.Msg { return update },
		}
	}
}

func (f *Feed) beginReaction(post *north.Post, kind reactionKind) (*north.Post, bool) {
	if !postcomponent.CanInteract(post) {
		return nil, false
	}
	var notice string
	switch kind {
	case reactionLike:
		notice = "Updating like…"
	case reactionRepost:
		notice = "Updating repost…"
	default:
		return nil, false
	}
	target := post.DisplayPost()
	key := reactionKey{postID: target.ID, kind: kind}
	if _, busy := f.acting[key]; busy {
		return nil, false
	}
	f.acting[key] = struct{}{}
	f.err = nil
	f.notice = notice

	return target, true
}

func (f *Feed) applyReaction(msg reactionMsg) {
	if msg.target != f {
		return
	}
	delete(f.acting, reactionKey{postID: msg.postID, kind: msg.kind})
	f.setResponse(msg.resp)
	if msg.err != nil {
		f.err = msg.err
		f.notice = ""

		return
	}

	f.err = nil
	state := f.localState[msg.postID]
	if msg.kind == reactionLike {
		like := msg.like
		state.like = &like
		f.notice = "Like removed"
		if like.Liked {
			f.notice = "Liked"
		}
	} else {
		repost := msg.repost
		state.repost = &repost
		f.notice = "Repost removed"
		if repost.Reposted {
			f.notice = "Reposted"
		}
	}
	f.localState[msg.postID] = state

	for index := range f.posts {
		post := f.posts[index].DisplayPost()
		if post == nil || post.ID != msg.postID {
			continue
		}
		if msg.kind == reactionLike {
			post.Liked, post.LikeCount = msg.like.Liked, msg.like.LikeCount
		} else {
			post.Reposted, post.RepostCount = msg.repost.Reposted, msg.repost.RepostCount
		}
	}
}
