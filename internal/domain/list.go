package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type ListItem = north.List

type ListCollection = north.ListCollection

type ListAPI interface {
	Lists(context.Context, string) (ListCollection, *north.Response, error)
	List(context.Context, string) (ListItem, *north.Response, error)
	ListTimeline(context.Context, string, string) (north.PostPage, *north.Response, error)
	FollowList(context.Context, string) (bool, *north.Response, error)
	UnfollowList(context.Context, string) (bool, *north.Response, error)
	PinList(context.Context, string) (bool, *north.Response, error)
	UnpinList(context.Context, string) (bool, *north.Response, error)
}

type ListEditorAPI interface {
	CreateList(context.Context, north.CreateListRequest) (ListItem, *north.Response, error)
	UpdateList(context.Context, string, north.UpdateListRequest) (ListItem, *north.Response, error)
	DeleteList(context.Context, string) (bool, *north.Response, error)
}

type ListMemberAPI interface {
	ListMembers(context.Context, string, string) (north.UserPage, *north.Response, error)
	ListMemberships(context.Context, string, string) ([]north.ListMembership, *north.Response, error)
	AddListMember(context.Context, string, string) (bool, *north.Response, error)
	RemoveListMember(context.Context, string, string) (bool, *north.Response, error)
}

var _ ListAPI = (*north.Client)(nil)
var _ ListEditorAPI = (*north.Client)(nil)
var _ ListMemberAPI = (*north.Client)(nil)
