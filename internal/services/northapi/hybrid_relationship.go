package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
)

type officialRelationshipAPI interface {
	Follow(context.Context, string) (bool, *north.Response, error)
	Unfollow(context.Context, string) (bool, *north.Response, error)
	Block(context.Context, string) (bool, *north.Response, error)
	Unblock(context.Context, string) (bool, *north.Response, error)
	Mute(context.Context, string) (bool, *north.Response, error)
	Unmute(context.Context, string) (bool, *north.Response, error)
}

type officialPostNotificationAPI interface {
	EnablePostNotifications(context.Context, string) (bool, *north.Response, error)
	DisablePostNotifications(context.Context, string) (bool, *north.Response, error)
}

func (c *Hybrid) SupportsRelationships() bool {
	_, ok := c.official.(officialRelationshipAPI)

	return ok && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsPostNotifications() bool {
	_, ok := c.official.(officialPostNotificationAPI)

	return ok && c.officialSupportsScopedAPI()
}

func (c *Hybrid) Follow(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.relationshipAPI()
	if err != nil {
		return false, nil, err
	}

	return api.Follow(ctx, handle)
}

func (c *Hybrid) Unfollow(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.relationshipAPI()
	if err != nil {
		return false, nil, err
	}

	return api.Unfollow(ctx, handle)
}

func (c *Hybrid) Block(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.relationshipAPI()
	if err != nil {
		return false, nil, err
	}

	return api.Block(ctx, handle)
}

func (c *Hybrid) Unblock(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.relationshipAPI()
	if err != nil {
		return false, nil, err
	}

	return api.Unblock(ctx, handle)
}

func (c *Hybrid) Mute(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.relationshipAPI()
	if err != nil {
		return false, nil, err
	}

	return api.Mute(ctx, handle)
}

func (c *Hybrid) Unmute(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.relationshipAPI()
	if err != nil {
		return false, nil, err
	}

	return api.Unmute(ctx, handle)
}

func (c *Hybrid) EnablePostNotifications(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.postNotificationAPI()
	if err != nil {
		return false, nil, err
	}

	return api.EnablePostNotifications(ctx, handle)
}

func (c *Hybrid) DisablePostNotifications(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, err := c.postNotificationAPI()
	if err != nil {
		return false, nil, err
	}

	return api.DisablePostNotifications(ctx, handle)
}

func (c *Hybrid) relationshipAPI() (officialRelationshipAPI, error) {
	api, ok := c.official.(officialRelationshipAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return nil, errors.New("north public API authentication is not configured")
	}

	return api, nil
}

func (c *Hybrid) postNotificationAPI() (officialPostNotificationAPI, error) {
	api, ok := c.official.(officialPostNotificationAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return nil, errors.New("post notifications are not available with the configured credentials")
	}

	return api, nil
}
