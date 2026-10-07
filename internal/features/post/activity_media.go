package post

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
)

func (s *Activity) loadImages(ctx context.Context, width int, posts []north.Post) tea.Cmd {
	if s.images == nil || !s.images.Enabled() {
		return nil
	}
	commands := make([]tea.Cmd, 0, len(posts)*2+len(s.users))
	for _, user := range s.users {
		if user.AvatarURL != nil {
			commands = append(commands, s.images.Load(ctx, *user.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows))
		}
	}
	columns, rows := postcomponent.CardMediaSize(width)
	if command := postcomponent.LoadImages(ctx, s.images, posts, columns, rows); command != nil {
		commands = append(commands, command)
	}

	return tea.Batch(commands...)
}
