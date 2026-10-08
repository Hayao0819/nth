package app

import (
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/reactea/v2"
)

func (r *root) openBrowser(rawURL string) tea.Cmd {
	return func() tea.Msg {
		if r.openURL == nil {
			return navigation.BrowserOpenedMsg{URL: rawURL, Err: errors.New("browser launcher is unavailable")}
		}

		return navigation.BrowserOpenedMsg{URL: rawURL, Err: r.openURL(rawURL)}
	}
}

func (r *root) handleBrowserOpened(ctx *reactea.Ctx, result navigation.BrowserOpenedMsg) tea.Cmd {
	if result.Err != nil {
		r.problem = fmt.Errorf("open browser: %w", result.Err)
		r.notice = ""
	} else {
		r.problem = nil
		r.notice = "Opened in browser"
	}

	return r.Wrapper.Update(ctx, result)
}
