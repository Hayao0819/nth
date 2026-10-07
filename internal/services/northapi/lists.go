package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialListAPI interface {
	domain.ListAPI
}

type officialListEditorAPI interface {
	domain.ListEditorAPI
}

type officialListMemberAPI interface {
	domain.ListMemberAPI
}

func (c *Hybrid) Lists(ctx context.Context, cursor string) (north.ListCollection, *north.Response, error) {
	official, ok := c.official.(officialListAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.ListCollection{}, nil, errors.New("lists are not available with the configured credentials")
	}

	return official.Lists(ctx, cursor)
}

func (c *Hybrid) List(ctx context.Context, id string) (north.List, *north.Response, error) {
	official, ok := c.official.(officialListAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.List{}, nil, errors.New("lists are not available with the configured credentials")
	}

	return official.List(ctx, id)
}

func (c *Hybrid) ListTimeline(ctx context.Context, id, cursor string) (north.PostPage, *north.Response, error) {
	official, ok := c.official.(officialListAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.PostPage{}, nil, errors.New("lists are not available with the configured credentials")
	}

	return official.ListTimeline(ctx, id, cursor)
}

func (c *Hybrid) FollowList(ctx context.Context, id string) (bool, *north.Response, error) {
	return c.setListState(func(api officialListAPI) (bool, *north.Response, error) {
		return api.FollowList(ctx, id)
	})
}

func (c *Hybrid) UnfollowList(ctx context.Context, id string) (bool, *north.Response, error) {
	return c.setListState(func(api officialListAPI) (bool, *north.Response, error) {
		return api.UnfollowList(ctx, id)
	})
}

func (c *Hybrid) PinList(ctx context.Context, id string) (bool, *north.Response, error) {
	return c.setListState(func(api officialListAPI) (bool, *north.Response, error) {
		return api.PinList(ctx, id)
	})
}

func (c *Hybrid) UnpinList(ctx context.Context, id string) (bool, *north.Response, error) {
	return c.setListState(func(api officialListAPI) (bool, *north.Response, error) {
		return api.UnpinList(ctx, id)
	})
}

func (c *Hybrid) setListState(set func(officialListAPI) (bool, *north.Response, error)) (bool, *north.Response, error) {
	official, ok := c.official.(officialListAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errors.New("lists are not available with the configured credentials")
	}

	return set(official)
}

func (c *Hybrid) CreateList(ctx context.Context, request north.CreateListRequest) (north.List, *north.Response, error) {
	api, ok := c.official.(officialListEditorAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.List{}, nil, errors.New("list editing is not available with the configured credentials")
	}

	return api.CreateList(ctx, request)
}

func (c *Hybrid) UpdateList(ctx context.Context, id string, request north.UpdateListRequest) (north.List, *north.Response, error) {
	api, ok := c.official.(officialListEditorAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.List{}, nil, errors.New("list editing is not available with the configured credentials")
	}

	return api.UpdateList(ctx, id, request)
}

func (c *Hybrid) DeleteList(ctx context.Context, id string) (bool, *north.Response, error) {
	api, ok := c.official.(officialListEditorAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errors.New("list editing is not available with the configured credentials")
	}

	return api.DeleteList(ctx, id)
}

func (c *Hybrid) ListMembers(ctx context.Context, id, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.official.(officialListMemberAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.UserPage{}, nil, errors.New("list members are not available with the configured credentials")
	}

	return api.ListMembers(ctx, id, cursor)
}

func (c *Hybrid) ListMemberships(ctx context.Context, handle, cursor string) ([]north.ListMembership, *north.Response, error) {
	api, ok := c.official.(officialListMemberAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return nil, nil, errors.New("list members are not available with the configured credentials")
	}

	return api.ListMemberships(ctx, handle, cursor)
}

func (c *Hybrid) AddListMember(ctx context.Context, id, handle string) (bool, *north.Response, error) {
	api, ok := c.official.(officialListMemberAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errors.New("list members are not available with the configured credentials")
	}

	return api.AddListMember(ctx, id, handle)
}

func (c *Hybrid) RemoveListMember(ctx context.Context, id, handle string) (bool, *north.Response, error) {
	api, ok := c.official.(officialListMemberAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errors.New("list members are not available with the configured credentials")
	}

	return api.RemoveListMember(ctx, id, handle)
}
