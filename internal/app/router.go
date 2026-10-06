package app

import (
	"strings"

	"github.com/Hayao0819/go-north"
	bookmarkpage "github.com/Hayao0819/nth/internal/pages/bookmark"
	infopage "github.com/Hayao0819/nth/internal/pages/info"
	messagepage "github.com/Hayao0819/nth/internal/pages/message"
	notificationpage "github.com/Hayao0819/nth/internal/pages/notification"
	postpage "github.com/Hayao0819/nth/internal/pages/post"
	searchpage "github.com/Hayao0819/nth/internal/pages/search"
	userpage "github.com/Hayao0819/nth/internal/pages/user"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/router"
)

func (r *root) pageRoutes() router.Routes {
	routes := router.Routes{
		listsRoute: router.Page(func() reactea.Component {
			return infopage.New(r.theme, "Lists", "Lists are not available in this version of nth.")
		}),
		"/search/:query?": func(params router.Params) reactea.Component {
			return searchpage.NewWithImages(r.api, r.theme, unescapeRoutePart(params.Get("query")), r.images)
		},
		"/users/:handle": func(params router.Params) reactea.Component {
			handle := unescapeRoutePart(params.Get("handle"))
			user := r.users[strings.ToLower(handle)]
			if strings.TrimSpace(user.Handle) == "" {
				user.Handle = handle
			}

			return userpage.NewPageWithImages(r.api, r.theme, user, r.images)
		},
		"/posts/:id": func(params router.Params) reactea.Component {
			id := unescapeRoutePart(params.Get("id"))
			if post, ok := r.posts[id]; ok {
				return r.newPostPage(post)
			}

			return newPostRoute(r, id)
		},
	}
	if r.notifications != nil {
		routes[notificationsRoute] = router.Page(func() reactea.Component {
			return notificationpage.NewPageWithImages(r.notifications, r.theme, r.images)
		})
	}
	if r.messages != nil {
		routes[messagesRoute] = router.Page(func() reactea.Component {
			return messagepage.New(r.messages, r.theme)
		})
	}
	if r.bookmarks != nil {
		routes[bookmarksRoute] = router.Page(func() reactea.Component {
			return bookmarkpage.New(r.api, r.bookmarks, r.theme, r.images)
		})
	}

	return routes
}

func (r *root) newPostPage(post north.Post) *postpage.Screen {
	editor, _ := r.api.(postpage.Editor)

	return postpage.NewPageWithEditorAndImages(r.api, editor, r.theme, post, r.ownsPost(post), r.images)
}

func (r *root) currentPage() reactea.Component {
	if r.pages == nil {
		return nil
	}

	return r.pages.Current()
}
