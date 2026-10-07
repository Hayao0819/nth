package post

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type activityAPI struct {
	quotes  north.PostPage
	likes   north.UserPage
	history north.EditHistory
}

func (a *activityAPI) Quotes(context.Context, string, string) (north.PostPage, *north.Response, error) {
	return a.quotes, nil, nil
}

func (a *activityAPI) PostLikes(context.Context, string, string) (north.UserPage, *north.Response, error) {
	return a.likes, nil, nil
}

func (a *activityAPI) PostReposts(context.Context, string, string) (north.UserPage, *north.Response, error) {
	return a.likes, nil, nil
}

func (a *activityAPI) PostEditHistory(context.Context, string) (north.EditHistory, *north.Response, error) {
	return a.history, nil, nil
}

func TestQuoteActivityOpensPostAndAuthorSeparately(t *testing.T) {
	t.Parallel()

	quoted := north.Post{ID: "quote-1", Text: "quoted with context", Author: north.User{ID: "alice", Name: "Alice", Handle: "alice"}}
	api := &activityAPI{quotes: north.PostPage{Items: []north.Post{quoted}}}
	screen := NewActivity(api, nil, ui.NewTheme(), north.Post{ID: "source"}, navigation.PostQuotes, nil)
	program := reactea.New(screen, reactea.WithSize(72, 18))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "quoted with context") {
		t.Fatalf("quote activity was not rendered:\n%s", plain)
	}

	command := screen.Update(program.Ctx(), testkit.Key("enter"))
	opened, ok := command().(navigation.OpenPostMsg)
	if !ok || opened.Post.ID != quoted.ID {
		t.Fatalf("enter result = %#v", opened)
	}
	command = screen.Update(program.Ctx(), testkit.Key("u"))
	user, ok := command().(navigation.OpenUserMsg)
	if !ok || user.User.Handle != "alice" {
		t.Fatalf("u result = %#v", user)
	}
}

func TestActivityRendersUsersAndEditHistory(t *testing.T) {
	t.Parallel()

	api := &activityAPI{
		likes: north.UserPage{Items: []north.User{{Name: "Bob", Handle: "bob", FollowerCount: 12}}},
		history: north.EditHistory{Versions: []north.PostVersion{{
			ID: "version-1", Text: "previous text", CreatedAt: time.Date(2026, 10, 7, 9, 30, 0, 0, time.Local),
		}}},
	}
	for _, test := range []struct {
		kind navigation.PostActivity
		want string
	}{
		{kind: navigation.PostLikes, want: "Bob"},
		{kind: navigation.PostHistory, want: "previous text"},
	} {
		screen := NewActivity(api, nil, ui.NewTheme(), north.Post{ID: "source"}, test.kind, nil)
		program := reactea.New(screen, reactea.WithSize(72, 18))
		program.Start()
		if plain := testkit.Plain(program); !strings.Contains(plain, test.want) {
			t.Errorf("activity %d missing %q:\n%s", test.kind, test.want, plain)
		}
	}
}
