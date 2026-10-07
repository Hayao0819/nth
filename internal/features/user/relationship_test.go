package user

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type relationshipProfileAPI struct {
	*profileAPI
	relationErr   error
	relationCalls []string
}

func (a *relationshipProfileAPI) relationship(name, handle string) (bool, *north.Response, error) {
	a.relationCalls = append(a.relationCalls, name+":"+handle)

	return a.relationErr == nil, nil, a.relationErr
}

func (a *relationshipProfileAPI) Follow(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("follow", handle)
}

func (a *relationshipProfileAPI) Unfollow(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("unfollow", handle)
}

func (a *relationshipProfileAPI) Block(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("block", handle)
}

func (a *relationshipProfileAPI) Unblock(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("unblock", handle)
}

func (a *relationshipProfileAPI) Mute(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("mute", handle)
}

func (a *relationshipProfileAPI) Unmute(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("unmute", handle)
}

func (a *relationshipProfileAPI) EnablePostNotifications(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("notify", handle)
}

func (a *relationshipProfileAPI) DisablePostNotifications(_ context.Context, handle string) (bool, *north.Response, error) {
	return a.relationship("unnotify", handle)
}

func TestPageChangesRelationshipsWithKeyboard(t *testing.T) {
	t.Parallel()

	api := &relationshipProfileAPI{profileAPI: &profileAPI{
		user: north.User{ID: "user-1", Handle: "alice", Name: "Alice", FollowerCount: 4, FollowingCount: 3},
	}}
	screen := NewPage(api, ui.NewTheme(), north.User{Handle: "alice"})
	screen.SetViewer(north.User{ID: "viewer-1", Handle: "viewer"})
	program := reactea.New(modal.New(screen), reactea.WithSize(64, 24))
	program.Start()

	plain := testkit.Plain(program)
	for _, label := range []string{"F  Follow", "B  Block", "M  Mute", "N  Notify"} {
		if !strings.Contains(plain, label) {
			t.Fatalf("relationship action %q is missing:\n%s", label, plain)
		}
	}

	testkit.SendKeys(program, "F")
	if plain = testkit.Plain(program); !strings.Contains(plain, "F  Following") || !strings.Contains(plain, "Following @alice") {
		t.Fatalf("follow state was not rendered:\n%s", plain)
	}
	testkit.SendKeys(program, "F")
	if plain = testkit.Plain(program); !strings.Contains(plain, "F  Follow") || !strings.Contains(plain, "Unfollowed @alice") {
		t.Fatalf("unfollow state was not rendered:\n%s", plain)
	}

	testkit.SendKeys(program, "M")
	if plain = testkit.Plain(program); !strings.Contains(plain, "M  Unmute") || !strings.Contains(plain, "Account muted") {
		t.Fatalf("mute state was not rendered:\n%s", plain)
	}
	testkit.SendKeys(program, "M")
	if plain = testkit.Plain(program); !strings.Contains(plain, "M  Mute") || !strings.Contains(plain, "Account unmuted") {
		t.Fatalf("unmute state was not rendered:\n%s", plain)
	}

	testkit.SendKeys(program, "N")
	if plain = testkit.Plain(program); !strings.Contains(plain, "N  Notifying") || !strings.Contains(plain, "Post notifications enabled") {
		t.Fatalf("post notification state was not rendered:\n%s", plain)
	}
	testkit.SendKeys(program, "N")
	if plain = testkit.Plain(program); !strings.Contains(plain, "N  Notify") || !strings.Contains(plain, "Post notifications disabled") {
		t.Fatalf("disabled post notification state was not rendered:\n%s", plain)
	}

	testkit.SendKeys(program, "B")
	if plain = testkit.Plain(program); !strings.Contains(plain, "Block @alice?") {
		t.Fatalf("block confirmation is missing:\n%s", plain)
	}
	if len(api.relationCalls) != 6 {
		t.Fatalf("block ran before confirmation: %#v", api.relationCalls)
	}
	testkit.SendKeys(program, "enter")
	if plain = testkit.Plain(program); !strings.Contains(plain, "B  Unblock") || !strings.Contains(plain, "Account blocked") {
		t.Fatalf("block state was not rendered:\n%s", plain)
	}
	testkit.SendKeys(program, "B")
	if plain = testkit.Plain(program); !strings.Contains(plain, "B  Block") || !strings.Contains(plain, "Account unblocked") {
		t.Fatalf("unblock state was not rendered:\n%s", plain)
	}

	want := []string{
		"follow:alice",
		"unfollow:alice",
		"mute:alice",
		"unmute:alice",
		"notify:alice",
		"unnotify:alice",
		"block:alice",
		"unblock:alice",
	}
	if got := strings.Join(api.relationCalls, ","); got != strings.Join(want, ",") {
		t.Fatalf("relationship calls = %q", got)
	}
}

func TestPageDoesNotShowRelationshipActionsForViewer(t *testing.T) {
	t.Parallel()

	api := &relationshipProfileAPI{profileAPI: &profileAPI{
		user: north.User{ID: "user-1", Handle: "alice", Name: "Alice"},
	}}
	screen := NewPage(api, ui.NewTheme(), north.User{Handle: "alice"})
	screen.SetViewer(north.User{ID: "user-1", Handle: "alice"})
	program := reactea.New(screen, reactea.WithSize(64, 20))
	program.Start()

	plain := testkit.Plain(program)
	for _, label := range []string{"F  Follow", "B  Block", "M  Mute", "N  Notify"} {
		if strings.Contains(plain, label) {
			t.Fatalf("own profile contains %q:\n%s", label, plain)
		}
	}
}

func TestRelationshipButtonsAreClickable(t *testing.T) {
	t.Parallel()

	api := &relationshipProfileAPI{profileAPI: &profileAPI{
		user: north.User{ID: "user-1", Handle: "alice", Name: "Alice"},
	}}
	screen := NewPage(api, ui.NewTheme(), north.User{Handle: "alice"})
	screen.SetViewer(north.User{ID: "viewer-1", Handle: "viewer"})
	program := reactea.New(screen, reactea.WithSize(64, 20))
	program.Start()

	positions := testkit.Find(program, "M  Mute")
	if len(positions) == 0 {
		t.Fatalf("mute action is missing:\n%s", testkit.Plain(program))
	}
	point := positions[0]
	command := screen.Update(program.Ctx(), tea.MouseClickMsg{X: point.X, Y: point.Y, Button: tea.MouseLeft})
	if command == nil {
		t.Fatal("clicking mute returned no command")
	}
	message := command()
	result, ok := message.(relationshipResultMsg)
	if !ok {
		t.Fatalf("mute command returned %T", message)
	}
	screen.Update(program.Ctx(), result)
	if len(api.relationCalls) != 1 || api.relationCalls[0] != "mute:alice" || !screen.user.Muting {
		t.Fatalf("mute click = calls %#v, user %#v", api.relationCalls, screen.user)
	}
}
