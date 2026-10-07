package post

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
	conversationdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Screen struct {
	reactea.BasicComponent

	api              PostAPI
	conversation     conversationdomain.ConversationAPI
	activity         conversationdomain.PostActivityAPI
	polls            conversationdomain.PollAPI
	activityEnabled  bool
	historyEnabled   bool
	editor           Editor
	theme            ui.Theme
	images           *termimage.Renderer
	post             north.Post
	offset           int
	acting           map[postcomponent.Action]bool
	notice           string
	manage           bool
	editable         bool
	checkingEditable bool
	editableChecked  bool
	confirmingDelete bool
	linkedUser       int
	pollBusy         bool

	ancestors          []north.Post
	loadingAncestors   bool
	ancestorErr        error
	ancestorsTruncated bool

	replies             []north.Post
	nextReplyCursor     *string
	loadingConversation bool
	loadingMoreReplies  bool
	replyErr            error
	replyIndex          int
	replyFocused        bool
}

func NewPage(api PostAPI, theme ui.Theme, post north.Post, manage bool) *Screen {
	return NewPageWithEditorAndImages(api, nil, theme, post, manage, nil)
}

func NewPageWithEditor(api PostAPI, editor Editor, theme ui.Theme, post north.Post, manage bool) *Screen {
	return NewPageWithEditorAndImages(api, editor, theme, post, manage, nil)
}

func NewPageWithEditorAndImages(api PostAPI, editor Editor, theme ui.Theme, post north.Post, manage bool, images *termimage.Renderer) *Screen {
	conversation, _ := api.(conversationdomain.ConversationAPI)
	activity, _ := api.(conversationdomain.PostActivityAPI)
	activityEnabled := activity != nil
	if support, ok := api.(interface{ SupportsPostActivity() bool }); ok {
		activityEnabled = activity != nil && support.SupportsPostActivity()
	}
	historyEnabled := activity != nil
	if support, ok := api.(interface{ SupportsPostEditHistory() bool }); ok {
		historyEnabled = activity != nil && support.SupportsPostEditHistory()
	}
	polls, _ := api.(conversationdomain.PollAPI)
	if support, ok := api.(interface{ SupportsPolls() bool }); ok && !support.SupportsPolls() {
		polls = nil
	}
	page := &Screen{
		api: api, conversation: conversation, activity: activity, polls: polls,
		activityEnabled: activityEnabled, historyEnabled: historyEnabled,
		editor: editor, theme: theme, images: images, post: post,
		manage: manage, acting: make(map[postcomponent.Action]bool),
	}
	page.loadingAncestors = conversation == nil && api != nil && replyParentID(post) != ""
	page.checkingEditable = manage && editor != nil

	return page
}

func (d *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	commands := []tea.Cmd{d.loadImages(ctx.Context(), ctx.Width(), []north.Post{d.post})}
	if d.loadingAncestors {
		commands = append(commands, d.loadAncestors(ctx.Context()))
	}
	if d.conversation != nil && conversationPostID(d.post) != "" {
		commands = append(commands, d.loadConversation(ctx.Context(), false))
	}
	if d.checkingEditable {
		commands = append(commands, d.loadEditable(ctx.Context()))
	}

	return tea.Batch(commands...)
}

var _ reactea.Component = (*Screen)(nil)
