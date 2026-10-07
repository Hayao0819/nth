package post

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Activity struct {
	reactea.BasicComponent

	api      domain.PostActivityAPI
	feedAPI  feed.API
	theme    ui.Theme
	images   *termimage.Renderer
	post     north.Post
	kind     navigation.PostActivity
	posts    []north.Post
	users    []north.User
	versions []north.PostVersion
	next     *string
	selected int
	top      int
	loading  bool
	more     bool
	err      error
}

func NewActivity(api domain.PostActivityAPI, feedAPI feed.API, theme ui.Theme, post north.Post, kind navigation.PostActivity, images *termimage.Renderer) *Activity {
	return &Activity{api: api, feedAPI: feedAPI, theme: theme, post: post, kind: kind, images: images}
}

func (s *Activity) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.load(ctx.Context(), false)
}

var _ reactea.Component = (*Activity)(nil)
