package feed

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type API interface {
	HomeTimeline(context.Context, north.TimelineOptions) (north.PostPage, *north.Response, error)
	SearchPosts(context.Context, string, north.SearchOptions) (north.PostPage, *north.Response, error)
	Like(context.Context, string) (north.LikeState, *north.Response, error)
	Unlike(context.Context, string) (north.LikeState, *north.Response, error)
	Repost(context.Context, string) (north.RepostState, *north.Response, error)
	UndoRepost(context.Context, string) (north.RepostState, *north.Response, error)
}

type Loader func(context.Context, string) (north.PostPage, *north.Response, error)
