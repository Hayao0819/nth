package user

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type API interface {
	User(context.Context, string) (north.User, *north.Response, error)
	UserPosts(context.Context, string, string) (north.PostPage, *north.Response, error)
}

type Screen struct {
	reactea.BasicComponent

	api           API
	activity      domain.UserActivityAPI
	profile       domain.ProfileAPI
	profileMedia  bool
	connections   domain.UserConnectionsAPI
	accountSafety bool
	listMembers   domain.ListMemberAPI
	relationships relationshipAPI
	postNotices   postNotificationAPI
	theme         ui.Theme
	images        *termimage.Renderer
	user          north.User
	viewer        *north.User
	offset        int

	profileLoading bool
	profileLoaded  bool
	profileErr     error

	posts        []north.Post
	tab          profileTab
	nextCursor   *string
	selected     int
	postsLoading bool
	loadingMore  bool
	postsErr     error
	fillLoads    int

	relationshipBusy   relationshipKind
	relationshipErr    error
	relationshipNotice string
	followRequested    bool
	postNotifications  bool
	confirmingBlock    bool
	profileBusy        bool
	profileNotice      string
	profileUpdateErr   error
}

func NewPage(api API, theme ui.Theme, user north.User) *Screen {
	return NewPageWithImages(api, theme, user, nil)
}

func NewPageWithImages(api API, theme ui.Theme, user north.User, images *termimage.Renderer) *Screen {
	loading := api != nil && strings.TrimSpace(user.Handle) != ""
	relationships, _ := api.(relationshipAPI)
	if support, ok := api.(interface{ SupportsRelationships() bool }); ok && !support.SupportsRelationships() {
		relationships = nil
	}
	postNotices, _ := api.(postNotificationAPI)
	if support, ok := api.(interface{ SupportsPostNotifications() bool }); ok && !support.SupportsPostNotifications() {
		postNotices = nil
	}
	activity, _ := api.(domain.UserActivityAPI)
	if support, ok := api.(interface{ SupportsUserActivity() bool }); ok && !support.SupportsUserActivity() {
		activity = nil
	}
	profile, _ := api.(domain.ProfileAPI)
	if support, ok := api.(interface{ SupportsProfileUpdate() bool }); ok && !support.SupportsProfileUpdate() {
		profile = nil
	}
	_, profileMedia := api.(domain.MediaAPI)
	profileMedia = profileMedia && profile != nil
	if support, ok := api.(interface{ SupportsProfileMediaUpdate() bool }); ok {
		profileMedia = support.SupportsProfileMediaUpdate() && profile != nil
	}
	connections, _ := api.(domain.UserConnectionsAPI)
	if support, ok := api.(interface{ SupportsUserConnections() bool }); ok && !support.SupportsUserConnections() {
		connections = nil
	}
	_, accountSafety := api.(domain.AccountSafetyAPI)
	if support, ok := api.(interface{ SupportsAccountSafety() bool }); ok && !support.SupportsAccountSafety() {
		accountSafety = false
	}
	listMembers, _ := api.(domain.ListMemberAPI)
	if support, ok := api.(interface{ SupportsListMembers() bool }); ok && !support.SupportsListMembers() {
		listMembers = nil
	}

	return &Screen{
		api:            api,
		activity:       activity,
		profile:        profile,
		profileMedia:   profileMedia,
		connections:    connections,
		accountSafety:  accountSafety,
		listMembers:    listMembers,
		relationships:  relationships,
		postNotices:    postNotices,
		theme:          theme,
		images:         images,
		user:           user,
		profileLoading: loading,
		profileLoaded:  !loading,
	}
}

func (d *Screen) SetViewer(user north.User) {
	d.viewer = &user
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
