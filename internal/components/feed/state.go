package feed

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type localPostState struct {
	like     *north.LikeState
	repost   *north.RepostState
	bookmark *bool
	text     *string
	editedAt *time.Time
	media    *[]north.Media
}

func (f *Feed) SetBookmarked(id string, bookmarked bool) {
	state := f.localState[id]
	state.bookmark = &bookmarked
	f.localState[id] = state
	for index := range f.posts {
		target := f.posts[index].DisplayPost()
		if target != nil && target.ID == id {
			target.Bookmarked = bookmarked
		}
	}
}

func (f *Feed) SetMode(ctx *reactea.Ctx, mode Mode, query string) tea.Cmd {
	// Invalidate a request from the previous mode before starting the next one.
	f.seq++
	f.loading = false
	f.loadingMore = false
	f.refreshReady = true
	f.mode, f.query = mode, query
	f.posts = nil
	f.nextCursor = nil
	f.selected, f.top = 0, 0
	f.err, f.notice = nil, ""

	return f.load(ctx, false)
}

// ClearSearch invalidates pending searches and clears the result list.
func (f *Feed) ClearSearch(query string) {
	f.seq++
	f.loading = false
	f.loadingMore = false
	f.refreshReady = true
	f.mode, f.query = Search, query
	f.posts = nil
	f.nextCursor = nil
	f.selected, f.top = 0, 0
	f.err, f.notice = nil, ""
}

func (f *Feed) Refresh(ctx *reactea.Ctx) tea.Cmd { return f.load(ctx, false) }

func (f *Feed) Clear() {
	f.seq++
	f.loading = false
	f.loadingMore = false
	f.posts = nil
	f.nextCursor = nil
	f.selected, f.top = 0, 0
	f.err, f.notice = nil, ""
}

func (f *Feed) SelectedPost() *north.Post {
	if f.selected < 0 || f.selected >= len(f.posts) {
		return nil
	}

	return &f.posts[f.selected]
}

func (f *Feed) RemovePost(ctx *reactea.Ctx, id string) {
	f.removed[id] = struct{}{}
	posts := f.posts[:0]
	for index := range f.posts {
		post := f.posts[index]
		target := post.DisplayPost()
		if post.ID == id || target != nil && target.ID == id {
			continue
		}
		posts = append(posts, post)
	}
	f.posts = posts
	f.clampSelection()
	f.ensureVisible(ctx.Width(), ctx.Height())
}

func (f *Feed) UpdatePost(ctx *reactea.Ctx, id, text string, editedAt time.Time) {
	state := f.localState[id]
	state.text = &text
	state.editedAt = &editedAt
	f.localState[id] = state
	for index := range f.posts {
		target := f.posts[index].DisplayPost()
		if target == nil || target.ID != id {
			continue
		}
		target.Text = text
		target.EditedAt = &editedAt
	}
	f.ensureVisible(ctx.Width(), ctx.Height())
}

func (f *Feed) ReplacePost(ctx *reactea.Ctx, id string, post north.Post) {
	updated := post.DisplayPost()
	if updated == nil || updated.ID != id {
		return
	}
	text := updated.Text
	media := append([]north.Media(nil), updated.Media...)
	state := f.localState[id]
	state.text = &text
	if updated.EditedAt != nil {
		editedAt := *updated.EditedAt
		state.editedAt = &editedAt
	}
	state.media = &media
	f.localState[id] = state
	for index := range f.posts {
		target := f.posts[index].DisplayPost()
		if target == nil || target.ID != id {
			continue
		}
		*target = *updated
		target.Media = append([]north.Media(nil), updated.Media...)
	}
	f.ensureVisible(ctx.Width(), ctx.Height())
}

func (f *Feed) CurrentMode() Mode { return f.mode }

func (f *Feed) SearchQuery() string { return f.query }

func (f *Feed) Status() string {
	switch {
	case f.loading:
		return "Loading…"
	case f.err != nil:
		return ui.FriendlyError(f.err)
	default:
		return f.notice
	}
}

func (f *Feed) Progress() string {
	if len(f.posts) == 0 {
		return ""
	}

	progress := fmt.Sprintf("%d of %d", f.selected+1, len(f.posts))
	switch {
	case f.loadingMore:
		return progress + " · loading more…"
	case f.nextCursor != nil:
		return progress + " · more below"
	default:
		return progress + " · end"
	}
}

func (f *Feed) Failed() bool { return f.err != nil }

func (f *Feed) LastResponse() *north.Response { return f.resp }

func (f *Feed) LastResponseAt() time.Time { return f.respAt }

func (f *Feed) setResponse(response *north.Response) {
	if response == nil {
		return
	}
	f.resp = response
	f.respAt = time.Now()
}

func (f *Feed) applyLocalState(posts []north.Post) []north.Post {
	result := make([]north.Post, 0, len(posts))
	for _, post := range posts {
		target := post.DisplayPost()
		if target == nil {
			result = append(result, post)
			continue
		}
		if _, removed := f.removed[post.ID]; removed {
			continue
		}
		if _, removed := f.removed[target.ID]; removed {
			continue
		}
		state, ok := f.localState[target.ID]
		if ok {
			if post.RepostOf != nil {
				clone := *post.RepostOf
				post.RepostOf = &clone
				target = post.RepostOf
			}
			if state.like != nil {
				target.Liked = state.like.Liked
				target.LikeCount = state.like.LikeCount
			}
			if state.repost != nil {
				target.Reposted = state.repost.Reposted
				target.RepostCount = state.repost.RepostCount
			}
			if state.bookmark != nil {
				target.Bookmarked = *state.bookmark
			}
			if state.text != nil {
				target.Text = *state.text
			}
			if state.editedAt != nil {
				edited := *state.editedAt
				target.EditedAt = &edited
			}
			if state.media != nil {
				target.Media = append([]north.Media(nil), (*state.media)...)
			}
		}
		result = append(result, post)
	}

	return result
}
