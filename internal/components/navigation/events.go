package navigation

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
)

type OpenPostMsg struct{ Post north.Post }

type OpenUserMsg struct{ User north.User }

func OpenPost(post north.Post) tea.Cmd {
	return func() tea.Msg { return OpenPostMsg{Post: post} }
}

func OpenUser(user north.User) tea.Cmd {
	return func() tea.Msg { return OpenUserMsg{User: user} }
}
