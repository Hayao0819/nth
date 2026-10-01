package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type navigationAction uint8

const (
	navigateTimeline navigationAction = iota
	navigateSearch
	navigateCompose
	navigateNotifications
	navigateMessages
	navigateBookmarks
	navigateLists
	navigateProfile
	navigateSettings
)

type navigationMsg struct {
	action navigationAction
	mode   feed.Mode
	query  string
}

type sidebar struct {
	reactea.BasicComponent
	root *root
}

type sidebarRenderKey struct {
	pageKind      pageKind
	pageKey       string
	mode          feed.Mode
	unread        int
	unreadFailed  bool
	notifications bool
	messages      bool
	bookmarks     bool
	account       bool
	name          string
	handle        string
	following     int
	followers     int
}

type sidebarHit struct {
	first int
	last  int
	event navigationMsg
}

type sidebarLayout struct {
	content      string
	hits         []sidebarHit
	accountFirst int
}

func (s *sidebar) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return nil
	}
	_, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil
	}
	layout := s.build(ctx.Width(), ctx.Height())
	for _, hit := range layout.hits {
		if y >= hit.first && y <= hit.last {
			event := hit.event

			return func() tea.Msg { return event }
		}
	}
	if s.root.me != nil && y >= layout.accountFirst {
		user := *s.root.me

		return navigation.OpenUser(user)
	}

	return nil
}

func (s *sidebar) Render(ctx *reactea.Ctx) string {
	return s.build(ctx.Width(), ctx.Height()).content
}

func (s *sidebar) renderKey() sidebarRenderKey {
	r := s.root
	key := sidebarRenderKey{
		pageKind:      r.page.kind,
		pageKey:       r.page.key,
		mode:          r.feed.CurrentMode(),
		unread:        r.unread,
		unreadFailed:  r.unreadErr != nil,
		notifications: r.notifications != nil,
		messages:      r.messages != nil,
		bookmarks:     r.bookmarks != nil,
		account:       r.me != nil,
	}
	if r.me != nil {
		key.name = r.me.Name
		key.handle = r.me.Handle
		key.following = r.me.FollowingCount
		key.followers = r.me.FollowerCount
	}

	return key
}

func (s *sidebar) build(width, height int) sidebarLayout {
	r := s.root
	innerWidth := max(1, width-4)
	onTimeline := r.page.kind == timelinePage
	homeActive := onTimeline && (r.feed.CurrentMode() == feed.Home || r.feed.CurrentMode() == feed.Ranked)
	lines := []string{
		"",
		"  " + r.theme.Brand.Render("north"),
		"",
	}
	result := sidebarLayout{accountFirst: height}
	result.hits = append(result.hits, sidebarHit{first: 0, last: 2, event: navigationMsg{action: navigateTimeline, mode: feed.Home}})
	addItem := func(active bool, key, label string, event navigationMsg) {
		first := len(lines)
		lines = append(lines, sidebarItem(active, key, label, innerWidth, r.theme), "")
		result.hits = append(result.hits, sidebarHit{first: first, last: first + 1, event: event})
	}
	addItem(homeActive, "1", "Home", navigationMsg{action: navigateTimeline, mode: feed.Home})
	addItem(r.page.kind == searchPage, "/", "Explore", navigationMsg{action: navigateSearch})
	if r.notifications != nil {
		addItem(r.page.kind == notificationsPage, "2", notificationLabel(r), navigationMsg{action: navigateNotifications})
	}
	if r.messages != nil {
		addItem(r.page.kind == messagesPage, "3", "Messages", navigationMsg{action: navigateMessages})
	}
	if r.bookmarks != nil {
		addItem(r.page.kind == bookmarksPage, "4", "Bookmarks", navigationMsg{action: navigateBookmarks})
	}
	addItem(r.page.kind == listsPage, "", "Lists", navigationMsg{action: navigateLists})
	addItem(r.page.kind == profilePage && r.me != nil && strings.EqualFold(r.page.key, r.me.Handle), "", "Profile", navigationMsg{action: navigateProfile})
	addItem(r.page.kind == settingsPage, "", "Settings", navigationMsg{action: navigateSettings})
	first := len(lines)
	lines = append(lines, "  "+r.theme.Button.Width(min(18, innerWidth)).Align(lipgloss.Center).Render("n  Post"), "")
	result.hits = append(result.hits, sidebarHit{
		first: first,
		last:  first + 1,
		event: navigationMsg{action: navigateCompose},
	})

	var account []string
	if r.me != nil {
		account = []string{
			"  " + r.theme.Name.Render(ui.Clip(ui.SafeInline(r.me.Name), innerWidth)),
			"  " + r.theme.Handle.Render(ui.Clip("@"+ui.SafeInline(r.me.Handle), innerWidth)),
			"  " + r.theme.Dim.Render(ui.Clip(fmt.Sprintf("%d Following", r.me.FollowingCount), innerWidth)),
			"  " + r.theme.Dim.Render(ui.Clip(fmt.Sprintf("%d Followers", r.me.FollowerCount), innerWidth)),
		}
	}
	for len(lines)+len(account) < height {
		lines = append(lines, "")
	}
	result.accountFirst = len(lines)
	lines = append(lines, account...)
	result.content = ui.Fit(strings.Join(lines, "\n"), width, height)

	return result
}

func notificationLabel(r *root) string {
	if r.unreadErr != nil {
		return "Notifications !"
	}
	if r.unread > 0 {
		return fmt.Sprintf("Notifications (%d)", r.unread)
	}

	return "Notifications"
}

func sidebarItem(active bool, key, label string, width int, theme ui.Theme) string {
	marker := " "
	style := theme.Heading
	if active {
		marker = "●"
		style = theme.Active
	}
	line := style.Render(marker + "  " + label)

	return "  " + ui.Sides(line, theme.Dim.Render(key), width)
}

var _ reactea.Component = (*sidebar)(nil)
