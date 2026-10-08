package user

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
	"github.com/charmbracelet/x/ansi"
)

type profileAPI struct {
	user      north.User
	page      north.PostPage
	pages     map[string]north.PostPage
	err       error
	postErr   error
	calls     []string
	postCalls []string
}

func (a *profileAPI) User(_ context.Context, handle string) (north.User, *north.Response, error) {
	a.calls = append(a.calls, handle)

	return a.user, nil, a.err
}

func (a *profileAPI) UserPosts(_ context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	a.postCalls = append(a.postCalls, handle+":"+cursor)
	if a.pages != nil {
		return a.pages[cursor], nil, a.postErr
	}

	return a.page, nil, a.postErr
}

func TestPageLoadsAndRendersProfile(t *testing.T) {
	t.Parallel()

	bio := "Building a quieter place on the internet."
	location := "Tokyo"
	website := "https://example.com"
	created := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	api := &profileAPI{user: north.User{
		ID:             "user-1",
		Handle:         "alice",
		Name:           "Alice",
		Bio:            &bio,
		Location:       &location,
		Website:        &website,
		Verified:       true,
		FollowerCount:  12,
		FollowingCount: 34,
		PostCount:      56,
		CreatedAt:      &created,
		Following:      true,
		FollowedBy:     true,
	}, page: north.PostPage{Items: []north.Post{{
		ID:        "post-1",
		Text:      "A post from Alice",
		CreatedAt: time.Now(),
		Author:    north.User{ID: "user-1", Handle: "alice", Name: "Alice"},
	}}}}
	program := reactea.New(
		NewPage(api, ui.NewTheme(), north.User{Handle: "alice", Name: "Alice"}),
		reactea.WithSize(64, 28),
	)
	program.Start()

	plain := testkit.Plain(program)
	for _, want := range []string{
		"Profile", "Alice ✓", "@alice", bio, "56", "34", "12", "Posts", "Following",
		"Followers", "⌖ Tokyo", "↗ https://example.com", "Joined April 2025",
		"✓ Following · Follows you", "A post from Alice", "1 of 1",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("profile missing %q:\n%s", want, plain)
		}
	}
	countsRow := lineContaining(t, plain, "56")
	labelsRow := lineContaining(t, plain, "Posts")
	if labelsRow != countsRow+1 {
		t.Fatalf("profile counts and labels are not grouped:\n%s", plain)
	}
	if profileAt, postAt := strings.Index(plain, "Joined April 2025"), strings.Index(plain, "A post from Alice"); profileAt < 0 || postAt <= profileAt {
		t.Fatalf("posts are not below the profile details:\n%s", plain)
	}
	if len(api.calls) != 1 || api.calls[0] != "alice" {
		t.Fatalf("profile calls = %#v", api.calls)
	}
	if len(api.postCalls) != 1 || api.postCalls[0] != "alice:" {
		t.Fatalf("post calls = %#v", api.postCalls)
	}
	if width, height := lipgloss.Size(program.View().Content); width != 64 || height != 28 {
		t.Fatalf("profile size = %dx%d", width, height)
	}
}

func TestPageOpensProfileInBrowser(t *testing.T) {
	t.Parallel()

	screen := NewPage(nil, ui.NewTheme(), north.User{Handle: "alice", Name: "Alice"})
	program := reactea.New(screen, reactea.WithSize(64, 16))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, navigation.BrowserActionLabel) {
		t.Fatalf("browser action is missing:\n%s", plain)
	}

	for name, message := range map[string]tea.Msg{
		"keyboard": testkit.Key("w"),
		"mouse":    tea.MouseClickMsg{X: 63, Y: 1, Button: tea.MouseLeft},
	} {
		command := screen.Update(program.Ctx(), message)
		opened, ok := command().(navigation.OpenBrowserMsg)
		if !ok || opened.URL != "https://north.rip/alice" {
			t.Errorf("%s browser request = %#v", name, opened)
		}
	}
}

func TestPageKeepsInitialUserWhenLoadingFails(t *testing.T) {
	t.Parallel()

	api := &profileAPI{err: errors.New("offline"), postErr: errors.New("offline")}
	program := reactea.New(
		NewPage(api, ui.NewTheme(), north.User{Handle: "alice", Name: "Alice"}),
		reactea.WithSize(56, 14),
	)
	program.Start()

	plain := testkit.Plain(program)
	if !strings.Contains(plain, "Alice") || !strings.Contains(plain, "@alice") || !strings.Contains(plain, "Profile details unavailable: offline") {
		t.Fatalf("profile fallback:\n%s", plain)
	}
	if strings.Contains(plain, "Followers") {
		t.Fatalf("unknown profile counts were rendered as zero:\n%s", plain)
	}
	if !strings.Contains(plain, ". retry") {
		t.Fatalf("profile retry hint is missing:\n%s", plain)
	}

	bio := "Loaded after retry"
	api.err = nil
	api.postErr = nil
	api.user = north.User{Handle: "alice", Name: "Alice", Bio: &bio, PostCount: 7}
	api.page = north.PostPage{Items: []north.Post{{ID: "post-1", Text: "Loaded post", Author: north.User{Handle: "alice", Name: "Alice"}}}}
	testkit.SendKeys(program, ".")
	plain = testkit.Plain(program)
	if len(api.calls) != 2 || len(api.postCalls) != 2 || !strings.Contains(plain, bio) || !strings.Contains(plain, "7") || !strings.Contains(plain, "Posts") {
		t.Fatalf("profile retry failed, calls=%#v:\n%s", api.calls, plain)
	}
}

func TestPageLoadsOlderPostsToFillTheScreen(t *testing.T) {
	t.Parallel()

	next := "next"
	api := &profileAPI{
		user: north.User{ID: "user-1", Handle: "alice", Name: "Alice", PostCount: 2},
		pages: map[string]north.PostPage{
			"": {
				Items:      []north.Post{{ID: "post-1", Text: "newer post", Author: north.User{Handle: "alice", Name: "Alice"}}},
				NextCursor: &next,
			},
			"next": {
				Items: []north.Post{{ID: "post-2", Text: "older post", Author: north.User{Handle: "alice", Name: "Alice"}}},
			},
		},
	}
	program := reactea.New(
		NewPage(api, ui.NewTheme(), north.User{Handle: "alice", Name: "Alice"}),
		reactea.WithSize(64, 30),
	)
	program.Start()

	plain := testkit.Plain(program)
	if !strings.Contains(plain, "newer post") || !strings.Contains(plain, "older post") {
		t.Fatalf("profile did not fill with both pages:\n%s", plain)
	}
	if got := strings.Join(api.postCalls, ","); got != "alice:,alice:next" {
		t.Fatalf("post calls = %q", got)
	}
}

func TestProfileLinesFitNarrowWidths(t *testing.T) {
	t.Parallel()

	bio := "ｶﾞｯﾂﾎﾟｰｽﾞをしながら長い自己紹介を書いています"
	location := "とても長い場所の名前"
	website := "https://example.com/a/very/long/profile/path"
	user := north.User{
		Handle: "alice", Name: "ｶﾞｯﾂﾎﾟｰｽﾞ Alice", Bio: &bio,
		Location: &location, Website: &website,
		Following: true, FollowedBy: true, Muting: true,
	}
	for _, width := range []int{12, 24, 48} {
		for row, line := range profileLines(ui.NewTheme(), user, width) {
			if got := lipgloss.Width(line); got > width {
				t.Errorf("width %d row %d has lipgloss width %d: %q", width, row, got, line)
			}
			if got := ansi.StringWidthWc(line); got > width {
				t.Errorf("width %d row %d has terminal width %d: %q", width, row, got, line)
			}
		}
	}
}

func lineContaining(t *testing.T, value, fragment string) int {
	t.Helper()
	for index, line := range strings.Split(value, "\n") {
		if strings.Contains(line, fragment) {
			return index
		}
	}
	t.Fatalf("line containing %q not found:\n%s", fragment, value)

	return -1
}

var _ API = (*profileAPI)(nil)
