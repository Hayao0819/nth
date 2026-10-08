// Package cli defines the nth command line.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/Hayao0819/nth/internal/app"
	diagnosticfeature "github.com/Hayao0819/nth/internal/features/diagnostic"
	setupui "github.com/Hayao0819/nth/internal/features/setup"
	"github.com/Hayao0819/nth/internal/services/auth"
	"github.com/Hayao0819/nth/internal/services/northapi"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type credentialService interface {
	Load() (auth.Settings, error)
	Profiles(context.Context) []auth.Profile
	Save(auth.Settings) error
	Reset() (auth.ResetResult, error)
	StartOAuth(context.Context) (auth.OAuthSession, error)
	SaveOAuthToken(string, *oauth2.Token) error
	RefreshCookie(context.Context, auth.Profile) auth.CookieStatus
}

type dependencies struct {
	credentials credentialService
	setup       func(context.Context, auth.Settings, []auth.Profile, auth.OAuthStartFunc, func(auth.Settings) error) (bool, error)
	start       func(context.Context, app.API, app.Options) error
	diagnose    func(context.Context, diagnosticfeature.API) error
}

func defaultDependencies() dependencies {
	return dependencies{
		credentials: auth.NewManager(),
		setup:       setupui.Run,
		start:       app.Run,
		diagnose:    diagnosticfeature.Run,
	}
}

func newCommand(version string) *cobra.Command {
	return newCommandWith(version, defaultDependencies())
}

func newCommandWith(version string, deps dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:           "nth",
		Short:         "An unofficial north client for the terminal",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return run(command, version, deps)
		},
	}
	command.AddCommand(&cobra.Command{
		Use:           "setup",
		Short:         "Configure authentication",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, _, err := configure(command.Context(), deps, true)

			return err
		},
	})
	command.AddCommand(newResetCommand(deps))
	command.AddCommand(newTestCommand(version, deps))

	return command
}

// Execute runs nth and returns its exit status.
func Execute(version string) int {
	if err := newCommand(version).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "nth: %v\n", err)

		return 1
	}

	return 0
}

func run(command *cobra.Command, version string, deps dependencies) error {
	settings, proceed, err := configure(command.Context(), deps, false)
	if err != nil || !proceed {
		return err
	}
	session, err := openSession(command.Context(), version, deps.credentials, settings)
	if err != nil {
		return err
	}
	if err := deps.start(command.Context(), session.api, app.Options{StartupNotice: session.notice, Images: settings.Images}); err != nil {
		return err
	}

	return nil
}

type clientSession struct {
	api    app.API
	notice string
}

func openSession(
	ctx context.Context,
	version string,
	credentials credentialService,
	settings auth.Settings,
) (clientSession, error) {
	var (
		official northapi.OfficialAPI
		web      *northapi.Client
		notice   string
		err      error
	)
	officialOptions := []north.Option{
		north.WithUserAgent("nth/" + version),
		north.WithHTTPClient(northapi.NewOfficialHTTPClient()),
	}
	if settings.HasAPIToken() && !settings.PreferOAuth() {
		official, err = north.NewClient(settings.Token, officialOptions...)
		if err != nil {
			return clientSession{}, err
		}
	} else if settings.HasOAuth() {
		config := auth.OAuthConfig(settings.OAuth.ClientID)
		token := settings.OAuth.Token
		official, err = north.NewClientWithOAuth(
			ctx,
			config,
			&token,
			func(token *oauth2.Token) error {
				return credentials.SaveOAuthToken(settings.OAuth.ClientID, token)
			},
			officialOptions...,
		)
		if err != nil {
			return clientSession{}, err
		}
	}
	if settings.HasBrowser() {
		status := credentials.RefreshCookie(ctx, settings.Browser)
		if status.Header != "" {
			web, err = northapi.New(status.Header, unofficial.WithUserAgent("nth/"+version))
			if err != nil {
				return clientSession{}, err
			}
		}
		notice = cookieNotice(status)
		refresh := func(ctx context.Context) (*northapi.Client, error) {
			return refreshedWebClient(ctx, credentials, settings.Browser, version)
		}

		return clientSession{
			api:    northapi.NewHybrid(official, web, settings.Browser.Label(), refresh),
			notice: notice,
		}, nil
	}
	if official == nil {
		return clientSession{}, setupError("authentication is not configured")
	}

	return clientSession{api: northapi.NewHybrid(official, nil, "", nil)}, nil
}

func refreshedWebClient(
	ctx context.Context,
	credentials credentialService,
	profile auth.Profile,
	version string,
) (*northapi.Client, error) {
	status := credentials.RefreshCookie(ctx, profile)
	if status.Header == "" {
		if status.Err != nil {
			return nil, fmt.Errorf("read browser cookies: %w", status.Err)
		}

		return nil, errors.New("no north.rip cookies were found in the selected browser")
	}
	client, err := northapi.New(status.Header, unofficial.WithUserAgent("nth/"+version))
	if err != nil {
		return nil, err
	}
	if _, _, err := client.Me(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func configure(ctx context.Context, deps dependencies, force bool) (auth.Settings, bool, error) {
	settings, err := deps.credentials.Load()
	if err != nil {
		return settings, false, err
	}
	if !force && settings.Complete() && !settings.OAuthNeedsAuthorization() {
		return settings, true, nil
	}

	complete, err := deps.setup(
		ctx,
		settings,
		deps.credentials.Profiles(ctx),
		deps.credentials.StartOAuth,
		deps.credentials.Save,
	)
	if err != nil || !complete {
		return settings, false, err
	}

	settings, err = deps.credentials.Load()
	if err != nil {
		return settings, false, err
	}
	if !settings.Complete() {
		return settings, false, setupError("credentials were not saved")
	}
	if settings.OAuthNeedsAuthorization() {
		return settings, false, setupError("OAuth needs additional authorization")
	}

	return settings, true, nil
}

func cookieNotice(status auth.CookieStatus) string {
	switch {
	case status.Source == auth.CookieUnavailable:
		return "Browser session needs to be refreshed"
	case status.Source == auth.CookieKeyring:
		return "Browser cookies could not be refreshed; using the copy saved in your keyring"
	case status.Source == auth.CookieBrowser && status.Err != nil:
		return "Browser session loaded, but it could not be cached in your keyring"
	default:
		return ""
	}
}

func setupError(message string) error {
	return fmt.Errorf("%s; run nth setup", message)
}
