package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type MediaAPI interface {
	UploadMediaFile(context.Context, string, ...north.MediaUploadOption) (north.Media, *north.Response, error)
	DeleteMedia(context.Context, string) (bool, *north.Response, error)
	SetMediaAltText(context.Context, string, string) (bool, *north.Response, error)
	SetMediaWarning(context.Context, string, *north.MediaWarning) (*north.MediaWarning, *north.Response, error)
}

var _ MediaAPI = (*north.Client)(nil)
