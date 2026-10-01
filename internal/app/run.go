package app

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/domain/session"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type Options struct {
	StartupNotice string
	Images        bool
}

// Run starts the TUI.
func Run(ctx context.Context, api API, options Options) error {
	trends, err := unofficial.NewPublicClient(unofficial.WithUserAgent("nth"))
	if err != nil {
		return fmt.Errorf("configure trends: %w", err)
	}
	images := termimage.New(options.Images)
	root := newRootWithServices(api, trends, images)
	root.notice = options.StartupNotice
	stack := modal.New(root)
	program := reactea.New(
		&quitGuard{Wrapper: reactea.Wrap(stack)},
		reactea.WithAltScreen(),
		reactea.WithWindowTitle("nth — north for the terminal"),
	)
	defer program.Scope().Close()
	if source, ok := api.(interface {
		SetCookieRefreshHandler(func(*session.RefreshRequest))
	}); ok {
		source.SetCookieRefreshHandler(func(request *session.RefreshRequest) {
			program.Send(request)
		})
		defer source.SetCookieRefreshHandler(nil)
	}

	if err := program.Run(tea.WithContext(ctx)); err != nil {
		return fmt.Errorf("run terminal client: %w", err)
	}

	return nil
}

type quitGuard struct {
	reactea.Wrapper
}

func (g *quitGuard) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if reactea.Key(msg, "ctrl+c") {
		return tea.Quit
	}

	return g.Wrapper.Update(ctx, msg)
}
