package app

import (
	"net/url"
	"strings"

	"github.com/Hayao0819/nth/internal/components/navigation"
)

type pageKind uint8

const (
	timelinePage pageKind = iota
	notificationsPage
	messagesPage
	bookmarksPage
	listsPage
	profilePage
	postPage
	searchPage
	savedPage
	accountPage
)

const (
	timelineRoute        = "/"
	notificationsRoute   = "/notifications"
	messagesRoute        = "/messages"
	bookmarksRoute       = "/bookmarks"
	bookmarkFoldersRoute = "/bookmarks/folders"
	listsRoute           = "/lists"
	savedRoute           = "/saved"
	accountSafetyRoute   = "/account/privacy"
)

type pageState struct {
	kind pageKind
	key  string
}

func pageStateForRoute(route string) pageState {
	state := pageState{}
	switch {
	case route == "" || route == timelineRoute:
		state.kind = timelinePage
	case route == notificationsRoute:
		state.kind, state.key = notificationsPage, "notifications"
	case route == messagesRoute:
		state.kind, state.key = messagesPage, "messages"
	case route == bookmarksRoute:
		state.kind, state.key = bookmarksPage, "bookmarks"
	case route == bookmarkFoldersRoute:
		state.kind, state.key = bookmarksPage, "folders"
	case strings.HasPrefix(route, bookmarkFoldersRoute+"/"):
		state.kind = bookmarksPage
		state.key = unescapeRoutePart(strings.TrimPrefix(route, bookmarkFoldersRoute+"/"))
	case route == listsRoute:
		state.kind, state.key = listsPage, "lists"
	case strings.HasPrefix(route, listsRoute+"/"):
		state.kind = listsPage
		state.key = unescapeRoutePart(strings.TrimPrefix(route, listsRoute+"/"))
	case route == savedRoute:
		state.kind, state.key = savedPage, "saved"
	case route == accountSafetyRoute:
		state.kind, state.key = accountPage, "privacy"
	case route == "/search" || strings.HasPrefix(route, "/search/"):
		state.kind, state.key = searchPage, "search"
	case strings.HasPrefix(route, "/users/"):
		state.kind = profilePage
		value := strings.TrimPrefix(route, "/users/")
		handle, _, _ := strings.Cut(value, "/")
		state.key = strings.ToLower(unescapeRoutePart(handle))
	case strings.HasPrefix(route, "/posts/"):
		state.kind = postPage
		value := strings.TrimPrefix(route, "/posts/")
		id, _, _ := strings.Cut(value, "/")
		state.key = unescapeRoutePart(id)
	default:
		state.kind = listsPage
		state.key = route
	}

	return state
}

func searchRoute(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return "/search"
	}

	return "/search/" + url.PathEscape(query)
}

func userRoute(handle string) string {
	return "/users/" + url.PathEscape(strings.ToLower(strings.TrimSpace(handle)))
}

func userConnectionsRoute(handle string, following bool) string {
	resource := "followers"
	if following {
		resource = "following"
	}

	return userRoute(handle) + "/" + resource
}

func postRoute(id string) string {
	return "/posts/" + url.PathEscape(strings.TrimSpace(id))
}

func postActivityRoute(id string, activity navigation.PostActivity) string {
	return postRoute(id) + "/" + postActivityName(activity)
}

func postActivityName(activity navigation.PostActivity) string {
	switch activity {
	case navigation.PostLikes:
		return "likes"
	case navigation.PostReposts:
		return "reposts"
	case navigation.PostHistory:
		return "history"
	default:
		return "quotes"
	}
}

func postActivityFromRoute(value string) navigation.PostActivity {
	switch value {
	case "likes":
		return navigation.PostLikes
	case "reposts":
		return navigation.PostReposts
	case "history":
		return navigation.PostHistory
	default:
		return navigation.PostQuotes
	}
}

func listRoute(id string) string {
	return listsRoute + "/" + url.PathEscape(strings.TrimSpace(id))
}

func listMembersRoute(id string) string {
	return listRoute(id) + "/members"
}

func bookmarkFolderRoute(id string) string {
	return bookmarkFoldersRoute + "/" + url.PathEscape(strings.TrimSpace(id))
}

func unescapeRoutePart(value string) string {
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}

	return decoded
}
