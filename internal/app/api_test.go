package app

import (
	"context"
	"sync"
	"time"

	"github.com/Hayao0819/go-north"
)

type fakeAPI struct {
	mu sync.Mutex

	me        north.User
	meErr     error
	createErr error
	deleteErr error
	meCalls   int
	userCalls []string
	posts     map[string]north.Post

	homeCalls     []north.TimelineOptions
	mentionCalls  []string
	userPostCalls []string
	searchCalls   []string
	created       []north.CreatePostRequest
	deleted       []string
	liked         []string
	unliked       []string
	reposted      []string
	unreposted    []string
	listCalls     int
}

func (f *fakeAPI) response() *north.Response {
	return &north.Response{RateLimit: north.RateLimit{Present: true, Limit: 180, Remaining: 179}}
}

func (f *fakeAPI) Me(context.Context) (north.User, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.meCalls++
	if f.me.ID == "" {
		f.me = north.User{ID: "me", Handle: "alice", Name: "Alice"}
	}
	if f.meErr != nil {
		return north.User{}, f.response(), f.meErr
	}

	return f.me, f.response(), nil
}

func (f *fakeAPI) User(_ context.Context, handle string) (north.User, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.userCalls = append(f.userCalls, handle)
	bio := "Profile for @" + handle
	location := "North"
	website := "https://north.rip/" + handle
	created := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)

	return north.User{
		ID:             "user-" + handle,
		Handle:         handle,
		Name:           "User " + handle,
		Bio:            &bio,
		Location:       &location,
		Website:        &website,
		FollowerCount:  12,
		FollowingCount: 34,
		PostCount:      56,
		CreatedAt:      &created,
	}, f.response(), nil
}

func (f *fakeAPI) HomeTimeline(_ context.Context, options north.TimelineOptions) (north.PostPage, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.homeCalls = append(f.homeCalls, options)

	if options.Cursor == "next" {
		return north.PostPage{Items: []north.Post{testPost("2", "second"), testPost("3", "third")}}, f.response(), nil
	}
	next := "next"

	return north.PostPage{Items: []north.Post{testPost("1", "first"), testPost("2", "second")}, NextCursor: &next}, f.response(), nil
}

func (f *fakeAPI) SearchPosts(_ context.Context, query string, _ north.SearchOptions) (north.PostPage, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searchCalls = append(f.searchCalls, query)

	return north.PostPage{Items: []north.Post{testPost("search", query)}}, f.response(), nil
}

func (f *fakeAPI) Mentions(_ context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mentionCalls = append(f.mentionCalls, handle+":"+cursor)

	return north.PostPage{Items: []north.Post{testPost("mention", "@"+handle+" hello")}}, f.response(), nil
}

func (f *fakeAPI) UserPosts(_ context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.userPostCalls = append(f.userPostCalls, handle+":"+cursor)

	return north.PostPage{Items: []north.Post{testPost("mine", "my post")}}, f.response(), nil
}

func (f *fakeAPI) CreatePost(_ context.Context, request north.CreatePostRequest) (north.CreatedPost, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, request)
	if f.createErr != nil {
		return north.CreatedPost{}, f.response(), f.createErr
	}

	return north.CreatedPost{ID: "created", Text: request.Text}, f.response(), nil
}

func (f *fakeAPI) DeletePost(_ context.Context, id string) (bool, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, id)
	if f.deleteErr != nil {
		return false, f.response(), f.deleteErr
	}

	return true, f.response(), nil
}

func (f *fakeAPI) Post(_ context.Context, id string) (north.Post, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if post, ok := f.posts[id]; ok {
		return post, f.response(), nil
	}

	return testPost(id, "parent post"), f.response(), nil
}

func (f *fakeAPI) Like(_ context.Context, id string) (north.LikeState, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.liked = append(f.liked, id)

	return north.LikeState{Liked: true, LikeCount: 9}, f.response(), nil
}

func (f *fakeAPI) Unlike(_ context.Context, id string) (north.LikeState, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unliked = append(f.unliked, id)

	return north.LikeState{Liked: false, LikeCount: 8}, f.response(), nil
}

func (f *fakeAPI) Repost(_ context.Context, id string) (north.RepostState, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reposted = append(f.reposted, id)

	return north.RepostState{Reposted: true, RepostCount: 4}, f.response(), nil
}

func (f *fakeAPI) UndoRepost(_ context.Context, id string) (north.RepostState, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unreposted = append(f.unreposted, id)

	return north.RepostState{Reposted: false, RepostCount: 3}, f.response(), nil
}

func (f *fakeAPI) Lists(context.Context, string) (north.ListCollection, *north.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listCalls++

	return north.ListCollection{Items: []north.List{{
		ID:            "friends",
		Name:          "Friends",
		Owner:         north.User{ID: "me", Handle: "alice", Name: "Alice"},
		OwnedByViewer: true,
		MemberCount:   2,
	}}}, f.response(), nil
}

func (f *fakeAPI) List(_ context.Context, id string) (north.List, *north.Response, error) {
	return north.List{ID: id, Name: "Friends", Owner: north.User{Handle: "alice"}}, f.response(), nil
}

func (f *fakeAPI) ListTimeline(context.Context, string, string) (north.PostPage, *north.Response, error) {
	return north.PostPage{Items: []north.Post{testPost("list-post", "from a list")}}, f.response(), nil
}

func (f *fakeAPI) FollowList(context.Context, string) (bool, *north.Response, error) {
	return true, f.response(), nil
}

func (f *fakeAPI) UnfollowList(context.Context, string) (bool, *north.Response, error) {
	return true, f.response(), nil
}

func (f *fakeAPI) PinList(context.Context, string) (bool, *north.Response, error) {
	return true, f.response(), nil
}

func (f *fakeAPI) UnpinList(context.Context, string) (bool, *north.Response, error) {
	return true, f.response(), nil
}

func testPost(id, text string) north.Post {
	return north.Post{
		ID:        id,
		Text:      text,
		CreatedAt: time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC),
		Author:    north.User{ID: "user-" + id, Handle: "user" + id, Name: "User " + id},
		Media:     []north.Media{},
	}
}

var _ API = (*fakeAPI)(nil)
