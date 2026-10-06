package bookmark

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain/bookmark"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type bookmarkAPI struct {
	posts []north.Post
}

func (a *bookmarkAPI) Bookmarks(context.Context, string) (north.PostPage, *north.Response, error) {
	return north.PostPage{Items: append([]north.Post(nil), a.posts...)}, nil, nil
}

func (*bookmarkAPI) HomeTimeline(context.Context, north.TimelineOptions) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (*bookmarkAPI) SearchPosts(context.Context, string, north.SearchOptions) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (*bookmarkAPI) Like(context.Context, string) (north.LikeState, *north.Response, error) {
	return north.LikeState{}, nil, nil
}

func (*bookmarkAPI) Unlike(context.Context, string) (north.LikeState, *north.Response, error) {
	return north.LikeState{}, nil, nil
}

func (*bookmarkAPI) Repost(context.Context, string) (north.RepostState, *north.Response, error) {
	return north.RepostState{}, nil, nil
}

func (*bookmarkAPI) UndoRepost(context.Context, string) (north.RepostState, *north.Response, error) {
	return north.RepostState{}, nil, nil
}

func TestRenderedBookmarkTargetsUseTheirDisplayedCoordinates(t *testing.T) {
	t.Parallel()

	api := &bookmarkAPI{posts: []north.Post{
		{ID: "first", Text: "first bookmark", Author: north.User{Name: "Alice", Handle: "alice"}},
		{ID: "second", Text: "second bookmark", Author: north.User{Name: "Bob", Handle: "bob"}},
	}}
	screen := New(api, api, ui.NewTheme(), nil)
	program := reactea.New(screen, reactea.WithSize(70, 18))
	program.Start()
	testkit.SendKeys(program, "down")

	if !testkit.ClickText(program, "Alice") {
		t.Fatal("Alice was not rendered")
	}
	if selected := screen.feed.SelectedPost(); selected == nil || selected.ID != "first" {
		t.Fatalf("selected bookmark = %#v", selected)
	}

	positions := testkit.Find(program, "♡")
	if len(positions) == 0 {
		t.Fatal("like action was not rendered")
	}
	point := positions[0]
	command := screen.Update(program.Ctx(), tea.MouseClickMsg{X: point.X, Y: point.Y, Button: tea.MouseLeft})
	if command == nil {
		t.Fatal("clicking the like action returned no command")
	}
	message, ok := command().(postcomponent.ActionMsg)
	if !ok || message.Action != postcomponent.Like || message.Post.ID != "first" {
		t.Fatalf("like action = %#v", message)
	}
}

var _ feed.API = (*bookmarkAPI)(nil)
var _ bookmarkdomain.API = (*bookmarkAPI)(nil)
