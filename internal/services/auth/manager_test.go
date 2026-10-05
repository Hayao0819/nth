package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestManagerSavesAndLoadsSetup(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	manager := &Manager{
		vault:    vault,
		browsers: &fakeBrowserSource{},
		getenv:   func(string) string { return "" },
	}
	want := Settings{Method: MethodBrowser, Browser: Profile{Browser: "firefox", Name: "default-release"}, Images: true}
	if err := manager.Save(want); err != nil {
		t.Fatal(err)
	}

	got, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != MethodBrowser || !got.Browser.Same(want.Browser) || !got.Complete() || !got.Images {
		t.Fatalf("loaded settings = %#v", got)
	}
}

func TestManagerSavesAndLoadsAPIToken(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	manager := &Manager{vault: vault, browsers: &fakeBrowserSource{}, getenv: func(string) string { return "" }}
	if err := manager.Save(Settings{Method: MethodAPIToken, Token: "api-token"}); err != nil {
		t.Fatal(err)
	}
	settings, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Method != MethodAPIToken || settings.Token != "api-token" || !settings.Complete() {
		t.Fatalf("loaded settings = %#v", settings)
	}
}

func TestManagerSavesAndLoadsOAuth(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	manager := &Manager{vault: vault, browsers: &fakeBrowserSource{}, getenv: func(string) string { return "" }}
	if err := manager.Save(Settings{Method: MethodAPIToken, Token: "saved-api-token"}); err != nil {
		t.Fatal(err)
	}
	want := OAuthCredential{
		ClientID: "client-id",
		Token: oauth2.Token{
			AccessToken:  "access-token",
			TokenType:    "Bearer",
			RefreshToken: "refresh-token",
			Expiry:       time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
		},
	}
	profile := Profile{Browser: "firefox", Name: "default-release"}
	if err := manager.Save(Settings{Method: MethodOAuth, OAuth: want, Browser: profile}); err != nil {
		t.Fatal(err)
	}
	settings, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Method != MethodOAuth || !settings.HasOAuth() || !settings.PreferOAuth() || settings.Token != "saved-api-token" || settings.OAuth.ClientID != want.ClientID || settings.OAuth.Token.RefreshToken != want.Token.RefreshToken || !settings.OAuth.Token.Expiry.Equal(want.Token.Expiry) || !settings.Browser.Same(profile) {
		t.Fatalf("loaded settings = %#v", settings)
	}

	rotated := &oauth2.Token{AccessToken: "new-access", RefreshToken: "new-refresh"}
	if err := manager.SaveOAuthToken(want.ClientID, rotated); err != nil {
		t.Fatal(err)
	}
	settings, err = manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.OAuth.Token.AccessToken != "new-access" || settings.OAuth.Token.RefreshToken != "new-refresh" {
		t.Fatalf("rotated OAuth token = %#v", settings.OAuth.Token)
	}

	manager.getenv = func(name string) string {
		if name == tokenEnvironment {
			return "environment-api-token"
		}

		return ""
	}
	settings, err = manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.PreferOAuth() || !settings.HasEnvironmentToken() || settings.Token != "environment-api-token" {
		t.Fatalf("environment override settings = %#v", settings)
	}
}

func TestManagerDoesNotCopyEnvironmentTokenIntoKeyring(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	manager := &Manager{vault: vault, browsers: &fakeBrowserSource{}, getenv: func(name string) string {
		if name == tokenEnvironment {
			return "environment-token"
		}

		return ""
	}}
	if err := manager.Save(Settings{Method: MethodAPIToken}); err != nil {
		t.Fatal(err)
	}
	if _, exists := vault.values[apiKeyEntry]; exists {
		t.Fatal("environment token was copied into the keyring")
	}
	if vault.values[methodEntry] != string(MethodAPIToken) {
		t.Fatalf("saved method = %q", vault.values[methodEntry])
	}
}

func TestManagerKeepsCredentialsWhenSwitchingMethods(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	manager := &Manager{vault: vault, browsers: &fakeBrowserSource{}, getenv: func(string) string { return "" }}
	profile := Profile{Browser: "firefox", Name: "default"}
	if err := manager.Save(Settings{Method: MethodBrowser, Browser: profile}); err != nil {
		t.Fatal(err)
	}
	if err := manager.Save(Settings{Method: MethodAPIToken, Token: "api-token"}); err != nil {
		t.Fatal(err)
	}
	settings, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Method != MethodAPIToken || settings.Token != "api-token" || !settings.Browser.Same(profile) {
		t.Fatalf("token settings = %#v", settings)
	}
	settings.Method = MethodBrowser
	if err := manager.Save(settings); err != nil {
		t.Fatal(err)
	}
	settings, err = manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Method != MethodBrowser || settings.Token != "api-token" || !settings.Browser.Same(profile) {
		t.Fatalf("browser settings = %#v", settings)
	}
}

func TestEnvironmentOverridesSavedAPIToken(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	vault.values[methodEntry] = string(MethodAPIToken)
	vault.values[apiKeyEntry] = "saved-token"
	manager := &Manager{vault: vault, browsers: &fakeBrowserSource{}, getenv: func(name string) string {
		if name == tokenEnvironment {
			return "environment-token"
		}

		return ""
	}}
	settings, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Token != "environment-token" || !settings.Complete() {
		t.Fatalf("loaded settings = %#v", settings)
	}
}

func TestEnvironmentTokenSurvivesKeyringFailure(t *testing.T) {
	t.Parallel()

	vaultErr := errors.New("keyring is locked")
	manager := &Manager{
		vault:    &memoryVault{values: make(map[string]string), getErr: vaultErr},
		browsers: &fakeBrowserSource{},
		getenv: func(name string) string {
			if name == tokenEnvironment {
				return "environment-token"
			}

			return ""
		},
	}
	settings, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Method != MethodAPIToken || settings.Token != "environment-token" {
		t.Fatalf("loaded settings = %#v", settings)
	}
}

func TestSelectedCredentialIgnoresAnUnrelatedKeyringFailure(t *testing.T) {
	t.Parallel()

	vault := &entryErrorVault{
		values: map[string]string{
			methodEntry: string(MethodAPIToken),
			apiKeyEntry: "saved-token",
		},
		errors: map[string]error{browserEntry: errors.New("browser entry is unreadable")},
	}
	manager := &Manager{vault: vault, browsers: &fakeBrowserSource{}, getenv: func(string) string { return "" }}
	settings, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Method != MethodAPIToken || settings.Token != "saved-token" {
		t.Fatalf("loaded settings = %#v", settings)
	}
}

func TestLoadMigratesLegacyCredentials(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	vault.values[apiKeyEntry] = "legacy-token"
	encoded, err := json.Marshal(Profile{Browser: "firefox"})
	if err != nil {
		t.Fatal(err)
	}
	vault.values[browserEntry] = string(encoded)
	manager := &Manager{vault: vault, browsers: &fakeBrowserSource{}, getenv: func(string) string { return "" }}
	settings, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if settings.Method != MethodAPIToken || !settings.Complete() {
		t.Fatalf("legacy settings = %#v", settings)
	}
}

func TestRefreshCookieCachesAndReusesTheSameProfile(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	browser := &fakeBrowserSource{header: "session=fresh"}
	manager := &Manager{vault: vault, browsers: browser}
	profile := Profile{Browser: "firefox", Name: "default-release"}

	if status := manager.RefreshCookie(context.Background(), profile); status.Source != CookieBrowser || status.Header != "session=fresh" || status.Err != nil {
		t.Fatalf("fresh status = %#v", status)
	}
	browser.header = ""
	browser.err = errors.New("browser is closed")
	if status := manager.RefreshCookie(context.Background(), profile); status.Source != CookieKeyring || status.Header != "session=fresh" || !errors.Is(status.Err, browser.err) {
		t.Fatalf("fallback status = %#v", status)
	}

	var cached savedCookie
	if err := json.Unmarshal([]byte(vault.values[cookieEntry]), &cached); err != nil {
		t.Fatal(err)
	}
	if cached.Header != "session=fresh" || !cached.Profile.Same(profile) {
		t.Fatalf("cached cookie = %#v", cached)
	}
}

func TestRefreshCookieDoesNotReuseAnotherProfilesCookie(t *testing.T) {
	t.Parallel()

	vault := newMemoryVault()
	browser := &fakeBrowserSource{header: "session=first"}
	manager := &Manager{vault: vault, browsers: browser}
	first := Profile{Browser: "firefox", Name: "first"}
	second := Profile{Browser: "firefox", Name: "second"}
	if status := manager.RefreshCookie(context.Background(), first); status.Source != CookieBrowser {
		t.Fatalf("fresh status = %#v", status)
	}

	browser.err = errors.New("cannot read profile")
	if status := manager.RefreshCookie(context.Background(), second); status.Source != CookieUnavailable {
		t.Fatalf("mismatched fallback status = %#v", status)
	}
}

func TestSaveValidatesSelectedMethod(t *testing.T) {
	t.Parallel()

	manager := &Manager{vault: newMemoryVault(), browsers: &fakeBrowserSource{}}
	err := manager.Save(Settings{})
	if err == nil || !strings.Contains(err.Error(), "method") {
		t.Fatalf("error = %v", err)
	}
	err = manager.Save(Settings{Method: MethodBrowser})
	if err == nil || !strings.Contains(err.Error(), "browser") {
		t.Fatalf("browser error = %v", err)
	}
	err = manager.Save(Settings{Method: MethodAPIToken, Token: "unsafe\ntoken"})
	if err == nil || !strings.Contains(err.Error(), "newline") {
		t.Fatalf("token error = %v", err)
	}
	err = manager.Save(Settings{Method: MethodOAuth})
	if err == nil || !strings.Contains(err.Error(), "OAuth") {
		t.Fatalf("OAuth error = %v", err)
	}
}

type memoryVault struct {
	values map[string]string
	getErr error
	setErr error
}

type entryErrorVault struct {
	values map[string]string
	errors map[string]error
}

func (v *entryErrorVault) Get(name string) (string, error) {
	if err := v.errors[name]; err != nil {
		return "", err
	}
	value, exists := v.values[name]
	if !exists {
		return "", errNotFound
	}

	return value, nil
}

func (v *entryErrorVault) Set(name, value string) error {
	v.values[name] = value

	return nil
}

func newMemoryVault() *memoryVault {
	return &memoryVault{values: make(map[string]string)}
}

func (v *memoryVault) Get(name string) (string, error) {
	if v.getErr != nil {
		return "", v.getErr
	}
	value, exists := v.values[name]
	if !exists {
		return "", errNotFound
	}

	return value, nil
}

func (v *memoryVault) Set(name, value string) error {
	if v.setErr != nil {
		return v.setErr
	}
	v.values[name] = value

	return nil
}

type fakeBrowserSource struct {
	profiles []Profile
	header   string
	err      error
}

func (s *fakeBrowserSource) Profiles(context.Context) []Profile {
	return append([]Profile(nil), s.profiles...)
}

func (s *fakeBrowserSource) CookieHeader(context.Context, Profile) (string, error) {
	return s.header, s.err
}
