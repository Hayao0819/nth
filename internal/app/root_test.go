package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/domain/notification"
	searchpage "github.com/Hayao0819/nth/internal/pages/search"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/Hayao0819/reactea/v2/testkit"
)

func TestRootNavigationSearchAndCompose(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	if root.me == nil || root.me.Handle != "alice" {
		t.Fatalf("me = %#v", root.me)
	}

	testkit.SendKeys(program, "tab")
	if root.feed.CurrentMode() != feed.Ranked || len(service.homeCalls) < 2 || !service.homeCalls[len(service.homeCalls)-1].Ranked {
		t.Fatalf("tab mode = %v, calls = %#v", root.feed.CurrentMode(), service.homeCalls)
	}
	testkit.SendKeys(program, "shift+tab")
	if root.feed.CurrentMode() != feed.Home {
		t.Fatalf("shift+tab mode = %v", root.feed.CurrentMode())
	}

	testkit.SendKeys(program, "/", "g", "o")
	search, ok := root.page.component.(*searchpage.Screen)
	if root.page.kind != searchPage || !ok || search.Query() != "go" {
		t.Fatalf("search page = %v component = %T", root.page.kind, root.page.component)
	}
	if len(service.searchCalls) == 0 || service.searchCalls[len(service.searchCalls)-1] != "go" {
		t.Fatalf("search calls = %#v", service.searchCalls)
	}
	testkit.SendKeys(program, "esc")
	if root.page.kind != timelinePage || root.feed.CurrentMode() != feed.Home {
		t.Fatalf("esc from search page = %v, mode = %v", root.page.kind, root.feed.CurrentMode())
	}

	testkit.SendKeys(program, "n", "h", "e", "l", "l", "o", "ctrl+enter")
	if len(service.created) != 1 || service.created[0].Text != "hello" {
		t.Fatalf("created = %#v", service.created)
	}

	frame := program.View().Content
	width, height := lipgloss.Size(frame)
	if width != 80 || height != 24 {
		t.Fatalf("frame size = %dx%d", width, height)
	}
	plain := testkit.Plain(program)
	if !strings.Contains(plain, "Alice @alice") || !strings.Contains(plain, "Following") || !strings.Contains(plain, "1 of 3") {
		t.Fatalf("frame missing chrome:\n%s", plain)
	}
}

func TestOpeningHomeKeepsTheLoadedTimeline(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	testkit.SendKeys(program, "j")
	selected := root.feed.SelectedPost()
	if selected == nil {
		t.Fatal("timeline has no selection")
	}
	selectedID := selected.ID
	service.mu.Lock()
	loads := len(service.homeCalls)
	service.mu.Unlock()

	testkit.SendKeys(program, "1")
	if current := root.feed.SelectedPost(); current == nil || current.ID != selectedID {
		t.Fatalf("Home reset the selection: %#v", current)
	}
	service.mu.Lock()
	loadsAfterHome := len(service.homeCalls)
	service.mu.Unlock()
	if loadsAfterHome != loads {
		t.Fatalf("Home reloaded the timeline: before=%d after=%d", loads, loadsAfterHome)
	}

	testkit.SendKeys(program, "a", "1")
	if root.page.kind != timelinePage {
		t.Fatalf("Home did not return from profile: %v", root.page.kind)
	}
	if current := root.feed.SelectedPost(); current == nil || current.ID != selectedID {
		t.Fatalf("Home lost the preserved selection: %#v", current)
	}
	service.mu.Lock()
	loadsAfterProfile := len(service.homeCalls)
	service.mu.Unlock()
	if loadsAfterProfile != loads {
		t.Fatalf("Home reloaded after profile: before=%d after=%d", loads, loadsAfterProfile)
	}
}

func TestReplyAndQuoteUseDisplayedPost(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.SendKeys(program, "r")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Reply to @user1") || !strings.Contains(plain, "first") {
		t.Fatalf("reply dialog missing context:\n%s", plain)
	}
	testkit.SendKeys(program, "o", "k", "ctrl+s")
	if len(service.created) != 1 || service.created[0].Reply == nil || service.created[0].Reply.InReplyToPostID != "1" {
		t.Fatalf("reply = %#v", service.created)
	}

	testkit.SendKeys(program, "Q", "y", "e", "s", "ctrl+s")
	if len(service.created) != 2 || service.created[1].QuotePostID != "1" {
		t.Fatalf("quote = %#v", service.created)
	}
}

func TestProfileOpensFromUserNameAndKeyboard(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.Click(program, 5, 2)
	if plain := testkit.Plain(program); !strings.Contains(plain, "Profile") || !strings.Contains(plain, "@user1") || !strings.Contains(plain, "Profile for @user1") || strings.ContainsAny(plain, "╔╗║") {
		t.Fatalf("profile from user name:\n%s", plain)
	}
	if root.page.kind != profilePage {
		t.Fatalf("page = %v", root.page.kind)
	}
	testkit.SendKeys(program, "esc", "j", "u")
	if plain := testkit.Plain(program); !strings.Contains(plain, "@user2") || !strings.Contains(plain, "Profile for @user2") {
		t.Fatalf("profile from keyboard:\n%s", plain)
	}
	testkit.SendKeys(program, "esc", "a")
	if plain := testkit.Plain(program); !strings.Contains(plain, "@alice") || !strings.Contains(plain, "Profile for @alice") {
		t.Fatalf("own profile from keyboard:\n%s", plain)
	}
	if got := strings.Join(service.userCalls, ","); got != "user1,user2,alice" {
		t.Fatalf("profile calls = %q", got)
	}
}

func TestCompactLayoutAndHelp(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(48, 16))
	program.Start()

	plain := testkit.Plain(program)
	if !strings.Contains(plain, "Previous") || !strings.Contains(plain, "1 of 3") {
		t.Fatalf("compact frame missing navigation:\n%s", plain)
	}
	testkit.Click(program, 40, 1)
	if root.feed.CurrentMode() != feed.Ranked {
		t.Fatalf("compact next click mode = %v", root.feed.CurrentMode())
	}
	testkit.Click(program, 4, 1)
	if root.feed.CurrentMode() != feed.Home {
		t.Fatalf("compact previous click mode = %v", root.feed.CurrentMode())
	}

	testkit.SendKeys(program, "?")
	plain = testkit.Plain(program)
	if !strings.Contains(plain, "TIMELINES") || !strings.Contains(plain, "NAVIGATION") {
		t.Fatalf("help is missing sections:\n%s", plain)
	}
	testkit.SendKeys(program, "ctrl+d")
	if plain = testkit.Plain(program); !strings.Contains(plain, "POSTS") {
		t.Fatalf("help is missing post keys:\n%s", plain)
	}
	testkit.SendKeys(program, "ctrl+d")
	for range 40 {
		if strings.Contains(testkit.Plain(program), "Delete your own post") {
			break
		}
		testkit.SendKeys(program, "j")
	}
	if plain = testkit.Plain(program); !strings.Contains(plain, "Delete your own post") {
		t.Fatalf("help is missing the delete key:\n%s", plain)
	}
	testkit.SendKeys(program, "G")
	plain = testkit.Plain(program)
	if !strings.Contains(plain, "nth setup") || !strings.Contains(plain, "Credentials") {
		t.Fatalf("help footer is missing sections:\n%s", plain)
	}
}

func TestWideLayout(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(220, 30))
	program.Start()

	plain := testkit.Plain(program)
	lines := strings.Split(plain, "\n")
	if strings.Contains(lines[0], "For you") || !strings.Contains(lines[1], "For you") || strings.Contains(lines[2], "For you") {
		t.Fatalf("timeline tabs do not have one row of vertical padding:\n%s", plain)
	}
	width, height := lipgloss.Size(program.View().Content)
	if width != 220 || height != 30 {
		t.Fatalf("frame size = %dx%d", width, height)
	}
	if root.contentLeft != 34 {
		t.Fatalf("content left = %d", root.contentLeft)
	}
	if strings.Contains(plain, ". refresh") {
		t.Fatalf("wide header still has the refresh button:\n%s", plain)
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, strings.Repeat(" ", root.contentLeft)) {
			t.Fatalf("wide content is not centered: %q", line)
		}
	}
	for _, label := range []string{
		"north", "Home", "For you", "Search north", "SELECTED POST", "u/U · Enter",
		"ACTIONS", "r  ↩ Reply", "t  ↻ Repost", "l  ♡ Like", "Q  ❝ Quote", "0 Followers",
	} {
		if !strings.Contains(plain, label) {
			t.Errorf("wide frame missing %q:\n%s", label, plain)
		}
	}
	if resized := testkit.RenderAt(program, 152, 30); root.contentLeft != 0 || lipgloss.Width(strings.Split(resized, "\n")[0]) != 152 {
		t.Fatalf("resized content left = %d", root.contentLeft)
	}
	testkit.RenderAt(program, 220, 30)

	testkit.Click(program, root.contentLeft+wideSidebarWidth+4, 0)
	if root.feed.CurrentMode() != feed.Ranked {
		t.Fatalf("timeline tab mode = %v", root.feed.CurrentMode())
	}
	testkit.Click(program, root.contentLeft+wideSidebarWidth+80, 0)
	if root.feed.CurrentMode() != feed.Home {
		t.Fatalf("following tab mode = %v", root.feed.CurrentMode())
	}
	testkit.Click(program, root.contentLeft+8, 3)
	if root.feed.CurrentMode() != feed.Home {
		t.Fatalf("sidebar home click mode = %v", root.feed.CurrentMode())
	}
	testkit.Click(program, root.contentLeft+4, root.layoutHeight-5)
	if plain := testkit.Plain(program); !strings.Contains(plain, "Profile for @alice") || strings.ContainsAny(plain, "╔╗║") {
		t.Fatalf("account profile did not open:\n%s", plain)
	}
	testkit.Click(program, root.contentLeft+wideSidebarWidth+2, 0)
	if root.page.kind != timelinePage {
		t.Fatalf("profile back button left page %v active", root.page.kind)
	}
	contentWidth := min(root.layoutWidth, wideLayoutMaxWidth)
	asideLeft := root.contentLeft + contentWidth - wideAsideWidth
	hits := root.aside.(*inspector).build(wideAsideWidth, 29)
	testkit.Click(program, asideLeft+hits.authorLeft, hits.authorNameRow)
	if plain := testkit.Plain(program); !strings.Contains(plain, "Profile for @user1") {
		t.Fatalf("inspector profile did not open:\n%s", plain)
	}
	testkit.SendKeys(program, "left")
	if root.page.kind != timelinePage {
		t.Fatalf("left did not return from profile page: %v", root.page.kind)
	}
	testkit.Click(program, asideLeft+4, hits.actionFirst+1)
	if len(service.liked) != 1 || service.liked[0] != "1" {
		t.Fatalf("inspector like calls = %#v", service.liked)
	}
	testkit.Click(program, asideLeft+4, 2)
	testkit.SendKeys(program, "n", "o", "r", "t", "h")
	search, ok := root.page.component.(*searchpage.Screen)
	if root.page.kind != searchPage || !ok || search.Query() != "north" {
		t.Fatalf("aside search page = %v component = %T", root.page.kind, root.page.component)
	}
}

func TestWideLayoutTrendsOpenSearch(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	trends := &fakeTrendAPI{items: []north.Trend{
		{Tag: "golang", Count: 42, IsHashtag: true},
		{Tag: "North", Count: 12},
	}}
	root := newRootWithTrends(service, trends)
	program := reactea.New(modal.New(root), reactea.WithSize(220, 30))
	program.Start()

	plain := testkit.Plain(program)
	for _, want := range []string{"TRENDS FOR YOU", "#golang", "42 posts", "North", "12 posts"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("trend sidebar is missing %q:\n%s", want, plain)
		}
	}
	contentWidth := min(root.layoutWidth, wideLayoutMaxWidth)
	asideLeft := root.contentLeft + contentWidth - wideAsideWidth
	layout := root.aside.(*inspector).build(wideAsideWidth, root.layoutHeight-1)
	if len(layout.trends) != 2 {
		t.Fatalf("trend hits = %#v", layout.trends)
	}
	testkit.Click(program, asideLeft+4, layout.trends[0].first)
	search, ok := root.page.component.(*searchpage.Screen)
	if !ok {
		t.Fatalf("trend search = %T", root.page.component)
	}
	if search.Query() != "#golang" {
		t.Fatalf("trend query = %q", search.Query())
	}
	testkit.SendKeys(program, "esc", "7")
	search, ok = root.page.component.(*searchpage.Screen)
	if !ok {
		t.Fatalf("keyboard trend search = %T", root.page.component)
	}
	if search.Query() != "North" {
		t.Fatalf("keyboard trend query = %q", search.Query())
	}
}

func TestSidebarPaddingIsClickable(t *testing.T) {
	t.Parallel()

	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(220, 30))
	program.Start()

	testkit.Click(program, root.contentLeft+8, 6)
	if root.page.kind != searchPage {
		t.Fatalf("blank row below Search opened page %v", root.page.kind)
	}
	testkit.Click(program, root.contentLeft+8, 8)
	if root.page.kind != listsPage {
		t.Fatalf("blank row below Lists opened page %v", root.page.kind)
	}
}

type fakeTrendAPI struct {
	items []north.Trend
	err   error
}

func (f *fakeTrendAPI) Trends(context.Context, string) ([]north.Trend, *north.Response, error) {
	return append([]north.Trend(nil), f.items...), nil, f.err
}

func TestRateLimitStatusExplainsTheNumbers(t *testing.T) {
	t.Parallel()

	response := (&fakeAPI{}).response()
	if got := statusWithRate("", response); got != "last request: 179/180 left" {
		t.Fatalf("rate status = %q", got)
	}
}

func TestComposeKeepsDraftUntilSuccessfulPost(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{createErr: errors.New("offline")}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.SendKeys(program, "n", "d", "r", "a", "f", "t", "esc")
	if root.drafts["post"] != "draft" || !strings.Contains(testkit.Plain(program), "Draft saved") {
		t.Fatalf("draft after close = %#v\n%s", root.drafts, testkit.Plain(program))
	}
	testkit.SendKeys(program, "n")
	if plain := testkit.Plain(program); !strings.Contains(plain, "New post · Draft") || !strings.Contains(plain, "draft") {
		t.Fatalf("saved draft was not restored:\n%s", plain)
	}
	testkit.SendKeys(program, "ctrl+j")
	if root.drafts["post"] != "draft" || !strings.Contains(testkit.Plain(program), "Post failed · draft kept") {
		t.Fatalf("failed post lost draft: %#v\n%s", root.drafts, testkit.Plain(program))
	}

	service.mu.Lock()
	service.createErr = nil
	service.mu.Unlock()
	testkit.SendKeys(program, "n", "ctrl+s")
	if _, exists := root.drafts["post"]; exists {
		t.Fatalf("successful post kept draft: %#v", root.drafts)
	}
}

func TestComposeDraftsAreScopedToTheirTarget(t *testing.T) {
	t.Parallel()

	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.SendKeys(program, "n", "n", "e", "w", "esc")
	testkit.SendKeys(program, "r", "r", "e", "p", "l", "y", "esc")
	testkit.SendKeys(program, "Q", "q", "u", "o", "t", "e", "esc")
	for key, want := range map[string]string{"post": "new", "reply:1": "reply", "quote:1": "quote"} {
		if got := root.drafts[key]; got != want {
			t.Errorf("draft %q = %q, want %q", key, got, want)
		}
	}
}

func TestComposeAndSearchButtons(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.SendKeys(program, "n", "c", "l", "i", "c", "k")
	testkit.Click(program, 72, 18)
	if len(service.created) != 1 || service.created[0].Text != "click" {
		t.Fatalf("Post button created = %#v", service.created)
	}

	testkit.SendKeys(program, "/", "n", "o", "r", "t", "h")
	search, ok := root.page.component.(*searchpage.Screen)
	if root.page.kind != searchPage || !ok || search.Query() != "north" {
		t.Fatalf("search page = %v component = %T", root.page.kind, root.page.component)
	}
	if len(service.searchCalls) == 0 || service.searchCalls[len(service.searchCalls)-1] != "north" {
		t.Fatalf("search calls = %#v", service.searchCalls)
	}
}

func TestPostClickAndDetailAction(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	stack := modal.New(root)
	program := reactea.New(stack, reactea.WithSize(80, 24))
	program.Start()

	testkit.Click(program, 5, 3)
	plain := testkit.Plain(program)
	if !strings.Contains(plain, "Post") || !strings.Contains(plain, "r  Reply") {
		t.Fatalf("post details did not open:\n%s", plain)
	}
	if stack.Top() != nil || strings.Contains(plain, "second") {
		t.Fatalf("post details still use an overlay over the timeline:\n%s", plain)
	}
	if root.page.kind != postPage || root.page.key != "1" {
		t.Fatalf("post details are not a page: kind=%v key=%q", root.page.kind, root.page.key)
	}
	clickLastText(t, program, "User 1")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Profile for @user1") {
		t.Fatalf("profile from post details did not open:\n%s", plain)
	}
	testkit.SendKeys(program, "esc")
	if plain := testkit.Plain(program); !strings.Contains(plain, "first") || strings.Contains(plain, "Profile for @user1") || !strings.Contains(plain, "r  Reply") {
		t.Fatalf("profile did not return to the post:\n%s", plain)
	}
	clickLastText(t, program, "l  Like")
	if len(service.liked) != 1 || service.liked[0] != "1" {
		t.Fatalf("detail like calls = %#v", service.liked)
	}
	if plain := testkit.Plain(program); !strings.Contains(plain, "Unlike") || !strings.Contains(plain, "Liked") {
		t.Fatalf("detail did not remain open with updated state:\n%s", plain)
	}

	clickLastText(t, program, "r  Reply")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Reply to @user1") {
		t.Fatalf("reply click did not open the composer:\n%s", plain)
	}
	testkit.SendKeys(program, "r", "e", "p", "l", "y", "ctrl+s")
	if len(service.created) != 1 || service.created[0].Text != "reply" || service.created[0].Reply == nil || service.created[0].Reply.InReplyToPostID != "1" {
		t.Fatalf("reply from details = %#v", service.created)
	}
}

func TestQuotedPostOpensFromPostDetails(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	post := root.feed.SelectedPost()
	post.Quoted = &north.Post{
		ID:     "quoted",
		Text:   "quoted body",
		Source: "quoted client",
		Author: north.User{ID: "quoted-user", Handle: "quoted", Name: "Quoted"},
	}

	testkit.SendKeys(program, "enter")
	clickLastText(t, program, "quoted body")
	if plain := testkit.Plain(program); !strings.Contains(plain, "quoted client") {
		t.Fatalf("quoted post details did not open:\n%s", plain)
	}
	if root.page.kind != postPage || root.page.key != "quoted" {
		t.Fatalf("quoted post page = %v %q", root.page.kind, root.page.key)
	}
	testkit.SendKeys(program, "esc")
	if root.page.kind != postPage || root.page.key != "1" {
		t.Fatalf("quoted post did not return to its source: %v %q", root.page.kind, root.page.key)
	}
}

func TestReplySourceOpensAsAPostPage(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	parentID := "parent"
	root.feed.SelectedPost().InReplyToID = &parentID

	testkit.SendKeys(program, "enter")
	clickLastText(t, program, "parent post")
	if root.page.kind != postPage || root.page.key != parentID {
		t.Fatalf("reply source page = %v %q", root.page.kind, root.page.key)
	}
	testkit.SendKeys(program, "esc")
	if root.page.kind != postPage || root.page.key != "1" {
		t.Fatalf("reply source did not return to the reply: %v %q", root.page.kind, root.page.key)
	}
	testkit.SendKeys(program, "p")
	if root.page.kind != postPage || root.page.key != parentID {
		t.Fatalf("reply source keyboard navigation = %v %q", root.page.kind, root.page.key)
	}
}

func TestProfilePostOpensWithEnter(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 28))
	program.Start()
	testkit.SendKeys(program, "a", "enter")

	if plain := testkit.Plain(program); !strings.Contains(plain, "my post") || !strings.Contains(plain, "r  Reply") {
		t.Fatalf("profile post details did not open:\n%s", plain)
	}
	if root.page.kind != postPage || len(root.history) == 0 || root.history[len(root.history)-1].kind != profilePage {
		t.Fatalf("profile post page/history = %v %#v", root.page.kind, root.history)
	}
}

func TestReplyClickFocusesItsPostAndComposer(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.Click(program, 5, 8)
	if post := root.feed.SelectedPost(); post == nil || post.ID != "2" {
		t.Fatalf("selected after reply click = %#v", post)
	}
	if plain := testkit.Plain(program); !strings.Contains(plain, "Reply to @user2") {
		t.Fatalf("reply composer did not open for the clicked post:\n%s", plain)
	}

	testkit.SendKeys(program, "k", "e", "y", "b", "o", "a", "r", "d", "ctrl+s")
	if len(service.created) != 1 || service.created[0].Text != "keyboard" || service.created[0].Reply == nil || service.created[0].Reply.InReplyToPostID != "2" {
		t.Fatalf("keyboard reply = %#v", service.created)
	}
}

func TestOwnPostCanBeDeletedAfterConfirmation(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{me: north.User{ID: "user-1", Handle: "user1", Name: "User 1"}}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.Click(program, 5, 3)
	if plain := testkit.Plain(program); !strings.Contains(plain, "d  Delete") {
		t.Fatalf("delete action missing from own post:\n%s", plain)
	}
	testkit.SendKeys(program, "d")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Delete post?") || !strings.Contains(plain, "This cannot be undone.") || !strings.Contains(plain, "Enter  Delete") {
		t.Fatalf("delete confirmation missing:\n%s", plain)
	}
	testkit.SendKeys(program, "enter")

	if len(service.deleted) != 1 || service.deleted[0] != "1" {
		t.Fatalf("deleted posts = %#v", service.deleted)
	}
	if post := root.feed.SelectedPost(); post == nil || post.ID == "1" {
		t.Fatalf("selected post after deletion = %#v", post)
	}
	if root.page.kind != timelinePage || root.notice != "Post deleted" {
		t.Fatalf("delete did not return to the timeline: page=%v notice=%q", root.page.kind, root.notice)
	}
}

func TestDeleteConfirmationOnlyTargetsTheActivePostPage(t *testing.T) {
	t.Parallel()

	parentID := "parent"
	service := &fakeAPI{
		me: north.User{ID: "user-1", Handle: "user1", Name: "User 1"},
		posts: map[string]north.Post{
			parentID: {
				ID: parentID, Text: "owned parent", Author: north.User{ID: "user-1", Handle: "user1", Name: "User 1"},
			},
		},
	}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	root.feed.SelectedPost().InReplyToID = &parentID

	testkit.SendKeys(program, "enter")
	clickLastText(t, program, "owned parent")
	testkit.SendKeys(program, "d", "enter")
	if got := strings.Join(service.deleted, ","); got != parentID {
		t.Fatalf("delete targets = %q", got)
	}
	if root.page.kind != postPage || root.page.key != "1" {
		t.Fatalf("delete did not return to the reply: %v %q", root.page.kind, root.page.key)
	}
}

func TestDeleteCanBeRequestedAgainAfterCancelingConfirmation(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{me: north.User{ID: "user-1", Handle: "user1", Name: "User 1"}}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.SendKeys(program, "enter", "d", "esc", "d")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Delete post?") {
		t.Fatalf("delete confirmation did not reopen after cancellation:\n%s", plain)
	}
	if len(service.deleted) != 0 {
		t.Fatalf("cancelled deletion reached the API: %#v", service.deleted)
	}
}

func TestPostActionUsesMessageTarget(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	testkit.SendKeys(program, "j")

	program.Send(postcomponent.ActionMsg{Action: postcomponent.Like, Post: testPost("1", "first")})
	if len(service.liked) != 1 || service.liked[0] != "1" {
		t.Fatalf("reaction target = %#v", service.liked)
	}
}

func TestResponsiveLayouts(t *testing.T) {
	t.Parallel()

	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(121, 24))
	program.Start()
	if !root.medium || root.wide || root.contentWidth != mediumLayoutMaxWidth || root.contentLeft != 16 {
		t.Fatalf("121 columns: medium=%v wide=%v content=%dx%d", root.medium, root.wide, root.contentLeft, root.contentWidth)
	}
	testkit.RenderAt(program, 122, 24)
	if !root.medium || root.contentWidth != mediumLayoutMaxWidth || root.contentLeft != 17 {
		t.Fatalf("122 columns: medium=%v content=%dx%d", root.medium, root.contentLeft, root.contentWidth)
	}
	testkit.RenderAt(program, 160, 10)
	if root.wide || !root.medium || root.contentLeft != 36 {
		t.Fatalf("short terminal: medium=%v wide=%v left=%d", root.medium, root.wide, root.contentLeft)
	}
	testkit.RenderAt(program, 152, 24)
	if !root.wide || root.contentLeft != 0 {
		t.Fatalf("wide terminal: wide=%v left=%d", root.wide, root.contentLeft)
	}
}

func TestAccountFailureCanBeRetried(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{meErr: errors.New("account offline")}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	if root.me != nil || !strings.Contains(testkit.Plain(program), "account offline") {
		t.Fatalf("account failure not shown:\n%s", testkit.Plain(program))
	}

	service.mu.Lock()
	service.meErr = nil
	service.mu.Unlock()
	testkit.SendKeys(program, ".")
	if root.me == nil || service.meCalls < 2 {
		t.Fatalf("account retry: me=%#v calls=%d", root.me, service.meCalls)
	}
}

func TestNotificationsOpenAndMarkTheFeedRead(t *testing.T) {
	t.Parallel()

	service := &notificationAppAPI{
		fakeAPI: &fakeAPI{},
		unread:  3,
		page: notification.Page{Items: []notification.Item{{
			ID:         "like-1",
			Kind:       notification.Like,
			Actors:     []north.User{{Name: "Bob", Handle: "bob"}},
			ActorCount: 1,
			Post:       postPointer(testPost("notice", "from a notification")),
		}},
		},
	}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	if root.unread != 3 {
		t.Fatalf("unread count = %d", root.unread)
	}

	testkit.SendKeys(program, "2")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Bob liked your post") || !strings.Contains(plain, "from a notification") || strings.ContainsAny(plain, "╔╗║") {
		t.Fatalf("notification page did not open:\n%s", plain)
	}
	if root.page.kind != notificationsPage {
		t.Fatalf("page = %v", root.page.kind)
	}
	if service.markedRead != 1 || root.unread != 0 {
		t.Fatalf("read state: calls=%d unread=%d", service.markedRead, root.unread)
	}

	testkit.SendKeys(program, "u")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Profile for @bob") || strings.ContainsAny(plain, "╔╗║") {
		t.Fatalf("notification profile page did not open:\n%s", plain)
	}
	testkit.SendKeys(program, "esc")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Bob liked your post") {
		t.Fatalf("notification history was not restored:\n%s", plain)
	}
	testkit.SendKeys(program, "enter")
	if plain := testkit.Plain(program); !strings.Contains(plain, "from a notification") || !strings.Contains(plain, "r  Reply") {
		t.Fatalf("notification post did not open:\n%s", plain)
	}
}

func TestNotificationsAreHiddenWithoutNotificationAPI(t *testing.T) {
	t.Parallel()

	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	if plain := testkit.Plain(program); strings.Contains(plain, "Notifications") || strings.Contains(plain, "2 notices") {
		t.Fatalf("compact UI exposed unavailable notifications:\n%s", plain)
	}
	testkit.Click(program, 75, 1)
	if root.feed.CurrentMode() != feed.Home {
		t.Fatalf("following compact tab = %v, want Home", root.feed.CurrentMode())
	}
	testkit.SendKeys(program, "?")
	if plain := testkit.Plain(program); strings.Contains(plain, "Open notifications") {
		t.Fatalf("help exposed unavailable notifications:\n%s", plain)
	}
	testkit.SendKeys(program, "esc", "2")
	if root.page.kind != timelinePage || root.notice != "" {
		t.Fatalf("hidden notification shortcut changed state: page=%v notice=%q", root.page.kind, root.notice)
	}

	testkit.RenderAt(program, 180, 24)
	if plain := testkit.Plain(program); strings.Contains(plain, "Notifications") || strings.Contains(plain, "2 notices") {
		t.Fatalf("wide UI exposed unavailable notifications:\n%s", plain)
	}
}

func TestOwnEligiblePostCanBeEdited(t *testing.T) {
	t.Parallel()

	service := &editorAppAPI{fakeAPI: &fakeAPI{me: north.User{ID: "user-1", Handle: "user1", Name: "User 1"}}, eligible: true}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()

	testkit.SendKeys(program, "enter")
	if plain := testkit.Plain(program); !strings.Contains(plain, "e  Edit") {
		t.Fatalf("eligible post has no edit action:\n%s", plain)
	}
	testkit.SendKeys(program, "e")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Edit post") || !strings.Contains(plain, "first") {
		t.Fatalf("edit composer did not open:\n%s", plain)
	}
	for range len("first") {
		testkit.SendKeys(program, "backspace")
	}
	testkit.SendKeys(program, "u", "p", "d", "a", "t", "e", "d", "ctrl+s")
	if len(service.edits) != 1 || service.edits[0].id != "1" || service.edits[0].text != "updated" {
		t.Fatalf("edit calls = %#v", service.edits)
	}
	post := root.feed.SelectedPost()
	if post == nil || post.Text != "updated" || post.EditedAt == nil {
		t.Fatalf("edited timeline post = %#v", post)
	}
}

func TestHiddenPostDoesNotLeakThroughInspectorOrComposer(t *testing.T) {
	t.Parallel()

	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(180, 24))
	program.Start()
	post := root.feed.SelectedPost()
	post.Text = "SHOULD_NOT_RENDER"
	post.HiddenReason = north.HiddenReason("MUTED")

	if plain := testkit.Plain(program); strings.Contains(plain, "SHOULD_NOT_RENDER") || !strings.Contains(plain, "Hidden: MUTED") {
		t.Fatalf("wide inspector leaked a hidden post:\n%s", plain)
	}
	testkit.SendKeys(program, "r", "Q")
	if plain := testkit.Plain(program); strings.Contains(plain, "SHOULD_NOT_RENDER") || strings.Contains(plain, "Reply to @") || strings.Contains(plain, "Quote @") {
		t.Fatalf("hidden post reached a composer:\n%s", plain)
	}
}

func TestKeyboardCanOpenEveryLinkedProfile(t *testing.T) {
	t.Parallel()

	service := &fakeAPI{}
	root := newRoot(service)
	program := reactea.New(modal.New(root), reactea.WithSize(80, 24))
	program.Start()
	reply := "reply"
	post := root.feed.SelectedPost()
	post.InReplyToHandle = &reply
	post.Quoted = &north.Post{Author: north.User{Handle: "quoted", Name: "Quoted"}}

	testkit.SendKeys(program, "U")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Profile for @reply") {
		t.Fatalf("reply account did not open from the keyboard:\n%s", plain)
	}
	testkit.SendKeys(program, "esc", "U")
	if plain := testkit.Plain(program); !strings.Contains(plain, "Profile for @quoted") {
		t.Fatalf("quoted account did not open from the keyboard:\n%s", plain)
	}
}

func TestWideLayoutKeepsTheCenterColumnAndInspectorActions(t *testing.T) {
	t.Parallel()

	if got := wideLayoutMinWidth - wideSidebarWidth - wideAsideWidth; got < mediumLayoutMaxWidth {
		t.Fatalf("wide center column shrinks from %d to %d", mediumLayoutMaxWidth, got)
	}
	root := newRoot(&fakeAPI{})
	program := reactea.New(modal.New(root), reactea.WithSize(wideLayoutMinWidth, wideLayoutMinHeight))
	program.Start()
	plain := testkit.Plain(program)
	for _, action := range []string{"r  ↩ Reply", "t  ↻ Repost", "l  ♡ Like", "Q  ❝ Quote"} {
		if !strings.Contains(plain, action) {
			t.Errorf("minimum wide layout hides %q:\n%s", action, plain)
		}
	}
}

func clickLastText(t *testing.T, program *reactea.App, label string) {
	t.Helper()
	lines := testkit.Lines(program)
	for y := len(lines) - 1; y >= 0; y-- {
		if x := strings.Index(lines[y], label); x >= 0 {
			testkit.Click(program, x, y)

			return
		}
	}
	t.Fatalf("could not find %q in frame:\n%s", label, testkit.Plain(program))
}

type notificationAppAPI struct {
	*fakeAPI
	page       notification.Page
	unread     int
	markedRead int
}

func (a *notificationAppAPI) Notifications(context.Context, north.NotificationTab, string) (notification.Page, *north.Response, error) {
	return a.page, a.response(), nil
}

func (a *notificationAppAPI) NotificationUnreadCount(context.Context) (int, *north.Response, error) {
	return a.unread, a.response(), nil
}

func (a *notificationAppAPI) MarkNotificationsRead(context.Context) (int, *north.Response, error) {
	a.markedRead++

	return a.markedRead, a.response(), nil
}

type editCall struct {
	id   string
	text string
}

type editorAppAPI struct {
	*fakeAPI
	eligible bool
	edits    []editCall
}

func (a *editorAppAPI) EditablePost(_ context.Context, id string) (north.Post, bool, *north.Response, error) {
	return testPost(id, "first"), a.eligible, a.response(), nil
}

func (a *editorAppAPI) EditPost(_ context.Context, id, text string, _ []string) (*north.Response, error) {
	a.edits = append(a.edits, editCall{id: id, text: text})

	return a.response(), nil
}

func postPointer(post north.Post) *north.Post { return &post }
