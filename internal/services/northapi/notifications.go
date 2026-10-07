package northapi

import (
	"context"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialNotificationStreamAPI interface {
	StreamNotifications(context.Context) (*north.NotificationStream, *north.Response, error)
}

func (c *Hybrid) Notifications(ctx context.Context, tab north.NotificationTab, cursor string) (north.NotificationPage, *north.Response, error) {
	if official, ok := c.official.(domain.NotificationAPI); ok {
		return official.Notifications(ctx, tab, cursor)
	}

	return withWeb(ctx, c, func(client *Client) (north.NotificationPage, *north.Response, error) {
		return client.Notifications(ctx, tab, cursor)
	})
}

func (c *Hybrid) NotificationUnreadCount(ctx context.Context) (int, *north.Response, error) {
	if official, ok := c.official.(domain.NotificationAPI); ok {
		return official.NotificationUnreadCount(ctx)
	}

	return withWeb(ctx, c, func(client *Client) (int, *north.Response, error) {
		return client.NotificationUnreadCount(ctx)
	})
}

func (c *Hybrid) MarkNotificationsRead(ctx context.Context) (int, *north.Response, error) {
	if official, ok := c.official.(domain.NotificationAPI); ok {
		return official.MarkNotificationsRead(ctx)
	}

	return withWeb(ctx, c, func(client *Client) (int, *north.Response, error) {
		return client.MarkNotificationsRead(ctx)
	})
}
