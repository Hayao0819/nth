package notification

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	notificationdomain "github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/support/collection"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type PageLoadedMsg struct {
	target   *Screen
	page     notificationdomain.Page
	response *north.Response
	err      error
	more     bool
}

func (m PageLoadedMsg) Response() *north.Response { return m.response }

type ReadMsg struct {
	target   *Screen
	response *north.Response
	err      error
}

func (m ReadMsg) Response() *north.Response { return m.response }

func (m ReadMsg) Error() error { return m.err }

func (d *Screen) loadImages(ctx context.Context, items []notificationdomain.Item) tea.Cmd {
	if d.images == nil || !d.images.Enabled() {
		return nil
	}
	commands := make([]tea.Cmd, 0, len(items))
	for _, item := range items {
		for _, actor := range item.Actors {
			if actor.AvatarURL != nil {
				commands = append(commands, d.images.Load(ctx, *actor.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows))
			}
		}
	}

	return tea.Batch(commands...)
}

func (d *Screen) load(ctx context.Context, more bool) tea.Cmd {
	if d.api == nil || d.loading || more && d.nextCursor == nil {
		return nil
	}
	d.loading = true
	d.loadingMore = more
	d.err = nil
	cursor := ""
	if more {
		cursor = *d.nextCursor
	}

	return func() tea.Msg {
		page, response, err := d.api.Notifications(ctx, cursor)
		return PageLoadedMsg{target: d, page: page, response: response, err: err, more: more}
	}
}

func (d *Screen) markRead(ctx context.Context) tea.Cmd {
	if d.api == nil || d.markedRead || d.markingRead {
		return nil
	}
	d.markingRead = true

	return func() tea.Msg {
		response, err := d.api.MarkNotificationsRead(ctx)
		return ReadMsg{target: d, response: response, err: err}
	}
}

func (d *Screen) loadNearEnd(ctx context.Context) tea.Cmd {
	if len(d.items)-d.selected <= 3 {
		return d.load(ctx, true)
	}

	return nil
}

func (d *Screen) applyPageLoaded(ctx *reactea.Ctx, msg PageLoadedMsg) tea.Cmd {
	if msg.target != d {
		return nil
	}
	d.loading = false
	d.loadingMore = false
	d.err = msg.err
	if msg.err != nil {
		return nil
	}
	if msg.more {
		d.items = collection.AppendUniqueBy(d.items, msg.page.Items, func(item notificationdomain.Item) string {
			return item.ID
		})
	} else {
		d.items = append([]notificationdomain.Item(nil), msg.page.Items...)
		d.selected, d.top = 0, 0
	}
	if d.markedRead {
		for index := range d.items {
			d.items[index].Read = true
		}
	}
	d.nextCursor = msg.page.NextCursor
	d.ensureVisible(d.innerWidth(ctx.Width()), d.room(ctx.Height()))
	imageCommand := d.loadImages(ctx.Context(), msg.page.Items)
	if !msg.more && !d.markedRead {
		return tea.Batch(imageCommand, d.markRead(ctx.Context()))
	}

	return imageCommand
}

func (d *Screen) applyRead(msg ReadMsg) {
	if msg.target != d {
		return
	}
	d.markingRead = false
	if msg.err != nil {
		d.notice = "Could not mark notifications as read: " + ui.FriendlyError(msg.err)

		return
	}
	d.markedRead = true
	for index := range d.items {
		d.items[index].Read = true
	}
}
