// Package auth manages the credentials used by nth.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	tokenEnvironment = "NORTH_API_KEY"

	methodEntry  = "north-auth-method"
	apiKeyEntry  = "north-api-key"
	browserEntry = "north-browser"
	cookieEntry  = "north-cookie"
	imagesEntry  = "north-terminal-images"
)

// Method identifies the credential most recently configured by setup.
type Method string

const (
	MethodBrowser  Method = "browser"
	MethodAPIToken Method = "api-token"
)

func (m Method) Valid() bool {
	return m == MethodBrowser || m == MethodAPIToken
}

// Settings contains the values collected by initial setup.
type Settings struct {
	Method  Method
	Token   string
	Browser Profile
	Images  bool
}

func (s Settings) HasAPIToken() bool {
	return strings.TrimSpace(s.Token) != ""
}

func (s Settings) HasBrowser() bool {
	return s.Browser.Valid()
}

// Complete reports whether at least one authentication method is configured.
func (s Settings) Complete() bool {
	return s.HasAPIToken() || s.HasBrowser()
}

// CookieSource identifies where the current browser session came from.
type CookieSource uint8

const (
	CookieUnavailable CookieSource = iota
	CookieBrowser
	CookieKeyring
)

// CookieStatus reports the result of refreshing the cached browser session.
type CookieStatus struct {
	Source CookieSource
	Header string
	Err    error
}

// Manager reads credentials from the system keyring and local browsers.
type Manager struct {
	vault    vault
	browsers browserSource
	getenv   func(string) string
}

// NewManager uses the current user's system keyring and browser profiles.
func NewManager() *Manager {
	return &Manager{vault: systemKeyring{}, browsers: localBrowsers{}, getenv: os.Getenv}
}

// Load returns every available credential and the last setup selection.
func (m *Manager) Load() (Settings, error) {
	settings := Settings{Images: m.loadImages()}
	method, err := m.vault.Get(methodEntry)
	switch {
	case err == nil:
		settings.Method = Method(method)
	case errors.Is(err, errNotFound):
	default:
		if token := m.environmentToken(); token != "" {
			return Settings{Method: MethodAPIToken, Token: token, Images: settings.Images}, nil
		}
		return settings, fmt.Errorf("read authentication method from keyring: %w", err)
	}
	err = nil

	switch settings.Method {
	case MethodAPIToken:
		settings.Token, _, err = m.loadToken()
		settings.Browser, _ = m.loadBrowser()
	case MethodBrowser:
		settings.Browser, err = m.loadBrowser()
		settings.Token, _, _ = m.loadToken()
	default:
		var tokenErr, browserErr error
		var tokenAvailable bool
		settings.Token, tokenAvailable, tokenErr = m.loadToken()
		settings.Browser, browserErr = m.loadBrowser()
		switch {
		case tokenAvailable:
			settings.Method = MethodAPIToken
		case settings.Browser.Valid():
			settings.Method = MethodBrowser
		case tokenErr != nil || browserErr != nil:
			err = errors.Join(tokenErr, browserErr)
		}
	}
	if err != nil {
		return settings, err
	}

	return settings, nil
}

func (m *Manager) loadToken() (string, bool, error) {
	if token := m.environmentToken(); token != "" {
		return token, true, nil
	}
	token, err := m.vault.Get(apiKeyEntry)
	saved := false
	switch {
	case err == nil:
		token = strings.TrimSpace(token)
		saved = token != ""
	case errors.Is(err, errNotFound):
		token = ""
	default:
		return "", false, fmt.Errorf("read API token from keyring: %w", err)
	}
	return token, saved, nil
}

func (m *Manager) environmentToken() string {
	if m.getenv == nil {
		return ""
	}

	return strings.TrimSpace(m.getenv(tokenEnvironment))
}

func (m *Manager) loadBrowser() (Profile, error) {
	encoded, err := m.vault.Get(browserEntry)
	switch {
	case err == nil:
		var profile Profile
		if err := json.Unmarshal([]byte(encoded), &profile); err == nil {
			return profile, nil
		}

		return Profile{}, nil
	case errors.Is(err, errNotFound):
		return Profile{}, nil
	default:
		return Profile{}, fmt.Errorf("read browser from keyring: %w", err)
	}
}

func (m *Manager) loadImages() bool {
	value, err := m.vault.Get(imagesEntry)
	if err != nil {
		return false
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(value))

	return err == nil && enabled
}

// Profiles returns the browser profiles available to setup.
func (m *Manager) Profiles(ctx context.Context) []Profile {
	return m.browsers.Profiles(ctx)
}

// Save stores the selected credential without removing the other one.
func (m *Manager) Save(settings Settings) error {
	switch settings.Method {
	case MethodBrowser:
		if !settings.Browser.Valid() {
			return errors.New("choose a browser")
		}
		encoded, err := json.Marshal(settings.Browser)
		if err != nil {
			return fmt.Errorf("encode browser choice: %w", err)
		}
		if err := m.vault.Set(browserEntry, string(encoded)); err != nil {
			return fmt.Errorf("save browser in keyring: %w", err)
		}
	case MethodAPIToken:
		token := strings.TrimSpace(settings.Token)
		if token == "" {
			return errors.New("enter a north API token")
		}
		if strings.ContainsAny(token, "\r\n") {
			return errors.New("the north API token contains a newline")
		}
		if err := m.vault.Set(apiKeyEntry, token); err != nil {
			return fmt.Errorf("save API token in keyring: %w", err)
		}
	default:
		return errors.New("choose an authentication method")
	}
	if err := m.vault.Set(methodEntry, string(settings.Method)); err != nil {
		return fmt.Errorf("save authentication method in keyring: %w", err)
	}
	if err := m.vault.Set(imagesEntry, strconv.FormatBool(settings.Images)); err != nil {
		return fmt.Errorf("save terminal image preference in keyring: %w", err)
	}

	return nil
}

// RefreshCookie reloads north.rip cookies from the selected browser. If that
// fails, it checks the last cookie saved for the same browser profile.
func (m *Manager) RefreshCookie(ctx context.Context, profile Profile) CookieStatus {
	header, browserErr := m.browsers.CookieHeader(ctx, profile)
	if browserErr == nil && !validCookieHeader(header) {
		browserErr = errNorthCookiesNotFound
	}
	if browserErr == nil {
		cached := savedCookie{Profile: profile, Header: header}
		encoded, err := json.Marshal(cached)
		if err == nil {
			err = m.vault.Set(cookieEntry, string(encoded))
		}
		if err != nil {
			return CookieStatus{Source: CookieBrowser, Header: header, Err: fmt.Errorf("save browser cookies in keyring: %w", err)}
		}

		return CookieStatus{Source: CookieBrowser, Header: header}
	}

	encoded, cacheErr := m.vault.Get(cookieEntry)
	if cacheErr != nil {
		if errors.Is(cacheErr, errNotFound) {
			return CookieStatus{Source: CookieUnavailable, Err: browserErr}
		}

		return CookieStatus{Source: CookieUnavailable, Err: errors.Join(browserErr, fmt.Errorf("read saved cookies from keyring: %w", cacheErr))}
	}

	var cached savedCookie
	if err := json.Unmarshal([]byte(encoded), &cached); err != nil {
		return CookieStatus{Source: CookieUnavailable, Err: errors.Join(browserErr, fmt.Errorf("decode saved cookies: %w", err))}
	}
	if !cached.Profile.Same(profile) || !validCookieHeader(cached.Header) {
		return CookieStatus{Source: CookieUnavailable, Err: browserErr}
	}

	return CookieStatus{Source: CookieKeyring, Header: cached.Header, Err: browserErr}
}

type savedCookie struct {
	Profile Profile `json:"profile"`
	Header  string  `json:"cookie"`
}

func validCookieHeader(header string) bool {
	if strings.TrimSpace(header) == "" {
		return false
	}
	for _, char := range header {
		if char < ' ' || char == 0x7f {
			return false
		}
	}

	return true
}
