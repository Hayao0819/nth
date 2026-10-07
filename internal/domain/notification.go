package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type NotificationKind = north.NotificationKind

const (
	NotificationFollow  = north.NotificationFollow
	NotificationLike    = north.NotificationLike
	NotificationRepost  = north.NotificationRepost
	NotificationPost    = north.NotificationPost
	NotificationReply   = north.NotificationReply
	NotificationQuote   = north.NotificationQuote
	NotificationMention = north.NotificationMention
)

type NotificationItem = north.Notification

type NotificationPage = north.NotificationPage

type NotificationAPI interface {
	Notifications(context.Context, north.NotificationTab, string) (north.NotificationPage, *north.Response, error)
	NotificationUnreadCount(context.Context) (int, *north.Response, error)
	MarkNotificationsRead(context.Context) (int, *north.Response, error)
}

type NotificationStreamAPI interface {
	StreamNotifications(context.Context) (*north.NotificationStream, *north.Response, error)
}

var _ NotificationAPI = (*north.Client)(nil)
