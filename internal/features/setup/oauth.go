package setup

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/services/auth"
)

type oauthStartedMsg struct {
	attempt uint64
	session auth.OAuthSession
	err     error
}

type oauthCompletedMsg struct {
	attempt    uint64
	credential auth.OAuthCredential
	err        error
}

func (w *wizard) beginOAuth(ctx context.Context) tea.Cmd {
	w.stopOAuth()
	w.problem = ""
	w.prompt = auth.OAuthPrompt{}
	if w.startOAuth == nil {
		w.problem = "OAuth sign-in is not available"

		return nil
	}
	w.attempt++
	attempt := w.attempt
	w.oauthBusy = true
	w.oauthCtx, w.cancel = context.WithCancel(ctx)

	return func() tea.Msg {
		session, err := w.startOAuth(w.oauthCtx)

		return oauthStartedMsg{attempt: attempt, session: session, err: err}
	}
}

func (w *wizard) waitOAuth(attempt uint64, session auth.OAuthSession) tea.Cmd {
	return func() tea.Msg {
		credential, err := session.Wait(w.oauthCtx)

		return oauthCompletedMsg{attempt: attempt, credential: credential, err: err}
	}
}

func (w *wizard) cancelOAuth() {
	w.stopOAuth()
	w.attempt++
}

func (w *wizard) stopOAuth() {
	if w.cancel != nil {
		w.cancel()
	}
	w.cancel = nil
	w.oauthCtx = nil
	w.oauthBusy = false
}
