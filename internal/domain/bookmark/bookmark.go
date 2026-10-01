package bookmark

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type API interface {
	Bookmarks(context.Context, string) (north.PostPage, *north.Response, error)
}
