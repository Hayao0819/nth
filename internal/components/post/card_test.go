package post

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderCardFitsWidthAndSanitizes(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	post := north.Post{
		ID:        "1",
		Text:      "long text with \x1b[31mterminal control\x1b[0m, 日本語 and ｶﾞｯﾂﾎﾟｰｽﾞ that wrap",
		CreatedAt: created,
		Author:    north.User{Handle: "user\n1", Name: "User\n1"},
		LikeCount: 3,
	}
	rendered := RenderCard(post, 28, true, ui.NewTheme(), created.Add(time.Hour))
	for index, line := range strings.Split(rendered, "\n") {
		if width := lipgloss.Width(line); width != 28 {
			t.Errorf("line %d width = %d: %q", index, width, line)
		}
		if width := ansi.StringWidthWc(line); width != 28 {
			t.Errorf("line %d terminal width = %d: %q", index, width, line)
		}
	}
	if strings.Contains(rendered, "\x1b[31m") {
		t.Errorf("rendered post retained user styling: %q", rendered)
	}
	plain := ansi.Strip(rendered)
	if !strings.HasPrefix(plain, "> ") || strings.Contains(plain, "▎") {
		t.Errorf("selected post cursor = %q", plain)
	}
	if strings.Contains(plain, "User\n1") {
		t.Errorf("inline author contains a newline: %q", plain)
	}
	if !strings.Contains(plain, "\n    ────") {
		t.Errorf("post card is missing its inset separator: %q", plain)
	}
}

func TestMediaPreviewURLPrefersThumbnail(t *testing.T) {
	t.Parallel()

	thumbnail := "/media/photo.preview.jpg"
	media := north.Media{URL: "/media/photo.jpg", ThumbnailURL: &thumbnail}
	if got := MediaPreviewURL(media); got != thumbnail {
		t.Fatalf("preview URL = %q", got)
	}
	media.ThumbnailURL = nil
	if got := MediaPreviewURL(media); got != media.URL {
		t.Fatalf("fallback URL = %q", got)
	}
}

func TestCardUserAt(t *testing.T) {
	t.Parallel()

	replyTo := "dave"
	post := north.Post{
		Text:            "hello",
		Author:          north.User{Name: "Alice", Handle: "alice"},
		InReplyToHandle: &replyTo,
		Quoted:          &north.Post{Author: north.User{Name: "Carol", Handle: "carol"}},
	}
	if user, ok := CardUserAt(post, 5, 0, 60); !ok || user.Handle != "alice" {
		t.Fatal("visible author was not clickable")
	}
	if _, ok := CardUserAt(post, 4, 0, 60); ok {
		t.Fatal("space outside the author was clickable")
	}
	if user, ok := CardUserAt(post, 17, 1, 60); !ok || user.Handle != "dave" {
		t.Fatal("reply recipient was not clickable")
	}
	if user, ok := CardUserAt(post, 8, 3, 60); !ok || user.Handle != "carol" {
		t.Fatal("quoted author was not clickable")
	}

	repost := north.Post{
		Author:   north.User{Name: "Bob", Handle: "bob"},
		RepostOf: &post,
	}
	if user, ok := CardUserAt(repost, 7, 0, 60); !ok || user.Handle != "bob" {
		t.Fatal("reposting account was not clickable")
	}
	if user, ok := CardUserAt(repost, 5, 1, 60); !ok || user.Handle != "alice" {
		t.Fatal("reposted post author was not clickable")
	}
}

func TestCardQuotedPostAt(t *testing.T) {
	t.Parallel()

	quoted := north.Post{ID: "quoted", Text: "quoted body", Author: north.User{Name: "Carol", Handle: "carol"}}
	post := north.Post{
		Text:   "hello",
		Author: north.User{Name: "Alice", Handle: "alice"},
		Quoted: &quoted,
	}
	if got, ok := CardQuotedPostAt(post, 20, 3, 60); !ok || got.ID != quoted.ID {
		t.Fatalf("quoted post hit = %#v, %v", got, ok)
	}
	if _, ok := CardQuotedPostAt(post, 3, 3, 60); ok {
		t.Fatal("space outside the quoted card was clickable")
	}
	quoted.HiddenReason = north.HiddenReason("MUTED")
	if _, ok := CardQuotedPostAt(post, 20, 3, 60); ok {
		t.Fatal("concealed quoted post was clickable")
	}
}

func TestHiddenQuotedPostDoesNotRevealItsText(t *testing.T) {
	t.Parallel()

	post := north.Post{
		Text:   "visible",
		Author: north.User{Name: "Alice", Handle: "alice"},
		Quoted: &north.Post{
			Text:         "SHOULD_NOT_RENDER",
			HiddenReason: north.HiddenReason("MUTED"),
			Author:       north.User{Name: "Muted", Handle: "muted"},
		},
	}
	plain := ansi.Strip(RenderCard(post, 60, true, ui.NewTheme(), time.Now()))
	if strings.Contains(plain, "SHOULD_NOT_RENDER") || !strings.Contains(plain, "Hidden: MUTED") {
		t.Fatalf("hidden quote leaked through the card: %q", plain)
	}
}

func TestCardRendersAndTargetsBookmark(t *testing.T) {
	t.Parallel()

	post := north.Post{Author: north.User{Name: "Alice", Handle: "alice"}, Bookmarked: true}
	plain := ansi.Strip(RenderCard(post, 60, true, ui.NewTheme(), time.Now()))
	if !strings.Contains(plain, "♣") {
		t.Fatalf("bookmarked card = %q", plain)
	}
	if action := CardActionAt(55, 60); action != Bookmark {
		t.Fatalf("rightmost action = %v", action)
	}
}
