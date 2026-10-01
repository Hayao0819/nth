package notification

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	notificationdomain "github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type notificationAPI struct {
	pages    map[string]notificationdomain.Page
	calls    []string
	readCall int
	readErr  error
	loadErr  error
}

func (a *notificationAPI) Notifications(_ context.Context, _ north.NotificationTab, cursor string) (notificationdomain.Page, *north.Response, error) {
	a.calls = append(a.calls, cursor)

	return a.pages[cursor], nil, a.loadErr
}

func (a *notificationAPI) NotificationUnreadCount(context.Context) (int, *north.Response, error) {
	return 0, nil, nil
}

func (a *notificationAPI) MarkNotificationsRead(context.Context) (int, *north.Response, error) {
	a.readCall++

	return 0, nil, a.readErr
}

func TestPageLoadsMarksReadAndPaginates(t *testing.T) {
	t.Parallel()

	next := "next"
	api := &notificationAPI{pages: map[string]notificationdomain.Page{
		"": {
			Items: []notificationdomain.Item{{
				ID:         "like-1",
				Kind:       notificationdomain.Like,
				Actors:     []north.User{{Name: "Alice", Handle: "alice"}},
				ActorCount: 2,
				CreatedAt:  time.Now().Add(-time.Minute),
				Post:       &north.Post{ID: "post-1", Text: "hello"},
			}},
			NextCursor: &next,
		},
		"next": {
			Items: []notificationdomain.Item{{
				ID: "follow-1", Kind: notificationdomain.Follow,
				Actors: []north.User{{Name: "Bob", Handle: "bob"}},
			}},
		},
	}}
	page := NewPage(api, ui.NewTheme())
	program := reactea.New(page, reactea.WithSize(70, 18))
	program.Start()

	plain := testkit.Plain(program)
	for _, want := range []string{"Notifications", "Alice +1 liked your post", "hello"} {
		if !strings.Contains(plain, want) {
			t.Errorf("notification page missing %q:\n%s", want, plain)
		}
	}
	if api.readCall != 1 || !page.markedRead || !page.items[0].Read {
		t.Fatalf("read state: calls=%d marked=%v items=%#v", api.readCall, page.markedRead, page.items)
	}

	testkit.SendKeys(program, "G")
	if got := strings.Join(api.calls, ","); got != ",next" {
		t.Fatalf("notification cursors = %q", got)
	}
	if len(page.items) != 2 || page.items[1].ID != "follow-1" {
		t.Fatalf("paginated notifications = %#v", page.items)
	}
}

func TestPageConcealsNotificationPost(t *testing.T) {
	t.Parallel()

	api := &notificationAPI{pages: map[string]notificationdomain.Page{
		"": {Items: []notificationdomain.Item{{
			ID:     "hidden",
			Kind:   notificationdomain.Mention,
			Actors: []north.User{{Name: "Muted", Handle: "muted"}},
			Post: &north.Post{
				Text:         "SHOULD_NOT_RENDER",
				HiddenReason: north.HiddenReason("MUTED"),
			},
		}}},
	}}
	program := reactea.New(NewPage(api, ui.NewTheme()), reactea.WithSize(64, 14))
	program.Start()
	plain := testkit.Plain(program)
	if strings.Contains(plain, "SHOULD_NOT_RENDER") || !strings.Contains(plain, "Hidden: MUTED") {
		t.Fatalf("notification leaked a hidden post:\n%s", plain)
	}
}

func TestPageDoesNotMarkAnUnloadedFeedAsRead(t *testing.T) {
	t.Parallel()

	api := &notificationAPI{loadErr: errors.New("offline")}
	program := reactea.New(NewPage(api, ui.NewTheme()), reactea.WithSize(64, 14))
	program.Start()
	if api.readCall != 0 {
		t.Fatalf("mark-read calls = %d", api.readCall)
	}
	if plain := testkit.Plain(program); !strings.Contains(plain, "offline") || !strings.Contains(plain, "Press . to try again") {
		t.Fatalf("notification load error is incomplete:\n%s", plain)
	}
}

var _ notificationdomain.API = (*notificationAPI)(nil)
