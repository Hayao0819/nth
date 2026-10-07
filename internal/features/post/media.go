package post

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
)

func DetailMediaSize(width int) (int, int) {
	return max(1, min(64, width-4)), 10
}

func (d *Screen) loadImages(ctx context.Context, width int, posts []north.Post) tea.Cmd {
	columns, rows := DetailMediaSize(width)

	return postcomponent.LoadImages(ctx, d.images, posts, columns, rows)
}
