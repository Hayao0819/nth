package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Hayao0819/nth/internal/app"
	"github.com/Hayao0819/nth/internal/domain/bookmark"
	"github.com/Hayao0819/nth/internal/domain/message"
	"github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/services/auth"
	"golang.org/x/oauth2"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	command := newCommand("1.2.3")
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{"--version"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); !strings.Contains(got, "1.2.3") {
		t.Errorf("version output = %q", got)
	}
}

func TestFirstRunSavesSetupBeforeStarting(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{
		profiles: []auth.Profile{{Browser: "firefox", Name: "default"}},
		status:   auth.CookieStatus{Source: auth.CookieBrowser, Header: "session=secret"},
	}
	setupCalled := false
	started := false
	deps := dependencies{
		credentials: credentials,
		setup: func(_ context.Context, initial auth.Settings, profiles []auth.Profile, _ auth.OAuthStartFunc, save func(auth.Settings) error) (bool, error) {
			setupCalled = true
			if initial.Complete() || len(profiles) != 1 {
				t.Fatalf("setup initial = %#v, profiles = %#v", initial, profiles)
			}

			return true, save(auth.Settings{Method: auth.MethodBrowser, Browser: profiles[0]})
		},
		start: func(_ context.Context, _ app.API, options app.Options) error {
			started = true
			if options.StartupNotice != "" {
				t.Errorf("startup notice = %q", options.StartupNotice)
			}

			return nil
		},
	}

	command := newCommandWith("test", deps)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !setupCalled || !started || credentials.refreshes != 1 {
		t.Fatalf("setup = %v, started = %v, refreshes = %d", setupCalled, started, credentials.refreshes)
	}
}

func TestSavedCookieFallbackIsReported(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{
		settings: auth.Settings{
			Method:  auth.MethodBrowser,
			Browser: auth.Profile{Browser: "firefox", Name: "default"},
		},
		status: auth.CookieStatus{Source: auth.CookieKeyring, Header: "session=cached", Err: errors.New("browser unavailable")},
	}
	setupCalled := false
	var notice string
	deps := dependencies{
		credentials: credentials,
		setup: func(context.Context, auth.Settings, []auth.Profile, auth.OAuthStartFunc, func(auth.Settings) error) (bool, error) {
			setupCalled = true

			return false, nil
		},
		start: func(_ context.Context, _ app.API, options app.Options) error {
			notice = options.StartupNotice

			return nil
		},
	}

	command := newCommandWith("test", deps)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if setupCalled {
		t.Fatal("setup ran despite complete credentials")
	}
	if !strings.Contains(notice, "using the copy") {
		t.Fatalf("startup notice = %q", notice)
	}
}

func TestSetupCommandDoesNotStartTheClient(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{settings: auth.Settings{
		Method:  auth.MethodBrowser,
		Browser: auth.Profile{Browser: "chrome"},
	}}
	started := false
	deps := dependencies{
		credentials: credentials,
		setup: func(_ context.Context, _ auth.Settings, _ []auth.Profile, _ auth.OAuthStartFunc, save func(auth.Settings) error) (bool, error) {
			selected := auth.Settings{Method: auth.MethodBrowser, Browser: auth.Profile{Browser: "firefox"}}

			return true, save(selected)
		},
		start: func(context.Context, app.API, app.Options) error {
			started = true

			return nil
		},
	}

	command := newCommandWith("test", deps)
	command.SetArgs([]string{"setup"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if started || credentials.refreshes != 0 {
		t.Fatalf("started = %v, refreshes = %d", started, credentials.refreshes)
	}
	if credentials.settings.Browser.Browser != "firefox" {
		t.Fatalf("settings = %#v", credentials.settings)
	}
}

func TestAPITokenDoesNotReadBrowserCookies(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{settings: auth.Settings{Method: auth.MethodAPIToken, Token: "api-token", Images: true}}
	started := false
	images := false
	deps := dependencies{
		credentials: credentials,
		setup: func(context.Context, auth.Settings, []auth.Profile, auth.OAuthStartFunc, func(auth.Settings) error) (bool, error) {
			t.Fatal("setup ran despite complete token settings")

			return false, nil
		},
		start: func(_ context.Context, api app.API, options app.Options) error {
			started = true
			images = options.Images
			if _, ok := api.(notification.API); !ok {
				t.Fatal("API token client does not expose notifications")
			}
			if _, ok := api.(bookmark.API); !ok {
				t.Fatal("API token client does not expose bookmarks")
			}
			if _, ok := api.(message.API); !ok {
				t.Fatal("API token client does not expose messages")
			}

			return nil
		},
	}
	command := newCommandWith("test", deps)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !started || !images || credentials.refreshes != 0 {
		t.Fatalf("started = %v, images = %v, browser refreshes = %d", started, images, credentials.refreshes)
	}
}

func TestOAuthStartsTheOfficialClient(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{settings: auth.Settings{
		Method: auth.MethodOAuth,
		Token:  "unsafe\napi-token",
		OAuth: auth.OAuthCredential{
			ClientID: "client-id",
			Token:    oauth2.Token{AccessToken: "access-token", RefreshToken: "refresh-token"},
		},
	}}
	started := false
	deps := dependencies{
		credentials: credentials,
		setup: func(context.Context, auth.Settings, []auth.Profile, auth.OAuthStartFunc, func(auth.Settings) error) (bool, error) {
			t.Fatal("setup ran despite complete OAuth settings")

			return false, nil
		},
		start: func(_ context.Context, api app.API, _ app.Options) error {
			started = true
			if _, ok := api.(notification.API); !ok {
				t.Fatal("OAuth client does not expose notifications")
			}
			if _, ok := api.(bookmark.API); !ok {
				t.Fatal("OAuth client does not expose bookmarks")
			}
			if _, ok := api.(message.API); !ok {
				t.Fatal("OAuth client does not expose messages")
			}

			return nil
		},
	}
	command := newCommandWith("test", deps)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !started || credentials.refreshes != 0 {
		t.Fatalf("started = %v, browser refreshes = %d", started, credentials.refreshes)
	}
}

func TestAPITokenAndBrowserSessionAreCombined(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{
		settings: auth.Settings{
			Method:  auth.MethodBrowser,
			Token:   "api-token",
			Browser: auth.Profile{Browser: "firefox", Name: "default"},
		},
		status: auth.CookieStatus{Source: auth.CookieBrowser, Header: "session=secret"},
	}
	started := false
	deps := dependencies{
		credentials: credentials,
		setup: func(context.Context, auth.Settings, []auth.Profile, auth.OAuthStartFunc, func(auth.Settings) error) (bool, error) {
			t.Fatal("setup ran despite complete settings")

			return false, nil
		},
		start: func(_ context.Context, api app.API, _ app.Options) error {
			started = true
			if _, ok := api.(notification.API); !ok {
				t.Fatal("combined client does not expose notifications")
			}

			return nil
		},
	}
	command := newCommandWith("test", deps)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !started || credentials.refreshes != 1 {
		t.Fatalf("started = %v, browser refreshes = %d", started, credentials.refreshes)
	}
}

func TestMissingBrowserSessionStartsRefreshableClient(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{
		settings: auth.Settings{Method: auth.MethodBrowser, Browser: auth.Profile{Browser: "firefox"}},
		status:   auth.CookieStatus{Source: auth.CookieUnavailable, Err: errors.New("cookies not found")},
	}
	deps := dependencies{
		credentials: credentials,
		setup: func(context.Context, auth.Settings, []auth.Profile, auth.OAuthStartFunc, func(auth.Settings) error) (bool, error) {
			t.Fatal("setup ran despite a saved browser")

			return false, nil
		},
		start: func(_ context.Context, api app.API, options app.Options) error {
			if _, ok := api.(notification.API); !ok {
				t.Fatal("browser features disappeared while waiting for a new session")
			}
			if !strings.Contains(options.StartupNotice, "needs to be refreshed") {
				t.Fatalf("startup notice = %q", options.StartupNotice)
			}

			return nil
		},
	}
	command := newCommandWith("test", deps)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
}

type fakeCredentials struct {
	settings  auth.Settings
	profiles  []auth.Profile
	status    auth.CookieStatus
	refreshes int
}

func (f *fakeCredentials) Load() (auth.Settings, error) {
	return f.settings, nil
}

func (f *fakeCredentials) Profiles(context.Context) []auth.Profile {
	return append([]auth.Profile(nil), f.profiles...)
}

func (f *fakeCredentials) Save(settings auth.Settings) error {
	f.settings = settings

	return nil
}

func (f *fakeCredentials) StartOAuth(context.Context) (auth.OAuthSession, error) {
	return nil, errors.New("OAuth is not configured in this test")
}

func (f *fakeCredentials) SaveOAuthToken(clientID string, token *oauth2.Token) error {
	if token == nil {
		return errors.New("OAuth token is nil")
	}
	f.settings.OAuth = auth.OAuthCredential{ClientID: clientID, Token: *token}

	return nil
}

func (f *fakeCredentials) RefreshCookie(context.Context, auth.Profile) auth.CookieStatus {
	f.refreshes++

	return f.status
}
