package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type AccountSafetyAPI interface {
	FollowRequests(context.Context, string) (north.UserPage, *north.Response, error)
	AcceptFollowRequest(context.Context, string) (bool, *north.Response, error)
	RejectFollowRequest(context.Context, string) (bool, *north.Response, error)
	BlockedUsers(context.Context, string) (north.UserPage, *north.Response, error)
	MutedUsers(context.Context, string) (north.UserPage, *north.Response, error)
	Unblock(context.Context, string) (bool, *north.Response, error)
	Unmute(context.Context, string) (bool, *north.Response, error)
}

type MutedKeywordAPI interface {
	MutedKeywords(context.Context, string) ([]north.MutedKeyword, *north.Response, error)
	CreateMutedKeyword(context.Context, north.CreateMutedKeywordRequest) (north.MutedKeyword, *north.Response, error)
	DeleteMutedKeyword(context.Context, string) (bool, *north.Response, error)
}

var _ AccountSafetyAPI = (*north.Client)(nil)
var _ MutedKeywordAPI = (*north.Client)(nil)
