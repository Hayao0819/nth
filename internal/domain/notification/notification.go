package notification

import (
	"context"
	"time"

	"github.com/Hayao0819/go-north"
)

type Kind string

const (
	Follow  Kind = "FOLLOW"
	Like    Kind = "LIKE"
	Repost  Kind = "RETWEET"
	Post    Kind = "POST"
	Reply   Kind = "REPLY"
	Quote   Kind = "QUOTE"
	Mention Kind = "MENTION"
)

type Item struct {
	ID          string
	Kind        Kind
	Read        bool
	Actors      []north.User
	ActorCount  int
	TargetCount int
	CreatedAt   time.Time
	Post        *north.Post
}

type Page struct {
	Items      []Item
	NextCursor *string
}

type API interface {
	Notifications(context.Context, string) (Page, *north.Response, error)
	NotificationUnreadCount(context.Context) (int, *north.Response, error)
	MarkNotificationsRead(context.Context) (*north.Response, error)
}
