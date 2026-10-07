package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialAccountSafetyAPI interface {
	domain.AccountSafetyAPI
}

type officialMutedKeywordAPI interface {
	domain.MutedKeywordAPI
}

var errAccountSafetyUnavailable = errors.New("account safety controls are not available with the configured credentials")

func (c *Hybrid) FollowRequests(ctx context.Context, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.accountSafetyAPI()
	if !ok {
		return north.UserPage{}, nil, errAccountSafetyUnavailable
	}

	return api.FollowRequests(ctx, cursor)
}

func (c *Hybrid) AcceptFollowRequest(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, ok := c.accountSafetyAPI()
	if !ok {
		return false, nil, errAccountSafetyUnavailable
	}

	return api.AcceptFollowRequest(ctx, handle)
}

func (c *Hybrid) RejectFollowRequest(ctx context.Context, handle string) (bool, *north.Response, error) {
	api, ok := c.accountSafetyAPI()
	if !ok {
		return false, nil, errAccountSafetyUnavailable
	}

	return api.RejectFollowRequest(ctx, handle)
}

func (c *Hybrid) BlockedUsers(ctx context.Context, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.accountSafetyAPI()
	if !ok {
		return north.UserPage{}, nil, errAccountSafetyUnavailable
	}

	return api.BlockedUsers(ctx, cursor)
}

func (c *Hybrid) MutedUsers(ctx context.Context, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.accountSafetyAPI()
	if !ok {
		return north.UserPage{}, nil, errAccountSafetyUnavailable
	}

	return api.MutedUsers(ctx, cursor)
}

func (c *Hybrid) MutedKeywords(ctx context.Context, cursor string) ([]north.MutedKeyword, *north.Response, error) {
	api, ok := c.mutedKeywordAPI()
	if !ok {
		return nil, nil, errAccountSafetyUnavailable
	}

	return api.MutedKeywords(ctx, cursor)
}

func (c *Hybrid) CreateMutedKeyword(ctx context.Context, request north.CreateMutedKeywordRequest) (north.MutedKeyword, *north.Response, error) {
	api, ok := c.mutedKeywordAPI()
	if !ok {
		return north.MutedKeyword{}, nil, errAccountSafetyUnavailable
	}

	return api.CreateMutedKeyword(ctx, request)
}

func (c *Hybrid) DeleteMutedKeyword(ctx context.Context, id string) (bool, *north.Response, error) {
	api, ok := c.mutedKeywordAPI()
	if !ok {
		return false, nil, errAccountSafetyUnavailable
	}

	return api.DeleteMutedKeyword(ctx, id)
}

func (c *Hybrid) accountSafetyAPI() (officialAccountSafetyAPI, bool) {
	api, ok := c.official.(officialAccountSafetyAPI)

	return api, ok && c.officialSupportsScopedAPI()
}

func (c *Hybrid) mutedKeywordAPI() (officialMutedKeywordAPI, bool) {
	api, ok := c.official.(officialMutedKeywordAPI)

	return api, ok && c.officialSupportsScopedAPI()
}
