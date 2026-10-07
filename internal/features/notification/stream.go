package notification

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const notificationReconnectDelay = 5 * time.Second

type streamOpenedMsg struct {
	target   *Screen
	stream   *north.NotificationStream
	response *north.Response
	err      error
}

func (m streamOpenedMsg) Response() *north.Response { return m.response }

type streamEventMsg struct {
	target *Screen
	event  north.NotificationEvent
	retry  time.Duration
	err    error
}

type streamRetryMsg struct{ target *Screen }

func (d *Screen) connectStream(ctx context.Context) tea.Cmd {
	if d.streamAPI == nil || d.streaming || d.stream != nil {
		return nil
	}
	d.streaming = true

	return func() tea.Msg {
		stream, response, err := d.streamAPI.StreamNotifications(ctx)

		return streamOpenedMsg{target: d, stream: stream, response: response, err: err}
	}
}

func (d *Screen) receiveStream() tea.Cmd {
	stream := d.stream
	if stream == nil {
		return nil
	}

	return func() tea.Msg {
		event, err := stream.Receive()
		retry := stream.RetryAfter()
		if retry <= 0 {
			retry = notificationReconnectDelay
		}

		return streamEventMsg{target: d, event: event, retry: retry, err: err}
	}
}

func (d *Screen) applyStreamOpened(ctx *reactea.Ctx, msg streamOpenedMsg) tea.Cmd {
	if msg.target != d {
		if msg.stream != nil {
			_ = msg.stream.Close()
		}

		return nil
	}
	d.streaming = false
	if msg.err != nil {
		d.notice = "Live notifications unavailable: " + ui.FriendlyError(msg.err)

		return retryStream(d, notificationReconnectDelay)
	}
	d.stream = msg.stream
	ctx.OnDestroy(func() {
		if d.stream != nil {
			_ = d.stream.Close()
		}
	})

	return d.receiveStream()
}

func (d *Screen) applyStreamEvent(ctx *reactea.Ctx, msg streamEventMsg) tea.Cmd {
	if msg.target != d {
		return nil
	}
	if msg.err != nil {
		if d.stream != nil {
			_ = d.stream.Close()
		}
		d.stream = nil
		d.streaming = false
		d.notice = "Reconnecting live notifications…"

		return retryStream(d, msg.retry)
	}
	d.notice = ""

	return tea.Batch(d.load(ctx.Context(), false), d.receiveStream())
}

func retryStream(target *Screen, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(time.Time) tea.Msg {
		return streamRetryMsg{target: target}
	})
}
