package app

import (
	"context"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
)

// API is the north functionality used by the terminal interface.
type API interface {
	feed.API
	Me(context.Context) (north.User, *north.Response, error)
	User(context.Context, string) (north.User, *north.Response, error)
	UserPosts(context.Context, string, string) (north.PostPage, *north.Response, error)
	Post(context.Context, string) (north.Post, *north.Response, error)
	CreatePost(context.Context, north.CreatePostRequest) (north.CreatedPost, *north.Response, error)
	DeletePost(context.Context, string) (bool, *north.Response, error)
}
