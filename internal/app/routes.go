package app

import (
	"net/url"
	"strings"
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
)

const (
	timelineRoute      = "/"
	notificationsRoute = "/notifications"
	messagesRoute      = "/messages"
	bookmarksRoute     = "/bookmarks"
	listsRoute         = "/lists"
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
	case route == listsRoute:
		state.kind, state.key = listsPage, "lists"
	case route == "/search" || strings.HasPrefix(route, "/search/"):
		state.kind, state.key = searchPage, "search"
	case strings.HasPrefix(route, "/users/"):
		state.kind = profilePage
		state.key = strings.ToLower(unescapeRoutePart(strings.TrimPrefix(route, "/users/")))
	case strings.HasPrefix(route, "/posts/"):
		state.kind = postPage
		state.key = unescapeRoutePart(strings.TrimPrefix(route, "/posts/"))
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

func postRoute(id string) string {
	return "/posts/" + url.PathEscape(strings.TrimSpace(id))
}

func unescapeRoutePart(value string) string {
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}

	return decoded
}
