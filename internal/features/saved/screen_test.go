package saved

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type savedAPI struct {
	calls        []string
	draftRequest north.DraftRequest
	pendingID    string
}

func (*savedAPI) Drafts(context.Context, string) ([]north.Draft, *north.Response, error) {
	return []north.Draft{{ID: "draft", Items: []north.DraftItem{{Text: "unfinished"}}}}, nil, nil
}

func (a *savedAPI) CreateDraft(_ context.Context, request north.DraftRequest) (north.Draft, *north.Response, error) {
	a.calls = append(a.calls, "create-draft")
	a.draftRequest = request

	return north.Draft{ID: "created", Items: []north.DraftItem{{Text: request.Items[0].Text}}}, nil, nil
}

func (a *savedAPI) UpdateDraft(_ context.Context, id string, request north.DraftRequest) (north.Draft, *north.Response, error) {
	a.calls = append(a.calls, "update-draft:"+id)
	a.draftRequest = request

	return north.Draft{ID: id}, nil, nil
}

func (a *savedAPI) DeleteDrafts(_ context.Context, ids ...string) (int, *north.Response, error) {
	a.calls = append(a.calls, "delete-draft:"+strings.Join(ids, ","))

	return len(ids), nil, nil
}

func (*savedAPI) ScheduledPosts(context.Context, string) ([]north.ScheduledPost, *north.Response, error) {
	return []north.ScheduledPost{{Draft: north.Draft{ID: "scheduled", Items: []north.DraftItem{{Text: "later"}}}, ScheduledAt: time.Now().Add(time.Hour)}}, nil, nil
}

func (a *savedAPI) SchedulePost(_ context.Context, request north.SchedulePostRequest) (north.ScheduledPost, *north.Response, error) {
	a.calls = append(a.calls, "schedule")

	return north.ScheduledPost{Draft: north.Draft{ID: "scheduled"}, ScheduledAt: request.ScheduledAt}, nil, nil
}

func (a *savedAPI) UpdateScheduledPost(_ context.Context, id string, request north.SchedulePostRequest) (north.ScheduledPost, *north.Response, error) {
	a.calls = append(a.calls, "update-schedule:"+id)

	return north.ScheduledPost{Draft: north.Draft{ID: id}, ScheduledAt: request.ScheduledAt}, nil, nil
}

func (a *savedAPI) DeleteScheduledPosts(_ context.Context, ids ...string) (int, *north.Response, error) {
	a.calls = append(a.calls, "delete-schedule:"+strings.Join(ids, ","))

	return len(ids), nil, nil
}

func (*savedAPI) PendingPosts(context.Context, string) ([]north.PendingPost, *north.Response, error) {
	return []north.PendingPost{{Draft: north.Draft{ID: "pending", Items: []north.DraftItem{{Text: "queued"}}}, UndoUntil: time.Now().Add(time.Minute)}}, nil, nil
}

func (a *savedAPI) CreatePendingPost(_ context.Context, request north.CreatePendingPostRequest) (north.PendingPost, *north.Response, error) {
	a.calls = append(a.calls, "create-pending:"+request.Item.Text)

	return north.PendingPost{Draft: north.Draft{ID: "pending"}}, nil, nil
}

func (a *savedAPI) SendPendingPost(_ context.Context, id string) (north.PendingPostState, *north.Response, error) {
	a.calls = append(a.calls, "send:"+id)
	a.pendingID = id

	return north.PendingPostState{Sent: true}, nil, nil
}

func (a *savedAPI) UndoPendingPost(_ context.Context, id string) (north.Draft, *north.Response, error) {
	a.calls = append(a.calls, "undo:"+id)
	a.pendingID = id

	return north.Draft{ID: id}, nil, nil
}

func TestSavedPageLoadsEachTabAndSendsPendingPost(t *testing.T) {
	t.Parallel()

	api := &savedAPI{}
	screen := New(api, ui.NewTheme())
	program := reactea.New(screen, reactea.WithSize(76, 18))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "unfinished") || !strings.Contains(plain, "Draft") {
		t.Fatalf("draft tab:\n%s", plain)
	}
	testkit.SendKeys(program, "tab")
	if plain := testkit.Plain(program); !strings.Contains(plain, "later") || !strings.Contains(plain, "Scheduled for") {
		t.Fatalf("scheduled tab:\n%s", plain)
	}
	testkit.SendKeys(program, "tab", "s")
	if api.pendingID != "pending" || !strings.Contains(strings.Join(api.calls, ","), "send:pending") {
		t.Fatalf("calls = %#v", api.calls)
	}
}

func TestSavedPagePreservesDraftContextAndMedia(t *testing.T) {
	t.Parallel()

	reply := north.Post{ID: "parent"}
	quote := north.Post{ID: "quoted"}
	draft := north.Draft{
		ID:            "draft",
		Items:         []north.DraftItem{{Text: "old", ReplyPolicy: north.ReplyPolicy("FOLLOWING"), Media: []north.Media{{ID: "media"}}}},
		QuotedPost:    &quote,
		InReplyToPost: &reply,
	}
	request := draftRequest(&draft, "new")
	if len(request.Items) != 1 || request.Items[0].Text != "new" || len(request.Items[0].MediaIDs) != 1 || request.Items[0].MediaIDs[0] != "media" {
		t.Fatalf("request items = %#v", request.Items)
	}
	if request.QuotedID == nil || request.InReplyToID == nil {
		t.Fatalf("draft context was lost: %#v", request)
	}

	screen := New(&savedAPI{}, ui.NewTheme())
	result := dialog.FormResult{Values: map[string]string{scheduleTextKey: "later", scheduleTimeKey: "not a time"}}
	if command := screen.saveSchedule(context.Background(), result); command != nil || !strings.Contains(screen.notice, "YYYY-MM-DD") {
		t.Fatalf("invalid schedule = command %#v, notice %q", command, screen.notice)
	}
}
