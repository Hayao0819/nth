package list

import (
	"context"
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type detailAPI struct {
	actions []string
	updated north.UpdateListRequest
}

func (*detailAPI) HomeTimeline(context.Context, north.TimelineOptions) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (*detailAPI) SearchPosts(context.Context, string, north.SearchOptions) (north.PostPage, *north.Response, error) {
	return north.PostPage{}, nil, nil
}

func (*detailAPI) Like(context.Context, string) (north.LikeState, *north.Response, error) {
	return north.LikeState{}, nil, nil
}

func (*detailAPI) Unlike(context.Context, string) (north.LikeState, *north.Response, error) {
	return north.LikeState{}, nil, nil
}

func (*detailAPI) Repost(context.Context, string) (north.RepostState, *north.Response, error) {
	return north.RepostState{}, nil, nil
}

func (*detailAPI) UndoRepost(context.Context, string) (north.RepostState, *north.Response, error) {
	return north.RepostState{}, nil, nil
}

func (*detailAPI) Lists(context.Context, string) (north.ListCollection, *north.Response, error) {
	return north.ListCollection{}, nil, nil
}

func (*detailAPI) List(_ context.Context, id string) (north.List, *north.Response, error) {
	return north.List{ID: id, Name: "Friends"}, nil, nil
}

func (*detailAPI) ListTimeline(context.Context, string, string) (north.PostPage, *north.Response, error) {
	return north.PostPage{Items: []north.Post{{ID: "post", Text: "list post", Author: north.User{Handle: "alice"}}}}, nil, nil
}

func (a *detailAPI) FollowList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "follow:"+id)

	return true, nil, nil
}

func (a *detailAPI) UnfollowList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "unfollow:"+id)

	return true, nil, nil
}

func (a *detailAPI) PinList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "pin:"+id)

	return true, nil, nil
}

func (a *detailAPI) UnpinList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "unpin:"+id)

	return true, nil, nil
}

func (*detailAPI) CreateList(context.Context, north.CreateListRequest) (north.List, *north.Response, error) {
	return north.List{}, nil, nil
}

func (a *detailAPI) UpdateList(_ context.Context, id string, request north.UpdateListRequest) (north.List, *north.Response, error) {
	a.actions = append(a.actions, "update:"+id)
	a.updated = request
	name := ""
	if request.Name != nil {
		name = *request.Name
	}

	return north.List{ID: id, Name: name, OwnedByViewer: true}, nil, nil
}

func (a *detailAPI) DeleteList(_ context.Context, id string) (bool, *north.Response, error) {
	a.actions = append(a.actions, "delete:"+id)

	return true, nil, nil
}

func (*detailAPI) ListMembers(context.Context, string, string) (north.UserPage, *north.Response, error) {
	return north.UserPage{}, nil, nil
}

func (*detailAPI) ListMemberships(context.Context, string, string) ([]north.ListMembership, *north.Response, error) {
	return nil, nil, nil
}

func (*detailAPI) AddListMember(context.Context, string, string) (bool, *north.Response, error) {
	return true, nil, nil
}

func (*detailAPI) RemoveListMember(context.Context, string, string) (bool, *north.Response, error) {
	return true, nil, nil
}

func TestListPageLoadsTimelineAndTogglesState(t *testing.T) {
	t.Parallel()

	api := &detailAPI{}
	screen := NewDetail(api, api, ui.NewTheme(), north.List{ID: "list", Name: "Friends", Owner: north.User{Handle: "owner"}}, nil)
	program := reactea.New(screen, reactea.WithSize(72, 20))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "Friends") || !strings.Contains(plain, "list post") {
		t.Fatalf("list page:\n%s", plain)
	}
	testkit.SendKeys(program, "f", "p")
	if got := strings.Join(api.actions, ","); got != "follow:list,pin:list" || !screen.item.FollowedByViewer || !screen.item.PinnedByViewer {
		t.Fatalf("actions = %q, item = %#v", got, screen.item)
	}
}

func TestOwnedListCanBeUpdatedAndDeleted(t *testing.T) {
	t.Parallel()

	api := &detailAPI{}
	screen := NewDetail(api, api, ui.NewTheme(), north.List{ID: "list", Name: "Old", OwnedByViewer: true}, nil)
	result := dialog.FormResult{
		Values:  map[string]string{"name": "New", "description": "Description"},
		Toggles: map[string]bool{"private": true},
	}
	message, ok := screen.updateList(context.Background(), result)().(editedMsg)
	if !ok || message.err != nil {
		t.Fatalf("update = %#v", message)
	}
	screen.applyEdited(message)
	if screen.item.Name != "New" || api.updated.Name == nil || *api.updated.Name != "New" {
		t.Fatalf("updated item = %#v, request = %#v", screen.item, api.updated)
	}
	screen.acting = false
	deleted, ok := screen.deleteList(context.Background())().(deletedMsg)
	if !ok || deleted.err != nil || api.actions[len(api.actions)-1] != "delete:list" {
		t.Fatalf("delete = %#v, actions = %#v", deleted, api.actions)
	}
}
