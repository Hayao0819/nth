package navigation

import (
	"testing"

	"github.com/Hayao0819/go-north"
)

func TestNorthWebURLs(t *testing.T) {
	t.Parallel()

	user := north.User{Handle: "@alice"}
	if got, want := UserWebURL(user), "https://north.rip/alice"; got != want {
		t.Fatalf("user URL = %q, want %q", got, want)
	}
	post := north.Post{ID: "123", Author: user}
	if got, want := PostWebURL(post), "https://north.rip/alice/status/123"; got != want {
		t.Fatalf("post URL = %q, want %q", got, want)
	}
}

func TestPostWebURLUsesDisplayedPost(t *testing.T) {
	t.Parallel()

	original := north.Post{ID: "original", Author: north.User{Handle: "author"}}
	repost := north.Post{ID: "repost", Author: north.User{Handle: "reposter"}, RepostOf: &original}
	if got, want := PostWebURL(repost), "https://north.rip/author/status/original"; got != want {
		t.Fatalf("repost URL = %q, want %q", got, want)
	}
	if UserWebURL(north.User{}) != "" || PostWebURL(north.Post{ID: "123"}) != "" {
		t.Fatal("incomplete resources produced a web URL")
	}
}

func TestOpenBrowserMessage(t *testing.T) {
	t.Parallel()

	if OpenBrowser("  ") != nil {
		t.Fatal("empty URL produced a command")
	}
	message, ok := OpenBrowser(" https://north.rip/alice ")().(OpenBrowserMsg)
	if !ok || message.URL != "https://north.rip/alice" {
		t.Fatalf("browser message = %#v", message)
	}
}
