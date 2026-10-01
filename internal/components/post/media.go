package post

import (
	"strings"

	"github.com/Hayao0819/go-north"
)

func MediaPreviewURL(media north.Media) string {
	if media.ThumbnailURL != nil && strings.TrimSpace(*media.ThumbnailURL) != "" {
		return *media.ThumbnailURL
	}

	return media.URL
}
