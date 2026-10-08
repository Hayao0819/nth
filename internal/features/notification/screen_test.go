package notification

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	notificationdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type notificationAPI struct {
	pages    map[string]notificationdomain.NotificationPage
	calls    []string
	tabs     []north.NotificationTab
	readCall int
	readErr  error
	loadErr  error
}

func (a *notificationAPI) Notifications(_ context.Context, tab north.NotificationTab, cursor string) (notificationdomain.NotificationPage, *north.Response, error) {
	a.calls = append(a.calls, cursor)
	a.tabs = append(a.tabs, tab)

	return a.pages[cursor], nil, a.loadErr
}

func TestNotificationTabsReloadTheSelectedFeed(t *testing.T) {
	t.Parallel()

	api := &notificationAPI{pages: map[string]notificationdomain.NotificationPage{"": {}}}
	page := NewPage(api, ui.NewTheme())
	program := reactea.New(page, reactea.WithSize(70, 18))
	program.Start()
	testkit.SendKeys(program, "tab", "tab", "shift+tab")

	want := []north.NotificationTab{
		north.NotificationsAll,
		north.NotificationsVerified,
		north.NotificationsMentions,
		north.NotificationsVerified,
	}
	if len(api.tabs) != len(want) {
		t.Fatalf("tabs = %#v", api.tabs)
	}
	for index := range want {
		if api.tabs[index] != want[index] {
			t.Fatalf("tabs = %#v", api.tabs)
		}
	}
	if plain := testkit.Plain(program); !strings.Contains(plain, "Verified") {
		t.Fatalf("selected tab is not visible:\n%s", plain)
	}
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
	api := &notificationAPI{pages: map[string]notificationdomain.NotificationPage{
		"": {
			Items: []notificationdomain.NotificationItem{{
				ID:         "like-1",
				Kind:       notificationdomain.NotificationLike,
				Actors:     []north.User{{Name: "Alice", Handle: "alice"}},
				ActorCount: 2,
				CreatedAt:  time.Now().Add(-time.Minute),
				Post:       &north.Post{ID: "post-1", Text: "hello"},
			}},
			NextCursor: &next,
		},
		"next": {
			Items: []notificationdomain.NotificationItem{{
				ID: "follow-1", Kind: notificationdomain.NotificationFollow,
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

func TestInitialLoadFetchesEnoughNotificationsForTheViewport(t *testing.T) {
	t.Parallel()

	next := "next"
	first := notificationdomain.NotificationItem{
		ID: "first", Kind: notificationdomain.NotificationFollow,
		Actors: []north.User{{Name: "First", Handle: "first"}},
	}
	more := make([]notificationdomain.NotificationItem, 6)
	for index := range more {
		more[index] = notificationdomain.NotificationItem{
			ID: string(rune('a' + index)), Kind: notificationdomain.NotificationFollow,
			Actors: []north.User{{Name: "More", Handle: "more"}},
		}
	}
	api := &notificationAPI{pages: map[string]notificationdomain.NotificationPage{
		"":     {Items: []notificationdomain.NotificationItem{first}, NextCursor: &next},
		"next": {Items: more},
	}}
	page := NewPage(api, ui.NewTheme())
	program := reactea.New(page, reactea.WithSize(70, 18))
	program.Start()

	if got := strings.Join(api.calls, ","); got != ",next" {
		t.Fatalf("notification cursors = %q", got)
	}
	if len(page.items) != 7 {
		t.Fatalf("loaded notifications = %d, want 7", len(page.items))
	}
}

func TestPartiallyVisibleNotificationFillsAndUsesTheLastRow(t *testing.T) {
	t.Parallel()

	screen := NewPage(nil, ui.NewTheme())
	screen.items = []notificationdomain.NotificationItem{
		{ID: "first", Kind: notificationdomain.NotificationFollow, Actors: []north.User{{Name: "Alice", Handle: "alice"}}},
		{ID: "second", Kind: notificationdomain.NotificationFollow, Actors: []north.User{{Name: "Bob", Handle: "bob"}}},
	}
	program := reactea.New(screen, reactea.WithSize(70, 8))
	program.Start()
	plain := testkit.Plain(program)
	if !strings.Contains(plain, "Bob") {
		t.Fatalf("partially visible notification left the final body row blank:\n%s", plain)
	}

	bob := testkit.Find(program, "Bob")
	if len(bob) == 0 {
		t.Fatal("partially visible notification has no hit target")
	}
	command := screen.Update(program.Ctx(), tea.MouseClickMsg{X: bob[0].X, Y: bob[0].Y, Button: tea.MouseLeft})
	opened, ok := command().(navigation.OpenUserMsg)
	if !ok || opened.User.Handle != "bob" {
		t.Fatalf("open partially visible notification = %#v", opened)
	}
}

func TestPageConcealsNotificationPost(t *testing.T) {
	t.Parallel()

	api := &notificationAPI{pages: map[string]notificationdomain.NotificationPage{
		"": {Items: []notificationdomain.NotificationItem{{
			ID:     "hidden",
			Kind:   notificationdomain.NotificationMention,
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

func TestClickingActorOpensProfileAndCardOpensPost(t *testing.T) {
	t.Parallel()

	avatar := "/media/alice.png"
	api := &notificationAPI{pages: map[string]notificationdomain.NotificationPage{
		"": {Items: []notificationdomain.NotificationItem{{
			ID:     "reply-1",
			Kind:   notificationdomain.NotificationReply,
			Actors: []north.User{{Name: "Alice", Handle: "alice", AvatarURL: &avatar}},
			Post:   &north.Post{ID: "post-1", Text: "hello"},
		}}},
	}}
	screen := NewPage(api, ui.NewTheme())
	program := reactea.New(screen, reactea.WithSize(70, 18))
	program.Start()

	actor := testkit.Find(program, "Alice")
	if len(actor) == 0 {
		t.Fatalf("actor was not rendered:\n%s", testkit.Plain(program))
	}
	command := screen.Update(program.Ctx(), tea.MouseClickMsg{X: actor[0].X, Y: actor[0].Y, Button: tea.MouseLeft})
	if command == nil {
		t.Fatal("clicking an actor returned no command")
	}
	userMessage, ok := command().(navigation.OpenUserMsg)
	if !ok || userMessage.User.Handle != "alice" {
		t.Fatalf("open actor = %#v", userMessage)
	}

	postText := testkit.Find(program, "hello")
	if len(postText) == 0 {
		t.Fatal("notification post was not rendered")
	}
	command = screen.Update(program.Ctx(), tea.MouseClickMsg{X: postText[0].X, Y: postText[0].Y, Button: tea.MouseLeft})
	if command == nil {
		t.Fatal("clicking a notification post returned no command")
	}
	postMessage, ok := command().(navigation.OpenPostMsg)
	if !ok || postMessage.Post.ID != "post-1" {
		t.Fatalf("open post = %#v", postMessage)
	}
}

var _ notificationdomain.NotificationAPI = (*notificationAPI)(nil)
