package app

import (
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestRoutePartsRoundTrip(t *testing.T) {
	t.Parallel()

	query := "Go / 北 #north"
	search := searchRoute(query)
	if !strings.HasPrefix(search, "/search/") {
		t.Fatalf("search route = %q", search)
	}
	if got := unescapeRoutePart(strings.TrimPrefix(search, "/search/")); got != query {
		t.Fatalf("decoded query = %q, want %q", got, query)
	}
	if state := pageStateForRoute(postRoute("post/with slash")); state.kind != postPage || state.key != "post/with slash" {
		t.Fatalf("post route state = %#v", state)
	}
	if state := pageStateForRoute(userRoute("Alice")); state.kind != profilePage || state.key != "alice" {
		t.Fatalf("user route state = %#v", state)
	}
}

func TestDirectPostRouteLoadsThroughResource(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{posts: map[string]north.Post{
		"direct": testPost("direct", "loaded from its route"),
	}}
	root := newRoot(service)
	program := reactea.New(
		modal.New(root),
		reactea.WithRoute(postRoute("direct")),
		reactea.WithSize(80, 24),
	)
	program.Start()

	if plain := testkit.Plain(program); !strings.Contains(plain, "loaded from its route") {
		t.Fatalf("direct post route did not load:\n%s", plain)
	}
	if root.page.kind != postPage || root.page.key != "direct" {
		t.Fatalf("page state = %#v", root.page)
	}
	if cached, ok := root.posts["direct"]; !ok || cached.ID != "direct" {
		t.Fatalf("post cache = %#v", root.posts)
	}
}

func TestBackRestoresTheCurrentSearchQuery(t *testing.T) {
	t.Parallel()

	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	testkit.SendKeys(program, "/", "g", "o", "enter")
	if root.page.kind != postPage {
		t.Fatalf("search result did not open: page=%v route=%q", root.page.kind, program.Route())
	}

	testkit.SendKeys(program, "esc")
	search, ok := root.currentPage().(interface{ Query() string })
	if !ok {
		t.Fatalf("restored search = %T", root.currentPage())
	}
	if search.Query() != "go" {
		t.Fatalf("restored query = %q", search.Query())
	}
	if program.Route() != searchRoute("go") {
		t.Fatalf("restored route = %q", program.Route())
	}
}
