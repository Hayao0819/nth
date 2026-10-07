package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type PostActivityAPI interface {
	Quotes(context.Context, string, string) (north.PostPage, *north.Response, error)
	PostLikes(context.Context, string, string) (north.UserPage, *north.Response, error)
	PostReposts(context.Context, string, string) (north.UserPage, *north.Response, error)
	PostEditHistory(context.Context, string) (north.EditHistory, *north.Response, error)
}

type PollAPI interface {
	VotePoll(context.Context, string, string) (north.Post, *north.Response, error)
}

type ThreadAPI interface {
	CreateThread(context.Context, []north.ThreadItem) ([]north.Post, *north.Response, error)
}
