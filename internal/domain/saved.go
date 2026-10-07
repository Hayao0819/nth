package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type SavedPostAPI interface {
	Drafts(context.Context, string) ([]north.Draft, *north.Response, error)
	CreateDraft(context.Context, north.DraftRequest) (north.Draft, *north.Response, error)
	UpdateDraft(context.Context, string, north.DraftRequest) (north.Draft, *north.Response, error)
	DeleteDrafts(context.Context, ...string) (int, *north.Response, error)
	ScheduledPosts(context.Context, string) ([]north.ScheduledPost, *north.Response, error)
	SchedulePost(context.Context, north.SchedulePostRequest) (north.ScheduledPost, *north.Response, error)
	UpdateScheduledPost(context.Context, string, north.SchedulePostRequest) (north.ScheduledPost, *north.Response, error)
	DeleteScheduledPosts(context.Context, ...string) (int, *north.Response, error)
	PendingPosts(context.Context, string) ([]north.PendingPost, *north.Response, error)
	CreatePendingPost(context.Context, north.CreatePendingPostRequest) (north.PendingPost, *north.Response, error)
	SendPendingPost(context.Context, string) (north.PendingPostState, *north.Response, error)
	UndoPendingPost(context.Context, string) (north.Draft, *north.Response, error)
}

var _ SavedPostAPI = (*north.Client)(nil)
