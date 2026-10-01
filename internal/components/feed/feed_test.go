package feed

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type fakeAPI struct {
	homeCalls  []north.TimelineOptions
	home       []north.Post
	older      []north.Post
	latest     []north.Post
	liked      []string
	unliked    []string
	reposted   []string
	unreposted []string
}

func (f *fakeAPI) HomeTimeline(_ context.Context, options north.TimelineOptions) (north.PostPage, *north.Response, error) {
	f.homeCalls = append(f.homeCalls, options)
	if options.Cursor == "next" {
		if f.older != nil {
			return north.PostPage{Items: f.older}, nil, nil
		}
		return north.PostPage{Items: []north.Post{testPost("2", "second"), testPost("3", "third")}}, nil, nil
	}
	next := "next"
	if f.latest != nil {
		return north.PostPage{Items: f.latest, NextCursor: &next}, nil, nil
	}
	if f.home != nil {
		return north.PostPage{Items: f.home, NextCursor: &next}, nil, nil
	}

	return north.PostPage{Items: []north.Post{testPost("1", "first"), testPost("2", "second")}, NextCursor: &next}, nil, nil
}

func (f *fakeAPI) SearchPosts(_ context.Context, query string, _ north.SearchOptions) (north.PostPage, *north.Response, error) {
	return north.PostPage{Items: []north.Post{testPost("search", query)}}, nil, nil
}

func (f *fakeAPI) Mentions(_ context.Context, handle, _ string) (north.PostPage, *north.Response, error) {
	return north.PostPage{Items: []north.Post{testPost("mention", "@"+handle)}}, nil, nil
}

func (f *fakeAPI) UserPosts(_ context.Context, handle, _ string) (north.PostPage, *north.Response, error) {
	return north.PostPage{Items: []north.Post{testPost("mine", handle)}}, nil, nil
}

func (f *fakeAPI) Like(_ context.Context, id string) (north.LikeState, *north.Response, error) {
	f.liked = append(f.liked, id)

	return north.LikeState{Liked: true, LikeCount: 9}, nil, nil
}

func (f *fakeAPI) Unlike(_ context.Context, id string) (north.LikeState, *north.Response, error) {
	f.unliked = append(f.unliked, id)

	return north.LikeState{Liked: false, LikeCount: 8}, nil, nil
}

func (f *fakeAPI) Repost(_ context.Context, id string) (north.RepostState, *north.Response, error) {
	f.reposted = append(f.reposted, id)

	return north.RepostState{Reposted: true, RepostCount: 4}, nil, nil
}

func (f *fakeAPI) UndoRepost(_ context.Context, id string) (north.RepostState, *north.Response, error) {
	f.unreposted = append(f.unreposted, id)

	return north.RepostState{Reposted: false, RepostCount: 3}, nil, nil
}

func TestFeedLoadsFillsAndReacts(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	feed := New(service, ui.NewTheme())
	program := reactea.New(feed, reactea.WithSize(60, 18))
	program.Start()

	if len(feed.posts) != 3 || feed.SelectedPost().ID != "1" {
		t.Fatalf("initial posts = %#v, selected = %#v", feed.posts, feed.SelectedPost())
	}
	if len(service.homeCalls) != 2 || service.homeCalls[1].Cursor != "next" {
		t.Fatalf("fill calls = %#v", service.homeCalls)
	}
	frame := testkit.Plain(program)
	if !strings.Contains(frame, "first") || !strings.Contains(frame, "@user1") {
		t.Fatalf("initial frame missing post:\n%s", frame)
	}
	testkit.Click(program, 5, 5)
	if feed.SelectedPost().ID != "2" {
		t.Fatalf("selected after click = %s", feed.SelectedPost().ID)
	}

	testkit.SendKeys(program, "l", "t")
	target := feed.SelectedPost().DisplayPost()
	if !target.Liked || target.LikeCount != 9 {
		t.Errorf("like state = %#v", target)
	}
	if !target.Reposted || target.RepostCount != 4 {
		t.Errorf("repost state = %#v", target)
	}
	if len(service.liked) != 1 || service.liked[0] != "2" {
		t.Errorf("liked calls = %#v", service.liked)
	}
	if len(service.reposted) != 1 || service.reposted[0] != "2" {
		t.Errorf("repost calls = %#v", service.reposted)
	}

	testkit.SendKeys(program, "f", "t")
	if target.Liked || target.Reposted {
		t.Errorf("inverse reactions left state = %#v", target)
	}
	if len(service.unliked) != 1 || len(service.unreposted) != 1 {
		t.Errorf("inverse calls = unlike %#v, unrepost %#v", service.unliked, service.unreposted)
	}

	testkit.SendKeys(program, "space")
	if feed.SelectedPost().ID != "3" || len(service.liked) != 1 {
		t.Errorf("space selected = %s, likes = %#v", feed.SelectedPost().ID, service.liked)
	}
}

func TestFeedShowsNextPostAtBottom(t *testing.T) {
	t.Parallel()

	feed := New(&fakeAPI{}, ui.NewTheme())
	program := reactea.New(feed, reactea.WithSize(60, 5))
	program.Start()

	if frame := testkit.Plain(program); !strings.Contains(frame, "@user2") {
		t.Fatalf("next post was not drawn into the last row:\n%s", frame)
	}
}

func TestClickingUnselectedPostActionKeepsSelection(t *testing.T) {
	t.Parallel()

	feed := New(&fakeAPI{}, ui.NewTheme())
	program := reactea.New(feed, reactea.WithSize(60, 8))
	program.Start()
	selected, top := feed.selected, feed.top

	testkit.Click(program, 35, 6)
	if feed.selected != selected || feed.top != top {
		t.Fatalf("selection moved from %d/%d to %d/%d", selected, top, feed.selected, feed.top)
	}
}

func TestClickingReplyOnUnselectedPostMovesSelection(t *testing.T) {
	t.Parallel()

	feed := New(&fakeAPI{}, ui.NewTheme())
	program := reactea.New(feed, reactea.WithSize(60, 8))
	program.Start()

	testkit.Click(program, 5, 6)
	if post := feed.SelectedPost(); post == nil || post.ID != "2" {
		t.Fatalf("selected after reply click = %#v", post)
	}
}

func TestFeedLoadsOlderPostsBeforeReachingEnd(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{older: []north.Post{testPost("9", "loaded ahead")}}
	for index := 1; index <= 8; index++ {
		service.home = append(service.home, testPost(fmt.Sprint(index), "post"))
	}
	feed := New(service, ui.NewTheme())
	program := reactea.New(feed, reactea.WithSize(60, 30))
	program.Start()
	if len(service.homeCalls) != 1 {
		t.Fatalf("initial calls = %#v", service.homeCalls)
	}

	testkit.SendKeys(program, "down", "down", "down")
	if len(service.homeCalls) != 1 {
		t.Fatalf("loaded too early at post %d: %#v", feed.selected+1, service.homeCalls)
	}
	testkit.SendKeys(program, "down")
	if len(service.homeCalls) != 2 || service.homeCalls[1].Cursor != "next" {
		t.Fatalf("near-end calls = %#v", service.homeCalls)
	}
	if len(feed.posts) != 9 || feed.posts[8].ID != "9" {
		t.Fatalf("posts after prefetch = %#v", feed.posts)
	}
}

func TestScrollingAboveTopRefreshesLatestPosts(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	feed := New(service, ui.NewTheme())
	program := reactea.New(feed, reactea.WithSize(60, 5))
	program.Start()
	if len(service.homeCalls) != 1 {
		t.Fatalf("initial calls = %#v", service.homeCalls)
	}

	service.latest = []north.Post{testPost("0", "new"), testPost("1", "first")}
	testkit.SendKeys(program, "up")
	if len(service.homeCalls) != 2 || service.homeCalls[1].Cursor != "" {
		t.Fatalf("refresh calls = %#v", service.homeCalls)
	}
	if len(feed.posts) != 3 || feed.posts[0].ID != "0" || feed.posts[2].ID != "2" {
		t.Fatalf("refreshed posts = %#v", feed.posts)
	}
	testkit.Wheel(program, 5, 1, -3)
	if len(service.homeCalls) != 2 {
		t.Fatalf("one upward gesture made repeated requests: %#v", service.homeCalls)
	}
}

func TestResizeLoadsEnoughPostsForNewHeight(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	feed := New(service, ui.NewTheme())
	program := reactea.New(feed, reactea.WithSize(60, 3))
	program.Start()
	if len(service.homeCalls) != 1 {
		t.Fatalf("initial calls = %#v", service.homeCalls)
	}

	testkit.RenderAt(program, 60, 18)
	if len(service.homeCalls) != 2 || len(feed.posts) != 3 {
		t.Fatalf("calls after resize = %#v, posts = %#v", service.homeCalls, feed.posts)
	}
}

func TestLatePageCannotUndoLocalChanges(t *testing.T) {
	t.Parallel()

	feed := New(&fakeAPI{}, ui.NewTheme())
	feed.posts = []north.Post{testPost("1", "stale")}
	feed.seq = 7
	program := reactea.New(feed, reactea.WithSize(60, 4))

	program.Send(reactionMsg{
		target: feed,
		postID: "1",
		kind:   reactionLike,
		like:   north.LikeState{Liked: true, LikeCount: 9},
	})
	feed.RemovePost(program.Ctx(), "gone")
	program.Send(feedLoadedMsg{
		target: feed,
		seq:    7,
		page: north.PostPage{Items: []north.Post{
			testPost("1", "stale"),
			testPost("gone", "already deleted"),
		}},
	})

	if len(feed.posts) != 1 || feed.posts[0].ID != "1" {
		t.Fatalf("late page restored a deleted post: %#v", feed.posts)
	}
	target := feed.posts[0].DisplayPost()
	if !target.Liked || target.LikeCount != 9 {
		t.Fatalf("late page undid the like: %#v", target)
	}
}

func TestLatestPagePreservesTheTailCursor(t *testing.T) {
	t.Parallel()

	tail := "oldest-loaded-page"
	head := "second-page-from-the-top"
	feed := New(&fakeAPI{}, ui.NewTheme())
	feed.posts = []north.Post{testPost("1", "first"), testPost("2", "second")}
	feed.nextCursor = &tail
	feed.seq = 3
	program := reactea.New(feed, reactea.WithSize(60, 3))

	program.Send(feedLoadedMsg{
		target: feed,
		seq:    3,
		fresh:  true,
		page: north.PostPage{
			Items:      []north.Post{testPost("0", "new"), testPost("1", "first")},
			NextCursor: &head,
		},
	})

	if feed.nextCursor == nil || *feed.nextCursor != tail {
		t.Fatalf("tail cursor = %v, want %q", feed.nextCursor, tail)
	}
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
