package message

import (
	tea "charm.land/bubbletea/v2"
	messagedomain "github.com/Hayao0819/nth/internal/domain/message"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Screen struct {
	reactea.BasicComponent

	api   messagedomain.API
	theme ui.Theme

	conversations      []messagedomain.Conversation
	conversationCursor *string
	requestCount       int
	conversationIndex  int
	conversationTop    int

	current       *messagedomain.Conversation
	messages      []messagedomain.Message
	messageCursor *string
	messageIndex  int
	messageTop    int

	loading     bool
	loadingMore bool
	err         error
	notice      string
}

func New(api messagedomain.API, theme ui.Theme) *Screen {
	return &Screen{api: api, theme: theme}
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	return s.loadConversations(ctx.Context(), false)
}

var _ reactea.Component = (*Screen)(nil)
