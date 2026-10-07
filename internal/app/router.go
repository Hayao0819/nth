package app

import (
	"strings"

	"github.com/Hayao0819/go-north"
	accountfeature "github.com/Hayao0819/nth/internal/features/account"
	bookmarkfeature "github.com/Hayao0819/nth/internal/features/bookmark"
	listfeature "github.com/Hayao0819/nth/internal/features/list"
	messagefeature "github.com/Hayao0819/nth/internal/features/message"
	notificationfeature "github.com/Hayao0819/nth/internal/features/notification"
	postfeature "github.com/Hayao0819/nth/internal/features/post"
	savedfeature "github.com/Hayao0819/nth/internal/features/saved"
	searchfeature "github.com/Hayao0819/nth/internal/features/search"
	userfeature "github.com/Hayao0819/nth/internal/features/user"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/router"
)

func (r *root) pageRoutes() router.Routes {
	routes := router.Routes{
		"/search/:query?": func(params router.Params) reactea.Component {
			return searchfeature.NewWithImages(r.api, r.theme, unescapeRoutePart(params.Get("query")), r.images)
		},
		"/users/:handle": func(params router.Params) reactea.Component {
			handle := unescapeRoutePart(params.Get("handle"))
			user := r.users[strings.ToLower(handle)]
			if strings.TrimSpace(user.Handle) == "" {
				user.Handle = handle
			}

			page := userfeature.NewPageWithImages(r.api, r.theme, user, r.images)
			if r.me != nil {
				page.SetViewer(*r.me)
			}

			return page
		},
		"/posts/:id": func(params router.Params) reactea.Component {
			id := unescapeRoutePart(params.Get("id"))
			if post, ok := r.posts[id]; ok {
				return r.newPostPage(post)
			}

			return newPostRoute(r, id)
		},
	}
	if r.postActivity != nil {
		routes["/posts/:id/:activity"] = func(params router.Params) reactea.Component {
			id := unescapeRoutePart(params.Get("id"))
			post := r.posts[id]
			if post.ID == "" {
				post.ID = id
			}

			return postfeature.NewActivity(r.postActivity, r.api, r.theme, post, postActivityFromRoute(params.Get("activity")), r.images)
		}
	}
	if r.connections != nil {
		routes["/users/:handle/followers"] = func(params router.Params) reactea.Component {
			handle := unescapeRoutePart(params.Get("handle"))
			user := r.users[strings.ToLower(handle)]
			if user.Handle == "" {
				user.Handle = handle
			}

			return userfeature.NewConnections(r.connections, r.theme, user, false)
		}
		routes["/users/:handle/following"] = func(params router.Params) reactea.Component {
			handle := unescapeRoutePart(params.Get("handle"))
			user := r.users[strings.ToLower(handle)]
			if user.Handle == "" {
				user.Handle = handle
			}

			return userfeature.NewConnections(r.connections, r.theme, user, true)
		}
	}
	if r.lists != nil {
		routes[listsRoute] = router.Page(func() reactea.Component {
			return listfeature.NewIndex(r.lists, r.theme)
		})
		routes["/lists/:id"] = func(params router.Params) reactea.Component {
			id := unescapeRoutePart(params.Get("id"))
			item := r.listItems[id]
			if item.ID == "" {
				item.ID = id
			}

			return listfeature.NewDetail(r.api, r.lists, r.theme, item, r.images)
		}
	}
	if r.listMembers != nil {
		routes["/lists/:id/members"] = func(params router.Params) reactea.Component {
			id := unescapeRoutePart(params.Get("id"))
			item := r.listItems[id]
			if item.ID == "" {
				item.ID = id
			}

			return listfeature.NewMembers(r.listMembers, r.theme, item)
		}
	}
	if r.notifications != nil {
		routes[notificationsRoute] = router.Page(func() reactea.Component {
			return notificationfeature.NewPageWithImages(r.notifications, r.theme, r.images)
		})
	}
	if r.messages != nil {
		routes[messagesRoute] = router.Page(func() reactea.Component {
			page := messagefeature.New(r.messages, r.theme)
			if r.me != nil {
				page.SetViewer(*r.me)
			}

			return page
		})
	}
	if r.bookmarks != nil {
		routes[bookmarksRoute] = router.Page(func() reactea.Component {
			page := bookmarkfeature.New(r.api, r.bookmarks, r.theme, r.images)
			page.SetFolderAPI(r.bookmarkFolders)

			return page
		})
	}
	if r.bookmarkFolders != nil {
		routes[bookmarkFoldersRoute] = router.Page(func() reactea.Component {
			return bookmarkfeature.NewFolderIndex(r.bookmarkFolders, r.theme)
		})
		routes[bookmarkFoldersRoute+"/:id"] = func(params router.Params) reactea.Component {
			id := unescapeRoutePart(params.Get("id"))
			folder := r.bookmarkItems[id]
			if folder.ID == "" {
				folder.ID = id
			}

			return bookmarkfeature.NewFolderDetail(r.api, r.bookmarkFolders, r.theme, folder, r.images)
		}
	}
	if r.savedPosts != nil {
		routes[savedRoute] = router.Page(func() reactea.Component {
			return savedfeature.New(r.savedPosts, r.theme)
		})
	}
	if r.accountSafety != nil {
		routes[accountSafetyRoute] = router.Page(func() reactea.Component {
			return accountfeature.New(r.accountSafety, r.mutedKeywords, r.theme)
		})
	}

	return routes
}

func (r *root) newPostPage(post north.Post) *postfeature.Screen {
	editor, _ := r.api.(postfeature.Editor)

	return postfeature.NewPageWithEditorAndImages(r.api, editor, r.theme, post, r.ownsPost(post), r.images)
}

func (r *root) currentPage() reactea.Component {
	if r.pages == nil {
		return nil
	}

	return r.pages.Current()
}
