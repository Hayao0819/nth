package user

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

type connectionAPI struct {
	calls []string
}

func (a *connectionAPI) Followers(_ context.Context, handle, cursor string) (north.UserPage, *north.Response, error) {
	a.calls = append(a.calls, "followers:"+handle+":"+cursor)
	if cursor != "" {
		return north.UserPage{Items: []north.User{{ID: "two", Handle: "bob", Name: "Bob"}}}, nil, nil
	}
	next := "next"

	return north.UserPage{Items: []north.User{{ID: "one", Handle: "alice", Name: "Alice"}}, NextCursor: &next}, nil, nil
}

func (a *connectionAPI) Following(_ context.Context, handle, cursor string) (north.UserPage, *north.Response, error) {
	a.calls = append(a.calls, "following:"+handle+":"+cursor)

	return north.UserPage{Items: []north.User{{ID: "three", Handle: "carol", Name: "Carol"}}}, nil, nil
}

func TestConnectionsLoadMoreAndOpenSelectedAccount(t *testing.T) {
	t.Parallel()

	api := &connectionAPI{}
	screen := NewConnections(api, ui.NewTheme(), north.User{Handle: "@owner"}, false)
	program := reactea.New(screen, reactea.WithSize(60, 14))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "Alice") || !strings.Contains(plain, "L more") {
		t.Fatalf("followers page:\n%s", plain)
	}
	testkit.SendKeys(program, "L")
	if len(screen.items) != 2 || screen.items[1].Handle != "bob" {
		t.Fatalf("items = %#v", screen.items)
	}
	testkit.SendKeys(program, "j")
	message, ok := screen.open()().(navigation.OpenUserMsg)
	if !ok || message.User.Handle != "bob" {
		t.Fatalf("open = %#v", message)
	}
	if got := strings.Join(api.calls, ","); got != "followers:owner:,followers:owner:next" {
		t.Fatalf("calls = %q", got)
	}
}

func TestFollowingUsesFollowingEndpoint(t *testing.T) {
	t.Parallel()

	api := &connectionAPI{}
	screen := NewConnections(api, ui.NewTheme(), north.User{Handle: "owner"}, true)
	program := reactea.New(screen, reactea.WithSize(60, 14))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "Following") || !strings.Contains(plain, "Carol") {
		t.Fatalf("following page:\n%s", plain)
	}
	if len(api.calls) != 1 || api.calls[0] != "following:owner:" {
		t.Fatalf("calls = %#v", api.calls)
	}
}
