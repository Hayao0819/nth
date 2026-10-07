package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type UserActivityAPI interface {
	LikedPosts(context.Context, string, string) (north.PostPage, *north.Response, error)
}

type UserConnectionsAPI interface {
	Followers(context.Context, string, string) (north.UserPage, *north.Response, error)
	Following(context.Context, string, string) (north.UserPage, *north.Response, error)
}

type ProfileUpdate struct {
	Name       string
	Bio        string
	Location   string
	Website    string
	AvatarPath string
	HeaderPath string
}

type ProfileAPI interface {
	UpdateProfile(context.Context, ProfileUpdate) (north.User, *north.Response, error)
}
