package search

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestSearchRunsAfterInputSettles(t *testing.T) {
	t.Parallel()

	api := &searchAPI{}
	screen := New(api, ui.NewTheme(), "")
	program := reactea.New(screen, reactea.WithSize(72, 20))
	program.Start()
	program.Send(tea.KeyPressMsg{Code: 'n', Text: "north"})

	if screen.Query() != "north" || len(api.queries) != 1 || api.queries[0] != "north" {
		t.Fatalf("query = %q, calls = %#v", screen.Query(), api.queries)
	}
	plain := testkit.Plain(program)
	lines := strings.Split(plain, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "" || !strings.Contains(lines[1], "north") || strings.TrimSpace(lines[2]) != "" {
		t.Fatalf("search header does not use three rows:\n%s", plain)
	}
}

func TestSearchIgnoresStaleDebounce(t *testing.T) {
	t.Parallel()

	api := &searchAPI{}
	screen := New(api, ui.NewTheme(), "")
	app := reactea.New(screen, reactea.WithSize(72, 20))
	screen.seq = 2
	if command := screen.Update(app.Ctx(), debounceMsg{target: screen, seq: 1, query: "old"}); command != nil {
		t.Fatal("stale debounce returned a command")
	}
	command := screen.Update(app.Ctx(), debounceMsg{target: screen, seq: 2, query: "new"})
	if command == nil {
		t.Fatal("current debounce did not start a search")
	}
	loaded := command()
	screen.Update(app.Ctx(), loaded)
	if len(api.queries) != 1 || api.queries[0] != "new" {
		t.Fatalf("search calls = %#v", api.queries)
	}
}

func TestClickingRenderedSearchResultUsesItsDisplayedCoordinates(t *testing.T) {
	t.Parallel()

	api := &searchAPI{}
	screen := New(api, ui.NewTheme(), "north")
	program := reactea.New(screen, reactea.WithSize(72, 20))
	program.Start()
	positions := testkit.Find(program, "result: north")
	if len(positions) == 0 {
		t.Fatalf("search result was not rendered:\n%s", testkit.Plain(program))
	}
	point := positions[0]
	command := screen.Update(program.Ctx(), tea.MouseClickMsg{X: point.X, Y: point.Y, Button: tea.MouseLeft})
	if command == nil {
		t.Fatal("clicking a search result returned no command")
	}
	message, ok := command().(navigation.OpenPostMsg)
	if !ok || message.Post.ID != "result" {
		t.Fatalf("open result = %#v", message)
	}
}

type searchAPI struct{ queries []string }

func (s *searchAPI) SearchPosts(_ context.Context, query string, _ north.SearchOptions) (north.PostPage, *north.Response, error) {
	s.queries = append(s.queries, query)

	return north.PostPage{Items: []north.Post{{ID: "result", Text: "result: " + query}}}, nil, nil
}

func (*searchAPI) HomeTimeline(context.Context, north.TimelineOptions) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (*searchAPI) Mentions(context.Context, string, string) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (*searchAPI) UserPosts(context.Context, string, string) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (*searchAPI) Like(context.Context, string) (north.LikeState, *north.Response, error) {
	return north.LikeState{}, nil, nil
}

func (*searchAPI) Unlike(context.Context, string) (north.LikeState, *north.Response, error) {
	return north.LikeState{}, nil, nil
}

func (*searchAPI) Repost(context.Context, string) (north.RepostState, *north.Response, error) {
	return north.RepostState{}, nil, nil
}

func (*searchAPI) UndoRepost(context.Context, string) (north.RepostState, *north.Response, error) {
	return north.RepostState{}, nil, nil
}

var _ feed.API = (*searchAPI)(nil)
