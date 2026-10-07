package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type ConversationPage struct {
	Ancestors    []north.Post
	Post         north.Post
	Replies      []north.Post
	NextCursor   *string
	ReaderRootID *string
}

type ConversationAPI interface {
	PostConversation(context.Context, string, string) (ConversationPage, *north.Response, error)
}
