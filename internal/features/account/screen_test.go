package account

import (
	"context"
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type safetyAPI struct {
	actions []string
}

func (*safetyAPI) FollowRequests(context.Context, string) (north.UserPage, *north.Response, error) {
	return north.UserPage{Items: []north.User{{ID: "request", Handle: "alice", Name: "Alice"}}}, nil, nil
}

func (a *safetyAPI) AcceptFollowRequest(_ context.Context, handle string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "accept:"+handle)

	return true, nil, nil
}

func (a *safetyAPI) RejectFollowRequest(_ context.Context, handle string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "reject:"+handle)

	return true, nil, nil
}

func (*safetyAPI) BlockedUsers(context.Context, string) (north.UserPage, *north.Response, error) {
	return north.UserPage{Items: []north.User{{ID: "blocked", Handle: "bob", Name: "Bob"}}}, nil, nil
}

func (*safetyAPI) MutedUsers(context.Context, string) (north.UserPage, *north.Response, error) {
	return north.UserPage{Items: []north.User{{ID: "muted", Handle: "carol", Name: "Carol"}}}, nil, nil
}

func (a *safetyAPI) Unblock(_ context.Context, handle string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "unblock:"+handle)

	return true, nil, nil
}

func (a *safetyAPI) Unmute(_ context.Context, handle string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "unmute:"+handle)

	return true, nil, nil
}

type keywordAPI struct {
	created north.CreateMutedKeywordRequest
	deleted string
}

func (*keywordAPI) MutedKeywords(context.Context, string) ([]north.MutedKeyword, *north.Response, error) {
	return []north.MutedKeyword{{ID: "word", Word: "spoiler", Home: true}}, nil, nil
}

func (a *keywordAPI) CreateMutedKeyword(_ context.Context, request north.CreateMutedKeywordRequest) (north.MutedKeyword, *north.Response, error) {
	a.created = request

	return north.MutedKeyword{ID: "created", Word: request.Word}, nil, nil
}

func (a *keywordAPI) DeleteMutedKeyword(_ context.Context, id string) (bool, *north.Response, error) {
	a.deleted = id

	return true, nil, nil
}

func TestPrivacyPageLoadsTabsAndAppliesActions(t *testing.T) {
	t.Parallel()

	api := &safetyAPI{}
	screen := New(api, nil, ui.NewTheme())
	program := reactea.New(modal.New(screen), reactea.WithSize(72, 18))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "Alice") || !strings.Contains(plain, "y accept") {
		t.Fatalf("requests tab:\n%s", plain)
	}
	testkit.SendKeys(program, "y")
	if len(api.actions) != 1 || api.actions[0] != "accept:alice" {
		t.Fatalf("actions = %#v", api.actions)
	}
	testkit.SendKeys(program, "tab")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Bob") || !strings.Contains(plain, "x unblock") {
		t.Fatalf("blocked tab:\n%s", plain)
	}
}

func TestMutedKeywordRequestIsValidated(t *testing.T) {
	t.Parallel()

	keywords := &keywordAPI{}
	screen := New(&safetyAPI{}, keywords, ui.NewTheme())
	bad := dialog.FormResult{Values: map[string]string{keywordWordKey: "spoiler", keywordDurationKey: "later"}}
	if command := screen.createKeyword(context.Background(), bad); command != nil || !strings.Contains(screen.notice, "forever") {
		t.Fatalf("invalid duration = command %#v, notice %q", command, screen.notice)
	}
	good := dialog.FormResult{
		Values:  map[string]string{keywordWordKey: " spoiler ", keywordDurationKey: "24H"},
		Toggles: map[string]bool{keywordHomeKey: true, keywordNotificationsKey: true, keywordAnyoneKey: true},
	}
	message, ok := screen.createKeyword(context.Background(), good)().(changedMsg)
	if !ok || message.err != nil {
		t.Fatalf("create result = %#v", message)
	}
	if keywords.created.Word != "spoiler" || keywords.created.Duration != north.Mute24Hours || !keywords.created.Notifications {
		t.Fatalf("request = %#v", keywords.created)
	}
}
