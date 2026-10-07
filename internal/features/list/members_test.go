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

type memberAPI struct {
	users   []north.User
	actions []string
}

func (a *memberAPI) ListMembers(context.Context, string, string) (north.UserPage, *north.Response, error) {
	return north.UserPage{Items: append([]north.User(nil), a.users...)}, nil, nil
}

func (*memberAPI) ListMemberships(context.Context, string, string) ([]north.ListMembership, *north.Response, error) {
	return nil, nil, nil
}

func (a *memberAPI) AddListMember(_ context.Context, listID, handle string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "add:"+listID+":"+handle)

	return true, nil, nil
}

func (a *memberAPI) RemoveListMember(_ context.Context, listID, handle string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "remove:"+listID+":"+handle)

	return true, nil, nil
}

func TestMembersLoadOpenAddAndRemove(t *testing.T) {
	t.Parallel()

	api := &memberAPI{users: []north.User{{ID: "user", Handle: "alice", Name: "Alice"}}}
	screen := NewMembers(api, ui.NewTheme(), north.List{ID: "list", Name: "Friends", OwnedByViewer: true})
	program := reactea.New(screen, reactea.WithSize(64, 16))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "Alice") || !strings.Contains(plain, "a add") {
		t.Fatalf("members page:\n%s", plain)
	}
	opened, ok := screen.open()().(navigation.OpenUserMsg)
	if !ok || opened.User.Handle != "alice" {
		t.Fatalf("open = %#v", opened)
	}
	added, ok := screen.add(context.Background(), " @bob ")().(memberChangedMsg)
	if !ok || added.err != nil || len(api.actions) != 1 || api.actions[0] != "add:list:bob" {
		t.Fatalf("add = %#v, actions %#v", added, api.actions)
	}
	screen.acting = false
	removed, ok := screen.remove(context.Background())().(memberChangedMsg)
	if !ok || removed.err != nil || len(api.actions) != 2 || api.actions[1] != "remove:list:alice" {
		t.Fatalf("remove = %#v, actions %#v", removed, api.actions)
	}
	screen.applyMemberChanged(context.Background(), removed)
	if len(screen.users) != 0 {
		t.Fatalf("users after removal = %#v", screen.users)
	}
}

func TestReadOnlyListHidesMemberEditing(t *testing.T) {
	t.Parallel()

	screen := NewMembers(&memberAPI{}, ui.NewTheme(), north.List{ID: "list", Name: "Public"})
	program := reactea.New(screen, reactea.WithSize(60, 14))
	program.Start()
	if plain := testkit.Plain(program); strings.Contains(plain, "a add") || screen.promptAdd(program.Ctx()) != nil {
		t.Fatalf("read-only list exposes editing:\n%s", plain)
	}
}
