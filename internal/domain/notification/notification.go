package notification

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type Kind = north.NotificationKind

const (
	Follow  = north.NotificationFollow
	Like    = north.NotificationLike
	Repost  = north.NotificationRepost
	Post    = north.NotificationPost
	Reply   = north.NotificationReply
	Quote   = north.NotificationQuote
	Mention = north.NotificationMention
)

type Item = north.Notification

type Page = north.NotificationPage

type API interface {
	Notifications(context.Context, north.NotificationTab, string) (north.NotificationPage, *north.Response, error)
	NotificationUnreadCount(context.Context) (int, *north.Response, error)
	MarkNotificationsRead(context.Context) (int, *north.Response, error)
}

var _ API = (*north.Client)(nil)
