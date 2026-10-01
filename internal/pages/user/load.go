package user

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/support/collection"
)

type profileLoadedMsg struct {
	target *Screen
	user   north.User
	resp   *north.Response
	err    error
}

func (m profileLoadedMsg) Response() *north.Response { return m.resp }

type postsLoadedMsg struct {
	target *Screen
	more   bool
	page   north.PostPage
	resp   *north.Response
	err    error
}

func (m postsLoadedMsg) Response() *north.Response { return m.resp }

func (d *Screen) updateResult(ctx context.Context, msg tea.Msg, width, room int) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case profileLoadedMsg:
		if msg.target != d {
			return nil, false
		}
		d.profileLoading = false
		d.profileErr = msg.err
		if msg.err == nil {
			d.user = msg.user
			d.profileLoaded = true
		}

		return d.loadImages(ctx, width, msg.user, nil), true
	case postsLoadedMsg:
		if msg.target != d {
			return nil, false
		}
		d.postsLoading = false
		d.loadingMore = false
		d.postsErr = msg.err
		if msg.err != nil {
			return nil, true
		}

		added := len(msg.page.Items)
		if msg.more {
			before := len(d.posts)
			d.posts = collection.AppendUniqueBy(d.posts, msg.page.Items, func(post north.Post) string { return post.ID })
			added = len(d.posts) - before
			d.fillLoads++
		} else {
			d.posts = append([]north.Post(nil), msg.page.Items...)
			d.selected = 0
			d.fillLoads = 0
		}
		d.nextCursor = msg.page.NextCursor
		d.clampSelection()
		imageCommand := d.loadImages(ctx, width, north.User{}, msg.page.Items)
		if d.nextCursor != nil && d.fillLoads < 3 && (!msg.more || added > 0) && len(d.content(width).lines) < room {
			return tea.Batch(imageCommand, d.loadPosts(ctx, true)), true
		}

		return imageCommand, true
	default:
		return nil, false
	}
}

func (d *Screen) loadImages(ctx context.Context, width int, user north.User, posts []north.Post) tea.Cmd {
	if d.images == nil || !d.images.Enabled() {
		return nil
	}
	commands := make([]tea.Cmd, 0, len(posts)*2+1)
	if user.AvatarURL != nil {
		commands = append(commands, d.images.Load(ctx, *user.AvatarURL, 6, 3))
	}
	contentWidth := width
	if width >= 12 {
		contentWidth -= 4
	}
	columns, rows := postcomponent.CardMediaSize(contentWidth)
	for index := range posts {
		target := posts[index].DisplayPost()
		if target == nil {
			continue
		}
		if target.Author.AvatarURL != nil {
			commands = append(commands, d.images.Load(ctx, *target.Author.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows))
		}
		for _, media := range target.Media {
			if media.Kind == north.MediaPhoto || media.Kind == north.MediaGIF {
				commands = append(commands, d.images.Load(ctx, postcomponent.MediaPreviewURL(media), columns, rows))
			}
		}
	}

	return tea.Batch(commands...)
}

func (d *Screen) loadProfile(ctx context.Context) tea.Cmd {
	handle := strings.TrimPrefix(strings.TrimSpace(d.user.Handle), "@")

	return func() tea.Msg {
		user, response, err := d.api.User(ctx, handle)

		return profileLoadedMsg{target: d, user: user, resp: response, err: err}
	}
}

func (d *Screen) loadPosts(ctx context.Context, more bool) tea.Cmd {
	if d.api == nil || d.postsLoading || more && d.nextCursor == nil {
		return nil
	}
	handle := strings.TrimPrefix(strings.TrimSpace(d.user.Handle), "@")
	cursor := ""
	if more && d.nextCursor != nil {
		cursor = *d.nextCursor
	}
	d.postsLoading = true
	d.loadingMore = more
	d.postsErr = nil

	return func() tea.Msg {
		page, response, err := d.api.UserPosts(ctx, handle, cursor)

		return postsLoadedMsg{target: d, more: more, page: page, resp: response, err: err}
	}
}

func (d *Screen) retry(ctx context.Context) tea.Cmd {
	commands := make([]tea.Cmd, 0, 2)
	if d.profileErr != nil && !d.profileLoading {
		d.profileLoading = true
		d.profileErr = nil
		commands = append(commands, d.loadProfile(ctx))
	}
	if d.postsErr != nil && !d.postsLoading {
		d.postsErr = nil
		commands = append(commands, d.loadPosts(ctx, len(d.posts) > 0))
	}

	return tea.Batch(commands...)
}

func (d *Screen) canRetry() bool {
	return d.profileErr != nil && !d.profileLoading || d.postsErr != nil && !d.postsLoading
}

func (d *Screen) loadNearEnd(ctx context.Context, width, room int) tea.Cmd {
	if d.nextCursor == nil || d.postsLoading {
		return nil
	}
	content := d.content(width)
	nearSelection := len(d.posts)-d.selected <= 4
	nearViewport := d.offset+room >= len(content.lines)-3
	if nearSelection || nearViewport {
		return d.loadPosts(ctx, true)
	}

	return nil
}
