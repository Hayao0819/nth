package postpage

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
)

func DetailMediaSize(width int) (int, int) {
	return max(1, min(64, width-4)), 10
}

func (d *Screen) loadImages(ctx context.Context, width int, posts []north.Post) tea.Cmd {
	if d.images == nil || !d.images.Enabled() {
		return nil
	}
	commands := make([]tea.Cmd, 0, len(posts)*2)
	for index := range posts {
		target := posts[index].DisplayPost()
		if target == nil {
			continue
		}
		if target.Author.AvatarURL != nil {
			commands = append(commands, d.images.Load(ctx, *target.Author.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows))
		}
		columns, rows := DetailMediaSize(width)
		for _, media := range target.Media {
			if media.Kind == north.MediaPhoto || media.Kind == north.MediaGIF {
				commands = append(commands, d.images.Load(ctx, postcomponent.MediaPreviewURL(media), columns, rows))
			}
		}
	}

	return tea.Batch(commands...)
}
