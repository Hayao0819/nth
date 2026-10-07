package feed

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/support/collection"
	"github.com/Hayao0819/reactea/v2"
)

type feedLoadedMsg struct {
	target *Feed
	seq    uint64
	more   bool
	fresh  bool
	page   north.PostPage
	resp   *north.Response
	err    error
}

func (f *Feed) load(ctx *reactea.Ctx, more bool) tea.Cmd {
	return f.loadPage(ctx, more, false)
}

func (f *Feed) loadLatest(ctx *reactea.Ctx) tea.Cmd {
	return f.loadPage(ctx, false, true)
}

func (f *Feed) loadPage(ctx *reactea.Ctx, more, fresh bool) tea.Cmd {
	if f.loading && more {
		return nil
	}

	f.seq++
	seq := f.seq
	f.loading = true
	f.loadingMore = more
	f.err = nil
	f.notice = ""

	mode, query, loader := f.mode, f.query, f.loader
	cursor := ""
	if more && f.nextCursor != nil {
		cursor = *f.nextCursor
	}
	reqCtx := ctx.Context()

	return func() tea.Msg {
		var (
			page north.PostPage
			resp *north.Response
			err  error
		)

		if loader != nil {
			page, resp, err = loader(reqCtx, cursor)
		} else {
			switch mode {
			case Ranked:
				page, resp, err = f.api.HomeTimeline(reqCtx, north.TimelineOptions{Ranked: true, Cursor: cursor})
			case Search:
				page, resp, err = f.api.SearchPosts(reqCtx, query, north.SearchOptions{Tab: north.SearchLatest, Cursor: cursor})
			default:
				page, resp, err = f.api.HomeTimeline(reqCtx, north.TimelineOptions{Cursor: cursor})
			}
		}

		return feedLoadedMsg{target: f, seq: seq, more: more, fresh: fresh, page: page, resp: resp, err: err}
	}
}

func (f *Feed) applyLoadedPage(ctx *reactea.Ctx, msg feedLoadedMsg) tea.Cmd {
	if msg.target != f || msg.seq != f.seq {
		return nil
	}

	f.loading = false
	f.loadingMore = false
	f.setResponse(msg.resp)
	if msg.err != nil {
		f.err = msg.err
		f.notice = ""
		f.refreshReady = true

		return nil
	}

	f.err = nil
	msg.page.Items = f.applyLocalState(msg.page.Items)
	hadPosts := len(f.posts) > 0
	added := len(msg.page.Items)
	switch {
	case msg.more:
		before := len(f.posts)
		f.posts = collection.AppendUniqueBy(f.posts, msg.page.Items, func(post north.Post) string { return post.ID })
		added = len(f.posts) - before
		f.fillLoads++
		f.notice = loadedNotice(added, "older")
	case msg.fresh:
		before := len(f.posts)
		f.posts = collection.AppendUniqueBy(msg.page.Items, f.posts, func(post north.Post) string { return post.ID })
		added = len(f.posts) - before
		f.selected, f.top = 0, 0
		f.fillLoads = 0
		f.notice = "Up to date"
		if added > 0 {
			f.notice = loadedNotice(added, "new")
		}
	default:
		f.posts = append([]north.Post(nil), msg.page.Items...)
		f.selected, f.top = 0, 0
		f.fillLoads = 0
		f.notice = "Up to date"
	}
	if !msg.fresh || !hadPosts {
		f.nextCursor = msg.page.NextCursor
	}
	f.clampSelection()
	f.ensureVisible(ctx.Width(), ctx.Height())
	imageCommand := f.loadImages(ctx, msg.page.Items)
	if f.nextCursor != nil && f.fillLoads < 3 && (!msg.more || added > 0) && f.renderedHeight(ctx.Width()) < ctx.Height() {
		return tea.Batch(imageCommand, f.load(ctx, true))
	}

	return imageCommand
}

func loadedNotice(count int, age string) string {
	if count == 1 {
		return fmt.Sprintf("Loaded 1 %s post", age)
	}

	return fmt.Sprintf("Loaded %d %s posts", count, age)
}

func (f *Feed) loadNearEnd(ctx *reactea.Ctx) tea.Cmd {
	if f.nextCursor != nil && !f.loading && f.nearEnd(ctx.Width(), ctx.Height()) {
		return f.load(ctx, true)
	}

	return nil
}

func (f *Feed) loadImages(ctx *reactea.Ctx, posts []north.Post) tea.Cmd {
	columns, rows := postcomponent.CardMediaSize(ctx.Width())

	return postcomponent.LoadImages(ctx.Context(), f.images, posts, columns, rows)
}

func (f *Feed) nearEnd(width, height int) bool {
	if len(f.posts) == 0 || width <= 0 || height <= 0 {
		return false
	}
	selected := f.posts[f.selected]
	cardHeight := lipgloss.Height(postcomponent.RenderCardWithImages(selected, width, true, f.theme, time.Now(), f.images))
	tail := cardHeight - postcomponent.CardCursorRow(selected) - 1
	tail += f.heightBetween(f.selected+1, len(f.posts)-1, width)

	return tail <= height/2
}

func (f *Feed) atTop() bool {
	return len(f.posts) == 0 || f.selected == 0 && f.top == 0
}

func (f *Feed) refreshAtTop(ctx *reactea.Ctx) tea.Cmd {
	if !f.atTop() || f.loading || !f.refreshReady {
		return nil
	}
	f.refreshReady = false

	return f.loadLatest(ctx)
}
