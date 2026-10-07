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

	"golang.org/x/oauth2"
)

const (
	tokenEnvironment    = "NORTH_API_KEY"
	clientIDEnvironment = "NORTH_CLIENT_ID"

	methodEntry  = "north-auth-method"
	apiKeyEntry  = "north-api-key"
	oauthEntry   = "north-oauth"
	browserEntry = "north-browser"
	cookieEntry  = "north-cookie"
	imagesEntry  = "north-terminal-images"
)

// Manager reads credentials from the system keyring and local browsers.
type Manager struct {
	vault       vault
	browsers    browserSource
	getenv      func(string) string
	oauthConfig func(string) *oauth2.Config
}

// NewManager uses the current user's system keyring and browser profiles.
func NewManager() *Manager {
	return &Manager{
		vault:       systemKeyring{},
		browsers:    localBrowsers{},
		getenv:      os.Getenv,
		oauthConfig: newOAuthConfig,
	}
}

// Load returns every available credential and the last setup selection.
func (m *Manager) Load() (Settings, error) {
	settings := Settings{Images: m.loadImages()}
	method, methodErr := m.vault.Get(methodEntry)
	switch {
	case methodErr == nil:
		settings.Method = Method(method)
	case errors.Is(methodErr, errNotFound):
		methodErr = nil
	default:
		methodErr = fmt.Errorf("read authentication method from keyring: %w", methodErr)
	}

	var tokenAvailable bool
	var tokenErr, oauthErr, browserErr error
	settings.Token, tokenAvailable, tokenErr = m.loadToken()
	settings.apiTokenFromEnvironment = m.environmentToken() != ""
	settings.OAuth, _, oauthErr = m.loadOAuth()
	settings.Browser, browserErr = m.loadBrowser()

	selectedAvailable := map[Method]bool{
		MethodAPIToken: tokenAvailable,
		MethodOAuth:    settings.HasOAuth(),
		MethodBrowser:  settings.HasBrowser(),
	}[settings.Method]
	if !selectedAvailable {
		switch {
		case tokenAvailable:
			settings.Method = MethodAPIToken
		case settings.HasOAuth():
			settings.Method = MethodOAuth
		case settings.HasBrowser():
			settings.Method = MethodBrowser
		}
	}
	if settings.Complete() {
		return settings, nil
	}
	if err := errors.Join(methodErr, tokenErr, oauthErr, browserErr); err != nil {
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
	case MethodAPIToken:
		token := strings.TrimSpace(settings.Token)
		if token == "" && m.environmentToken() == "" {
			return errors.New("enter a north API token")
		}
		if strings.ContainsAny(token, "\r\n") {
			return errors.New("the north API token contains a newline")
		}
		if token != "" {
			if err := m.vault.Set(apiKeyEntry, token); err != nil {
				return fmt.Errorf("save API token in keyring: %w", err)
			}
		}
	case MethodOAuth:
		if !settings.OAuth.Valid() {
			return errors.New("sign in to north with OAuth")
		}
		if settings.OAuth.NeedsAuthorization() {
			return errors.New("authorize the additional north permissions")
		}
		if err := m.saveOAuth(settings.OAuth); err != nil {
			return err
		}
	default:
		return errors.New("choose an authentication method")
	}
	if settings.Browser.Valid() {
		encoded, err := json.Marshal(settings.Browser)
		if err != nil {
			return fmt.Errorf("encode browser choice: %w", err)
		}
		if err := m.vault.Set(browserEntry, string(encoded)); err != nil {
			return fmt.Errorf("save browser in keyring: %w", err)
		}
	}
	if err := m.vault.Set(methodEntry, string(settings.Method)); err != nil {
		return fmt.Errorf("save authentication method in keyring: %w", err)
	}
	if err := m.vault.Set(imagesEntry, strconv.FormatBool(settings.Images)); err != nil {
		return fmt.Errorf("save terminal image preference in keyring: %w", err)
	}

	return nil
}

// Reset deletes every nth entry from the system keyring.
func (m *Manager) Reset() (ResetResult, error) {
	if err := m.vault.DeleteAll(); err != nil {
		return ResetResult{}, fmt.Errorf("delete nth settings from keyring: %w", err)
	}

	var result ResetResult
	if m.getenv != nil {
		for _, name := range []string{tokenEnvironment, clientIDEnvironment} {
			if strings.TrimSpace(m.getenv(name)) != "" {
				result.Environment = append(result.Environment, name)
			}
		}
	}

	return result, nil
}
