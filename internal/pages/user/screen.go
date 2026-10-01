package user

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type API interface {
	User(context.Context, string) (north.User, *north.Response, error)
	UserPosts(context.Context, string, string) (north.PostPage, *north.Response, error)
}

type Screen struct {
	reactea.BasicComponent

	api    API
	theme  ui.Theme
	images *termimage.Renderer
	user   north.User
	offset int

	profileLoading bool
	profileLoaded  bool
	profileErr     error

	posts        []north.Post
	nextCursor   *string
	selected     int
	postsLoading bool
	loadingMore  bool
	postsErr     error
	fillLoads    int
}

func NewPage(api API, theme ui.Theme, user north.User) *Screen {
	return NewPageWithImages(api, theme, user, nil)
}

func NewPageWithImages(api API, theme ui.Theme, user north.User, images *termimage.Renderer) *Screen {
	loading := api != nil && strings.TrimSpace(user.Handle) != ""

	return &Screen{
		api:            api,
		theme:          theme,
		images:         images,
		user:           user,
		profileLoading: loading,
		profileLoaded:  !loading,
	}
}

func (d *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	if !d.profileLoading {
		return d.loadImages(ctx.Context(), ctx.Width(), d.user, nil)
	}

	return tea.Batch(
		d.loadProfile(ctx.Context()),
		d.loadPosts(ctx.Context(), false),
		d.loadImages(ctx.Context(), ctx.Width(), d.user, nil),
	)
}

var _ reactea.Component = (*Screen)(nil)
