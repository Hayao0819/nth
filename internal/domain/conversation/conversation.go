package conversation

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type Page struct {
	Ancestors    []north.Post
	Post         north.Post
	Replies      []north.Post
	NextCursor   *string
	ReaderRootID *string
}

type API interface {
	PostConversation(context.Context, string, string) (Page, *north.Response, error)
}
