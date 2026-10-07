package post

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
)

func MediaPreviewURL(media north.Media) string {
	if media.ThumbnailURL != nil && strings.TrimSpace(*media.ThumbnailURL) != "" {
		return *media.ThumbnailURL
	}

	return media.URL
}

// LoadImages queues the avatar and preview images rendered by post cards.
func LoadImages(ctx context.Context, images *termimage.Renderer, posts []north.Post, columns, rows int) tea.Cmd {
	if images == nil || !images.Enabled() {
		return nil
	}
	commands := make([]tea.Cmd, 0, len(posts)*2)
	for index := range posts {
		target := posts[index].DisplayPost()
		if target == nil {
			continue
		}
		if target.Author.AvatarURL != nil {
			commands = append(commands, images.Load(ctx, *target.Author.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows))
		}
		for _, media := range target.Media {
			if media.Kind == north.MediaPhoto || media.Kind == north.MediaGIF {
				commands = append(commands, images.Load(ctx, MediaPreviewURL(media), columns, rows))
			}
		}
	}

	return tea.Batch(commands...)
}

func CardMediaSize(width int) (int, int) {
	return max(1, min(48, width-8)), 6
}

func cardMediaBody(media north.Media, width int, images *termimage.Renderer) []string {
	var body []string
	if media.Kind == north.MediaPhoto || media.Kind == north.MediaGIF {
		columns, rows := CardMediaSize(width + 2)
		if rendered := images.View(MediaPreviewURL(media), columns, rows); rendered != "" {
			body = append(body, strings.Split(rendered, "\n")...)
		}
	}
	if media.AltText != nil && *media.AltText != "" {
		body = append(body, ui.WrappedLines(*media.AltText, max(1, width-6))...)
	}

	return body
}

func MediaLabel(media north.Media) string {
	kind := "Media"
	switch media.Kind {
	case north.MediaPhoto:
		kind = "Photo"
	case north.MediaGIF:
		kind = "GIF"
	case north.MediaVideo:
		kind = "Video"
	}
	detail := "▣ " + kind
	if media.Width > 0 && media.Height > 0 {
		detail += fmt.Sprintf(" · %d×%d", media.Width, media.Height)
	}
	if media.Sensitive {
		detail += " · sensitive"
	}
	if media.Warning != nil {
		detail += " · " + strings.ToLower(ui.SafeInline(string(*media.Warning)))
	}

	return detail
}
