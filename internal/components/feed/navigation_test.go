package feed

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/ui"
)

func TestSelectionScrollsAfterViewportMidpoint(t *testing.T) {
	t.Parallel()

	theme := ui.NewTheme()
	feed := New(&fakeAPI{}, theme)
	for index := range 6 {
		feed.posts = append(feed.posts, testPost(string(rune('a'+index)), "post"))
	}
	width := 60
	cardHeight := lipgloss.Height(postcomponent.RenderCard(feed.posts[0], width, false, theme, time.Now()))
	height := cardHeight * 4

	feed.move(1, width, height)
	feed.move(1, width, height)
	if feed.top != 0 {
		t.Fatalf("top before midpoint = %d", feed.top)
	}

	feed.move(1, width, height)
	if feed.top != 1 {
		t.Fatalf("top after midpoint = %d, want 1", feed.top)
	}

	feed.move(-1, width, height)
	if feed.top != 0 {
		t.Fatalf("top after moving back = %d", feed.top)
	}
}
