package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialMediaAPI interface {
	domain.MediaAPI
}

var errMediaUnavailable = errors.New("media uploads are not available with the configured credentials")

func (c *Hybrid) UploadMediaFile(ctx context.Context, path string, options ...north.MediaUploadOption) (north.Media, *north.Response, error) {
	api, ok := c.mediaAPI()
	if !ok {
		return north.Media{}, nil, errMediaUnavailable
	}

	return api.UploadMediaFile(ctx, path, options...)
}

func (c *Hybrid) DeleteMedia(ctx context.Context, id string) (bool, *north.Response, error) {
	api, ok := c.mediaAPI()
	if !ok {
		return false, nil, errMediaUnavailable
	}

	return api.DeleteMedia(ctx, id)
}

func (c *Hybrid) SetMediaAltText(ctx context.Context, id, text string) (bool, *north.Response, error) {
	api, ok := c.mediaAPI()
	if !ok {
		return false, nil, errMediaUnavailable
	}

	return api.SetMediaAltText(ctx, id, text)
}

func (c *Hybrid) SetMediaWarning(ctx context.Context, id string, warning *north.MediaWarning) (*north.MediaWarning, *north.Response, error) {
	api, ok := c.mediaAPI()
	if !ok {
		return nil, nil, errMediaUnavailable
	}

	return api.SetMediaWarning(ctx, id, warning)
}

func (c *Hybrid) mediaAPI() (officialMediaAPI, bool) {
	api, ok := c.official.(officialMediaAPI)

	return api, ok && c.officialSupportsScopedAPI()
}
