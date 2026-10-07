package diagnostic

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestChecksShowValuesReturnedByNorth(t *testing.T) {
	t.Parallel()

	api := &diagnosticAPI{}
	want := []string{
		"Alice  @alice",
		"@bob: hello from timeline",
		"like · @carol",
		"@dave: saved post",
		"North team: hello privately",
		"#golang · 42 posts",
	}
	for index, check := range checks(api) {
		summary, err := check.run(context.Background())
		if err != nil {
			t.Fatalf("%s: %v", check.name, err)
		}
		if !strings.Contains(summary, want[index]) {
			t.Errorf("%s summary = %q", check.name, summary)
		}
	}
	for index := range api.calls {
		if calls := api.calls[index].Load(); calls != 1 {
			t.Errorf("check %d calls = %d", index, calls)
		}
	}
}

func TestScreenRendersLiveResults(t *testing.T) {
	t.Parallel()

	program := reactea.New(New(&diagnosticAPI{}), reactea.WithSize(88, 22))
	program.Start()
	view := waitForText(t, program, "All 6 checks passed")
	for _, text := range []string{"Live API check", "Alice", "hello from timeline", "hello privately", "#golang", "q close"} {
		if !strings.Contains(view, text) {
			t.Errorf("view is missing %q:\n%s", text, view)
		}
	}
}

func TestScreenKeepsFailedChecksVisible(t *testing.T) {
	t.Parallel()

	program := reactea.New(New(&failingDiagnosticAPI{diagnosticAPI: &diagnosticAPI{}}), reactea.WithSize(88, 22))
	program.Start()
	view := waitForText(t, program, "5 passed · 1 failed")
	if !strings.Contains(view, "permission denied") || !strings.Contains(view, "× Notifications") {
		t.Fatalf("failed result is missing:\n%s", view)
	}
}

func TestScreenExplainsSmallTerminal(t *testing.T) {
	t.Parallel()

	program := reactea.New(New(&diagnosticAPI{}), reactea.WithSize(40, 12))
	program.Start()
	if view := testkit.Plain(program); !strings.Contains(view, "resize to at least 48×18") {
		t.Fatalf("small-terminal message is missing:\n%s", view)
	}
}

func waitForText(t *testing.T, program *reactea.App, text string) string {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		view := testkit.Plain(program)
		if strings.Contains(view, text) {
			return view
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %q:\n%s", text, view)
		}
		time.Sleep(time.Millisecond)
	}
}

type diagnosticAPI struct {
	calls [6]atomic.Int32
}

func (a *diagnosticAPI) Me(context.Context) (north.User, *north.Response, error) {
	a.calls[0].Add(1)

	return north.User{Name: "Alice", Handle: "alice"}, nil, nil
}

func (a *diagnosticAPI) HomeTimeline(context.Context, north.TimelineOptions) (north.PostPage, *north.Response, error) {
	a.calls[1].Add(1)

	return north.PostPage{Items: []north.Post{{Text: "hello from timeline", Author: north.User{Handle: "bob"}}}}, nil, nil
}

func (a *diagnosticAPI) Notifications(context.Context, north.NotificationTab, string) (north.NotificationPage, *north.Response, error) {
	a.calls[2].Add(1)

	return north.NotificationPage{Items: []north.Notification{{Kind: north.NotificationLike, Actors: []north.User{{Handle: "carol"}}}}}, nil, nil
}

func (a *diagnosticAPI) Bookmarks(context.Context, string) (north.PostPage, *north.Response, error) {
	a.calls[3].Add(1)

	return north.PostPage{Items: []north.Post{{Text: "saved post", Author: north.User{Handle: "dave"}}}}, nil, nil
}

func (a *diagnosticAPI) DMConversations(context.Context, string, bool) (north.DMConversationPage, *north.Response, error) {
	a.calls[4].Add(1)
	name := "North team"

	return north.DMConversationPage{Items: []north.DMConversation{{Name: &name, LastMessage: &north.DMMessage{Text: "hello privately"}}}}, nil, nil
}

func (a *diagnosticAPI) Trends(context.Context, string) ([]north.Trend, *north.Response, error) {
	a.calls[5].Add(1)

	return []north.Trend{{Tag: "golang", Count: 42, IsHashtag: true}}, nil, nil
}

type failingDiagnosticAPI struct {
	*diagnosticAPI
}

func (a *failingDiagnosticAPI) Notifications(context.Context, north.NotificationTab, string) (north.NotificationPage, *north.Response, error) {
	a.calls[2].Add(1)

	return north.NotificationPage{}, nil, errors.New("permission denied")
}
