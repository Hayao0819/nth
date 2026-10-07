package list

import (
	"context"
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type listAPI struct {
	collection north.ListCollection
	actions    []string
}

func (a *listAPI) Lists(context.Context, string) (north.ListCollection, *north.Response, error) {
	return a.collection, nil, nil
}

func (*listAPI) List(context.Context, string) (north.List, *north.Response, error) {
	return north.List{}, nil, nil
}

func (*listAPI) ListTimeline(context.Context, string, string) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (a *listAPI) FollowList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "follow:"+id)

	return true, nil, nil
}

func (a *listAPI) UnfollowList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "unfollow:"+id)

	return true, nil, nil
}

func (a *listAPI) PinList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "pin:"+id)

	return true, nil, nil
}

func (a *listAPI) UnpinList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "unpin:"+id)

	return true, nil, nil
}

func TestListsAreGroupedDeduplicatedAndActionable(t *testing.T) {
	t.Parallel()

	owner := north.User{Handle: "alice", Name: "Alice"}
	pinned := north.List{ID: "pinned", Name: "Pinned", Owner: owner, PinnedByViewer: true, FollowedByViewer: true}
	followed := north.List{ID: "followed", Name: "Followed", Owner: owner, FollowedByViewer: true}
	api := &listAPI{collection: north.ListCollection{
		Pinned:   []north.List{pinned},
		Items:    []north.List{pinned},
		Followed: []north.List{followed},
	}}
	screen := NewIndex(api, ui.NewTheme())
	program := reactea.New(screen, reactea.WithSize(72, 18))
	program.Start()

	if len(screen.entries) != 2 {
		t.Fatalf("entries = %#v", screen.entries)
	}
	plain := testkit.Plain(program)
	for _, want := range []string{"Pinned", "Pinned · 0 members", "Followed", "Following · 0 members"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("list page is missing %q:\n%s", want, plain)
		}
	}

	testkit.SendKeys(program, "f")
	if len(api.actions) != 1 || api.actions[0] != "unfollow:pinned" {
		t.Fatalf("actions = %#v", api.actions)
	}
	testkit.SendKeys(program, "p")
	if len(api.actions) != 2 || api.actions[1] != "unpin:pinned" {
		t.Fatalf("actions = %#v", api.actions)
	}
	message, ok := screen.open()().(navigation.OpenListMsg)
	if !ok || message.List.ID != "pinned" {
		t.Fatalf("open = %#v", message)
	}
}

func TestOwnedListCannotBeFollowed(t *testing.T) {
	t.Parallel()

	api := &listAPI{collection: north.ListCollection{Items: []north.List{{ID: "mine", Name: "Mine", OwnedByViewer: true}}}}
	screen := NewIndex(api, ui.NewTheme())
	program := reactea.New(screen, reactea.WithSize(60, 14))
	program.Start()
	testkit.SendKeys(program, "f")

	if len(api.actions) != 0 || !strings.Contains(testkit.Plain(program), "You own this list") {
		t.Fatalf("actions = %#v\n%s", api.actions, testkit.Plain(program))
	}
}
