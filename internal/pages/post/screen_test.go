package postpage

import (
	"context"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	conversationdomain "github.com/Hayao0819/nth/internal/domain/conversation"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestScreenRendersPostDetails(t *testing.T) {
	t.Parallel()

	replyTo := "bob"
	alt := "A mountain at sunset"
	voted := 1
	created := time.Date(2026, 9, 29, 12, 34, 0, 0, time.Local)
	post := north.Post{
		ID:              "post-1",
		Text:            "A detailed post",
		CreatedAt:       created,
		Source:          "north web",
		InReplyToHandle: &replyTo,
		Author:          north.User{Name: "Alice", Handle: "alice", Verified: true},
		ReplyCount:      1,
		RepostCount:     2,
		QuoteCount:      3,
		LikeCount:       4,
		Media: []north.Media{{
			Kind: north.MediaPhoto, URL: "https://cdn.example/image.jpg", Width: 1200, Height: 800, AltText: &alt,
		}},
		Quoted: &north.Post{Text: "quoted text", Author: north.User{Name: "Bob", Handle: "bob"}},
		Poll:   &north.Poll{Voted: &voted, Options: []north.PollOption{{Label: "One", Votes: 1}, {Label: "Two", Votes: 3}}},
	}
	program := reactea.New(NewPage(nil, ui.NewTheme(), post, false), reactea.WithSize(78, 26))
	program.Start()
	plain := testkit.Plain(program)
	for _, want := range []string{
		"Alice ✓", "@alice", "Replying to @bob", "A detailed post", "Photo · 1200×800",
		"A mountain at sunset", "cdn.example", "Bob @bob", "quoted text", "Two", "75%",
		"12:34 · 2026-09-29 · north web", "r  Reply", "Q  Quote",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("detail missing %q:\n%s", want, plain)
		}
	}
	testkit.SendKeys(program, "G")
	plain = testkit.Plain(program)
	for _, want := range []string{"1 Replies", "2 Reposts", "3 Quotes", "4 Likes"} {
		if !strings.Contains(plain, want) {
			t.Errorf("scrolled detail missing %q:\n%s", want, plain)
		}
	}
	if width, height := lipgloss.Size(program.View().Content); width != 78 || height != 26 {
		t.Errorf("screen size = %dx%d", width, height)
	}
}

func TestScreenScrollsMetadataAndActionsWithThePost(t *testing.T) {
	t.Parallel()

	post := north.Post{
		ID:        "long-post",
		Text:      strings.Repeat("A long post needs scrolling. ", 20),
		CreatedAt: time.Date(2026, 9, 29, 12, 34, 0, 0, time.Local),
		Source:    "north web",
		Author:    north.User{Name: "Alice", Handle: "alice"},
	}
	program := reactea.New(NewPage(nil, ui.NewTheme(), post, false), reactea.WithSize(60, 14))
	program.Start()

	if initial := testkit.Plain(program); strings.Contains(initial, "12:34 · 2026-09-29 · north web") || strings.Contains(initial, "r  Reply") {
		t.Fatalf("metadata or actions are pinned over the post body:\n%s", initial)
	}
	testkit.SendKeys(program, "G")
	after := testkit.Plain(program)
	metadataRow := lineContaining(t, after, "12:34 · 2026-09-29 · north web")
	if actionRow := lineContaining(t, after, "r  Reply"); actionRow <= metadataRow {
		t.Fatalf("action row %d is not below metadata row %d:\n%s", actionRow, metadataRow, after)
	}
}

func TestHiddenPostCannotBeActedOn(t *testing.T) {
	t.Parallel()

	post := north.Post{ID: "hidden", HiddenReason: north.HiddenReason("SENSITIVE")}
	if postcomponent.CanInteract(&post) {
		t.Fatal("hidden post is interactive")
	}
	program := reactea.New(NewPage(nil, ui.NewTheme(), post, false), reactea.WithSize(60, 16))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "Actions unavailable") {
		t.Fatalf("hidden post footer:\n%s", plain)
	}
}

func TestScreenConcealsNestedQuote(t *testing.T) {
	t.Parallel()

	post := north.Post{
		Text:   "visible",
		Author: north.User{Name: "Alice", Handle: "alice"},
		Quoted: &north.Post{
			Text:        "SHOULD_NOT_RENDER",
			Unavailable: true,
			Author:      north.User{Name: "Unavailable", Handle: "gone"},
		},
	}
	program := reactea.New(NewPage(nil, ui.NewTheme(), post, false), reactea.WithSize(60, 18))
	program.Start()
	plain := testkit.Plain(program)
	if strings.Contains(plain, "SHOULD_NOT_RENDER") || !strings.Contains(plain, "Post unavailable") {
		t.Fatalf("unavailable quote leaked through the screen:\n%s", plain)
	}
}

func TestScreenTracksQuotedPostHitArea(t *testing.T) {
	t.Parallel()

	quoted := north.Post{ID: "quoted", Text: "quoted body", Author: north.User{Name: "Bob", Handle: "bob"}}
	screen := NewPage(nil, ui.NewTheme(), north.Post{
		ID: "outer", Text: "outer body", Author: north.User{Name: "Alice", Handle: "alice"}, Quoted: &quoted,
	}, false)
	_, users, posts := screen.content(70)
	if len(posts) != 1 {
		t.Fatalf("quoted post hits = %#v", posts)
	}
	hit := posts[0]
	if got, ok := postAt(posts, 20, hit.top+1); !ok || got.ID != "quoted" {
		t.Fatalf("quoted post at body = %#v, %v", got, ok)
	}
	if user, ok := userAt(users, 4, hit.top); !ok || user.Handle != "bob" {
		t.Fatalf("quoted author at title = %#v, %v", user, ok)
	}
}

func TestScreenKeepsAllMetadataAtNarrowWidths(t *testing.T) {
	t.Parallel()

	post := north.Post{
		Text:        "narrow",
		CreatedAt:   time.Date(2026, 9, 29, 12, 34, 0, 0, time.Local),
		Source:      "north web client",
		ReplyCount:  1,
		RepostCount: 2,
		QuoteCount:  3,
		LikeCount:   4,
	}
	program := reactea.New(NewPage(nil, ui.NewTheme(), post, false), reactea.WithSize(30, 16))
	program.Start()
	plain := testkit.Plain(program)
	for _, want := range []string{"north", "web", "client", "1 Replies", "2 Reposts", "3 Quotes", "4 Likes"} {
		if !strings.Contains(plain, want) {
			t.Errorf("narrow detail missing %q:\n%s", want, plain)
		}
	}
}

func TestScreenShowsEditOnlyWhenThePostIsEligible(t *testing.T) {
	t.Parallel()

	post := north.Post{ID: "post-1", Text: "editable", Author: north.User{ID: "me", Handle: "alice"}}
	editor := &screenEditor{post: post, eligible: true}
	program := reactea.New(
		NewPageWithEditor(nil, editor, ui.NewTheme(), post, true),
		reactea.WithSize(72, 16),
	)
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "e  Edit") || !strings.Contains(plain, "d  Delete") {
		t.Fatalf("management actions are incomplete:\n%s", plain)
	}
}

func TestScreenEnablesManagementWhenAccountArrivesLater(t *testing.T) {
	t.Parallel()

	post := north.Post{ID: "post-1", Text: "editable", Author: north.User{ID: "me", Handle: "alice"}}
	editor := &screenEditor{post: post, eligible: true}
	program := reactea.New(
		NewPageWithEditor(nil, editor, ui.NewTheme(), post, false),
		reactea.WithSize(72, 16),
	)
	program.Start()
	if plain := testkit.Plain(program); strings.Contains(plain, "d  Delete") {
		t.Fatalf("management appeared before the account loaded:\n%s", plain)
	}
	program.Send(ManagementUpdate{Account: north.User{ID: "me", Handle: "alice"}})
	if plain := testkit.Plain(program); !strings.Contains(plain, "e  Edit") || !strings.Contains(plain, "d  Delete") {
		t.Fatalf("management did not appear after the account loaded:\n%s", plain)
	}
}

func TestLinkedUsersCoverEveryInteractiveAccount(t *testing.T) {
	t.Parallel()

	reply := "reply"
	post := north.Post{
		Author: north.User{Handle: "reposter"},
		RepostOf: &north.Post{
			Author:          north.User{Handle: "original"},
			InReplyToHandle: &reply,
			Quoted:          &north.Post{Author: north.User{Handle: "quoted"}},
		},
	}
	users := postcomponent.LinkedUsers(post, north.Post{Author: north.User{Handle: "ancestor"}})
	var handles []string
	for _, user := range users {
		handles = append(handles, user.Handle)
	}
	if got := strings.Join(handles, ","); got != "original,reposter,reply,quoted,ancestor" {
		t.Fatalf("linked users = %q", got)
	}
}

func TestScreenLoadsReplyAncestors(t *testing.T) {
	t.Parallel()

	rootID := "root"
	parentID := "parent"
	api := &screenPostAPI{posts: map[string]north.Post{
		rootID: {
			ID: rootID, Text: "first post", Author: north.User{Name: "Root", Handle: "root"},
		},
		parentID: {
			ID: parentID, Text: "second post", InReplyToID: &rootID,
			Author: north.User{Name: "Parent", Handle: "parent"},
		},
	}}
	post := north.Post{
		ID: "current", Text: "current post", InReplyToID: &parentID,
		Author: north.User{Name: "Current", Handle: "current"},
	}
	program := reactea.New(NewPage(api, ui.NewTheme(), post, false), reactea.WithSize(78, 32))
	program.Start()
	plain := testkit.Plain(program)

	rootAt := strings.Index(plain, "first post")
	parentAt := strings.Index(plain, "second post")
	currentAt := strings.Index(plain, "current post")
	if rootAt < 0 || parentAt <= rootAt || currentAt <= parentAt {
		t.Fatalf("ancestor order is wrong:\n%s", plain)
	}
	if !strings.Contains(plain, "Conversation") || !strings.Contains(plain, "│ Root @root") {
		t.Fatalf("conversation markers missing:\n%s", plain)
	}
	if strings.Join(api.calls, ",") != "parent,root" {
		t.Fatalf("post calls = %#v", api.calls)
	}
}

func TestScreenLoadsRepliesBelowThePost(t *testing.T) {
	t.Parallel()

	next := "next"
	api := &screenConversationAPI{pages: map[string]conversationdomain.Page{
		"": {
			Ancestors:  []north.Post{{ID: "parent", Text: "parent body", Author: north.User{Name: "Parent", Handle: "parent"}}},
			Replies:    []north.Post{{ID: "reply-1", Text: "first reply", Author: north.User{Name: "Bob", Handle: "bob"}}},
			NextCursor: &next,
		},
		"next": {
			Replies: []north.Post{{ID: "reply-2", Text: "second reply", Author: north.User{Name: "Carol", Handle: "carol"}}},
		},
	}}
	screen := NewPage(api, ui.NewTheme(), north.Post{
		ID: "post-1", Text: "original body", Author: north.User{Name: "Alice", Handle: "alice"},
	}, false)
	program := reactea.New(screen, reactea.WithSize(78, 50))
	program.Start()
	plain := testkit.Plain(program)

	parentAt := strings.Index(plain, "parent body")
	originalAt := strings.Index(plain, "original body")
	firstAt := strings.Index(plain, "first reply")
	secondAt := strings.Index(plain, "second reply")
	if parentAt < 0 || originalAt <= parentAt || firstAt <= originalAt || secondAt <= firstAt {
		t.Fatalf("conversation order is wrong:\n%s", plain)
	}
	if got := strings.Join(api.calls, ","); got != ",next" {
		t.Fatalf("conversation cursors = %q", got)
	}

	testkit.SendKeys(program, "J")
	if !screen.replyFocused || screen.replyIndex != 0 {
		t.Fatalf("selected reply = focused %v index %d", screen.replyFocused, screen.replyIndex)
	}
	command := screen.Update(program.Ctx(), testkit.Key("enter"))
	request, ok := command().(postcomponent.ActionMsg)
	if !ok || request.Action != postcomponent.ViewPost || request.Post.ID != "reply-1" {
		t.Fatalf("open reply request = %#v", request)
	}
}

func TestScreenTracksVisibleAccounts(t *testing.T) {
	t.Parallel()

	replyTo := "reply"
	screen := NewPage(nil, ui.NewTheme(), north.Post{
		Author:          north.User{Name: "Current", Handle: "current"},
		InReplyToHandle: &replyTo,
		Quoted:          &north.Post{Author: north.User{Name: "Quoted", Handle: "quoted"}},
	}, false)
	screen.ancestors = []north.Post{{Author: north.User{Name: "Ancestor", Handle: "ancestor"}}}

	_, hits, _ := screen.content(70)
	found := make(map[string]bool)
	for _, hit := range hits {
		found[hit.user.Handle] = true
	}
	for _, handle := range []string{"ancestor", "current", "reply", "quoted"} {
		if !found[handle] {
			t.Errorf("@%s is visible but not clickable", handle)
		}
	}

	repost := NewPage(nil, ui.NewTheme(), north.Post{
		Author:   north.User{Name: "Reposter", Handle: "reposter"},
		RepostOf: &north.Post{Author: north.User{Name: "Original", Handle: "original"}},
	}, false)
	_, hits, _ = repost.content(70)
	found = make(map[string]bool)
	for _, hit := range hits {
		found[hit.user.Handle] = true
	}
	for _, handle := range []string{"reposter", "original"} {
		if !found[handle] {
			t.Errorf("@%s is visible but not clickable", handle)
		}
	}
}

func TestScreenTracksReplySourceHitAreas(t *testing.T) {
	t.Parallel()

	parent := north.Post{ID: "parent", Text: "parent post", Author: north.User{Name: "Parent", Handle: "parent"}}
	screen := NewPage(nil, ui.NewTheme(), north.Post{ID: "reply"}, false)
	screen.ancestors = []north.Post{parent}
	lines, users, posts := screen.content(70)
	if len(posts) != 1 || posts[0].post.ID != parent.ID {
		t.Fatalf("reply source hits = %#v", posts)
	}
	row := lineContaining(t, strings.Join(lines, "\n"), "parent post")
	if got, ok := postAt(posts, 10, row); !ok || got.ID != parent.ID {
		t.Fatalf("reply source at body = %#v, %v", got, ok)
	}
	if user, ok := userAt(users, 3, posts[0].top); !ok || user.Handle != "parent" {
		t.Fatalf("reply source author = %#v, %v", user, ok)
	}
}

func TestScreenHitTestingUsesRenderedScrollOffset(t *testing.T) {
	t.Parallel()

	root := north.Post{ID: "root", Text: "root post", Author: north.User{Name: "Root", Handle: "root"}}
	parent := north.Post{ID: "parent", Text: "parent post", Author: north.User{Name: "Parent", Handle: "parent"}}
	screen := NewPage(nil, ui.NewTheme(), north.Post{ID: "reply", Text: "reply"}, false)
	screen.ancestors = []north.Post{root, parent}
	screen.offset = 1_000
	lines, _, posts := screen.content(70)
	if len(posts) != 2 {
		t.Fatalf("reply source hits = %#v", posts)
	}
	room := len(lines) - posts[1].top
	if got, ok := screen.postAtPosition(10, pageheader.Height, 70, room); !ok || got.ID != parent.ID {
		t.Fatalf("post under the rendered offset = %#v, %v", got, ok)
	}
}

type screenPostAPI struct {
	posts map[string]north.Post
	calls []string
}

type screenConversationAPI struct {
	pages map[string]conversationdomain.Page
	calls []string
}

func (a *screenConversationAPI) Post(context.Context, string) (north.Post, *north.Response, error) {
	return north.Post{}, nil, nil
}

func (a *screenConversationAPI) PostConversation(_ context.Context, _ string, cursor string) (conversationdomain.Page, *north.Response, error) {
	a.calls = append(a.calls, cursor)

	return a.pages[cursor], nil, nil
}

type screenEditor struct {
	post     north.Post
	eligible bool
}

func (e *screenEditor) EditablePost(context.Context, string) (north.Post, bool, *north.Response, error) {
	return e.post, e.eligible, nil, nil
}

func (e *screenEditor) EditPost(context.Context, string, string, []string) (*north.Response, error) {
	return nil, nil
}

func (a *screenPostAPI) Post(_ context.Context, id string) (north.Post, *north.Response, error) {
	a.calls = append(a.calls, id)

	return a.posts[id], nil, nil
}

func lineContaining(t *testing.T, value, fragment string) int {
	t.Helper()
	for row, line := range strings.Split(value, "\n") {
		if strings.Contains(line, fragment) {
			return row
		}
	}
	t.Fatalf("line containing %q not found:\n%s", fragment, value)

	return -1
}
