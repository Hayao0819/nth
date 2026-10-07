package message

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	messagedomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Screen struct {
	reactea.BasicComponent

	api       messagedomain.MessageAPI
	theme     ui.Theme
	write     messagedomain.MessageWriterAPI
	request   messagedomain.MessageRequestAPI
	group     messagedomain.MessageGroupAPI
	recipient messagedomain.MessageRecipientAPI
	viewer    north.User
	requests  bool

	conversations      []messagedomain.MessageConversation
	conversationCursor *string
	requestCount       int
	conversationIndex  int
	conversationTop    int

	current       *messagedomain.MessageConversation
	messages      []messagedomain.Message
	messageCursor *string
	messageIndex  int
	messageTop    int

	loading     bool
	loadingMore bool
	acting      bool
	err         error
	notice      string
	prompt      promptKind
	promptID    string
	confirm     confirmKind
	recipients  recipientPurpose
}

func New(api messagedomain.MessageAPI, theme ui.Theme) *Screen {
	screen := &Screen{api: api, theme: theme}
	if supportsDMWriting(api) {
		screen.write, _ = api.(messagedomain.MessageWriterAPI)
	}
	if supportsDMRequests(api) {
		screen.request, _ = api.(messagedomain.MessageRequestAPI)
	}
	if supportsDMGroups(api) {
		screen.group, _ = api.(messagedomain.MessageGroupAPI)
	}
	if supportsDMRecipients(api) {
		screen.recipient, _ = api.(messagedomain.MessageRecipientAPI)
	}

	return screen
}

func (s *Screen) SetViewer(user north.User) {
	s.viewer = user
}

func supportsDMWriting(api any) bool {
	reporter, ok := api.(interface{ SupportsDMWriting() bool })

	return !ok || reporter.SupportsDMWriting()
}

func supportsDMRequests(api any) bool {
	reporter, ok := api.(interface{ SupportsDMRequests() bool })

	return !ok || reporter.SupportsDMRequests()
}

func supportsDMGroups(api any) bool {
	reporter, ok := api.(interface{ SupportsDMGroups() bool })

	return !ok || reporter.SupportsDMGroups()
}

func supportsDMRecipients(api any) bool {
	reporter, ok := api.(interface{ SupportsDMRecipients() bool })

	return !ok || reporter.SupportsDMRecipients()
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.loadConversations(ctx.Context(), false)
}

var _ reactea.Component = (*Screen)(nil)
