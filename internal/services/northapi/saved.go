package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialSavedPostAPI interface {
	domain.SavedPostAPI
}

var errSavedPostsUnavailable = errors.New("saved posts are not available with the configured credentials")

func (c *Hybrid) Drafts(ctx context.Context, cursor string) ([]north.Draft, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return nil, nil, errSavedPostsUnavailable
	}

	return api.Drafts(ctx, cursor)
}

func (c *Hybrid) CreateDraft(ctx context.Context, request north.DraftRequest) (north.Draft, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return north.Draft{}, nil, errSavedPostsUnavailable
	}

	return api.CreateDraft(ctx, request)
}

func (c *Hybrid) UpdateDraft(ctx context.Context, id string, request north.DraftRequest) (north.Draft, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return north.Draft{}, nil, errSavedPostsUnavailable
	}

	return api.UpdateDraft(ctx, id, request)
}

func (c *Hybrid) DeleteDrafts(ctx context.Context, ids ...string) (int, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return 0, nil, errSavedPostsUnavailable
	}

	return api.DeleteDrafts(ctx, ids...)
}

func (c *Hybrid) ScheduledPosts(ctx context.Context, cursor string) ([]north.ScheduledPost, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return nil, nil, errSavedPostsUnavailable
	}

	return api.ScheduledPosts(ctx, cursor)
}

func (c *Hybrid) SchedulePost(ctx context.Context, request north.SchedulePostRequest) (north.ScheduledPost, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return north.ScheduledPost{}, nil, errSavedPostsUnavailable
	}

	return api.SchedulePost(ctx, request)
}

func (c *Hybrid) UpdateScheduledPost(ctx context.Context, id string, request north.SchedulePostRequest) (north.ScheduledPost, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return north.ScheduledPost{}, nil, errSavedPostsUnavailable
	}

	return api.UpdateScheduledPost(ctx, id, request)
}

func (c *Hybrid) DeleteScheduledPosts(ctx context.Context, ids ...string) (int, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return 0, nil, errSavedPostsUnavailable
	}

	return api.DeleteScheduledPosts(ctx, ids...)
}

func (c *Hybrid) PendingPosts(ctx context.Context, cursor string) ([]north.PendingPost, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return nil, nil, errSavedPostsUnavailable
	}

	return api.PendingPosts(ctx, cursor)
}

func (c *Hybrid) CreatePendingPost(ctx context.Context, request north.CreatePendingPostRequest) (north.PendingPost, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return north.PendingPost{}, nil, errSavedPostsUnavailable
	}

	return api.CreatePendingPost(ctx, request)
}

func (c *Hybrid) SendPendingPost(ctx context.Context, id string) (north.PendingPostState, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return north.PendingPostState{}, nil, errSavedPostsUnavailable
	}

	return api.SendPendingPost(ctx, id)
}

func (c *Hybrid) UndoPendingPost(ctx context.Context, id string) (north.Draft, *north.Response, error) {
	api, ok := c.savedPostAPI()
	if !ok {
		return north.Draft{}, nil, errSavedPostsUnavailable
	}

	return api.UndoPendingPost(ctx, id)
}

func (c *Hybrid) savedPostAPI() (officialSavedPostAPI, bool) {
	api, ok := c.official.(officialSavedPostAPI)

	return api, ok && c.officialSupportsScopedAPI()
}
