package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	searchfeature "github.com/Hayao0819/nth/internal/features/search"
	"github.com/Hayao0819/reactea/v2"
)

func (r *root) openNotifications(ctx *reactea.Ctx) tea.Cmd {
	if r.notifications == nil {
		return nil
	}

	return r.pushRoute(ctx, notificationsRoute)
}

func (r *root) openMessages(ctx *reactea.Ctx) tea.Cmd {
	if r.messages == nil {
		return nil
	}

	return r.pushRoute(ctx, messagesRoute)
}

func (r *root) openBookmarks(ctx *reactea.Ctx) tea.Cmd {
	if r.bookmarks == nil {
		return nil
	}

	return r.pushRoute(ctx, bookmarksRoute)
}

func (r *root) openBookmarkFolders(ctx *reactea.Ctx) tea.Cmd {
	if r.bookmarkFolders == nil {
		return nil
	}

	return r.pushRoute(ctx, bookmarkFoldersRoute)
}

func (r *root) openBookmarkFolder(ctx *reactea.Ctx, folder north.BookmarkFolder) tea.Cmd {
	if r.bookmarkFolders == nil || strings.TrimSpace(folder.ID) == "" {
		return nil
	}
	r.bookmarkItems[folder.ID] = folder

	return r.pushRoute(ctx, bookmarkFolderRoute(folder.ID))
}

func (r *root) openSavedPosts(ctx *reactea.Ctx) tea.Cmd {
	if r.savedPosts == nil {
		return nil
	}

	return r.pushRoute(ctx, savedRoute)
}

func (r *root) openAccountSafety(ctx *reactea.Ctx) tea.Cmd {
	if r.accountSafety == nil {
		return nil
	}

	return r.pushRoute(ctx, accountSafetyRoute)
}

func (r *root) openLists(ctx *reactea.Ctx) tea.Cmd {
	if r.lists == nil {
		return nil
	}

	return r.pushRoute(ctx, listsRoute)
}

func (r *root) openList(ctx *reactea.Ctx, item north.List) tea.Cmd {
	if r.lists == nil || strings.TrimSpace(item.ID) == "" {
		return nil
	}
	r.listItems[item.ID] = item

	return r.pushRoute(ctx, listRoute(item.ID))
}

func (r *root) openListMembers(ctx *reactea.Ctx, item north.List) tea.Cmd {
	if r.listMembers == nil || strings.TrimSpace(item.ID) == "" {
		return nil
	}
	r.listItems[item.ID] = item

	return r.pushRoute(ctx, listMembersRoute(item.ID))
}

func (r *root) search(ctx *reactea.Ctx) tea.Cmd {
	current := ""
	if page, ok := r.currentPage().(*searchfeature.Screen); ok {
		current = page.Query()
	} else if r.feed.CurrentMode() == feed.Search {
		current = r.feed.SearchQuery()
	}

	return r.searchFor(ctx, current)
}

func (r *root) searchFor(ctx *reactea.Ctx, query string) tea.Cmd {
	return r.pushRoute(ctx, searchRoute(query))
}

func (r *root) cycleTimeline(ctx *reactea.Ctx, delta int) tea.Cmd {
	modes := [...]feed.Mode{feed.Home, feed.Ranked}
	index := 0
	for i, mode := range modes {
		if mode == r.feed.CurrentMode() {
			index = i
			break
		}
	}
	if r.feed.CurrentMode() == feed.Search {
		if delta < 0 {
			index = len(modes) - 1
		} else {
			index = 0
		}
	} else {
		index = (index + delta + len(modes)) % len(modes)
	}

	return tea.Batch(r.showTimeline(ctx), r.feed.SetMode(ctx, modes[index], ""))
}

func (r *root) openTimeline(ctx *reactea.Ctx, mode feed.Mode) tea.Cmd {
	routeCommand := r.showTimeline(ctx)
	if r.feed.CurrentMode() == feed.Home || r.feed.CurrentMode() == feed.Ranked {
		return routeCommand
	}

	return tea.Batch(routeCommand, r.feed.SetMode(ctx, mode, ""))
}

func (r *root) openPost(ctx *reactea.Ctx, post north.Post) tea.Cmd {
	key := post.ID
	if target := post.DisplayPost(); target != nil && strings.TrimSpace(target.ID) != "" {
		key = target.ID
	}
	if strings.TrimSpace(key) == "" {
		r.notice = "This post is unavailable"

		return nil
	}

	r.posts[key] = post

	return r.pushRoute(ctx, postRoute(key))
}

func (r *root) openPostActivity(ctx *reactea.Ctx, post north.Post, activity navigation.PostActivity) tea.Cmd {
	if r.postActivity == nil {
		return nil
	}
	id := post.ID
	if target := post.DisplayPost(); target != nil && strings.TrimSpace(target.ID) != "" {
		id = target.ID
	}
	if strings.TrimSpace(id) == "" {
		return nil
	}
	r.posts[id] = post

	return r.pushRoute(ctx, postActivityRoute(id, activity))
}

func (r *root) openUser(ctx *reactea.Ctx, user north.User) tea.Cmd {
	handle := strings.TrimSpace(user.Handle)
	if handle == "" {
		r.notice = "This account is unavailable"

		return nil
	}

	r.users[strings.ToLower(handle)] = user

	return r.pushRoute(ctx, userRoute(handle))
}

func (r *root) openUserConnections(ctx *reactea.Ctx, user north.User, following bool) tea.Cmd {
	handle := strings.TrimSpace(user.Handle)
	if r.connections == nil || handle == "" {
		return nil
	}
	r.users[strings.ToLower(handle)] = user

	return r.pushRoute(ctx, userConnectionsRoute(handle, following))
}

func (r *root) openNextLinkedUser(ctx *reactea.Ctx, post north.Post) tea.Cmd {
	users := postcomponent.LinkedUsers(post)
	if len(users) == 0 {
		r.notice = "This post has no available account"

		return nil
	}
	target := post.DisplayPost()
	key := post.ID
	if target != nil && target.ID != "" {
		key = target.ID
	}
	index := r.linkedUser[key]
	if len(users) > 1 {
		index = (index + 1) % len(users)
	} else {
		index = 0
	}
	r.linkedUser[key] = index

	return r.openUser(ctx, users[index])
}
