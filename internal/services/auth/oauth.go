package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Hayao0819/go-north"
	"golang.org/x/oauth2"
)

const defaultOAuthClientID = "cmuql9qx700fswhmtdtda1ozo"

var oauthScopes = []north.Scope{
	north.ScopePostsRead,
	north.ScopeUsersRead,
	north.ScopePostsWrite,
	north.ScopePostsEdit,
	north.ScopePostsDelete,
	north.ScopePollsVote,
	north.ScopeReactionsWrite,
	north.ScopeFollowsWrite,
	north.ScopeModerationRead,
	north.ScopeModerationWrite,
	north.ScopeBookmarksRead,
	north.ScopeBookmarksWrite,
	north.ScopeProfileWrite,
	north.ScopeNotificationsRead,
	north.ScopeNotificationsWrite,
	north.ScopeTrendsRead,
	north.ScopeTrendsWrite,
	north.ScopeListsRead,
	north.ScopeListsWrite,
	north.ScopeMediaWrite,
	north.ScopeDraftsRead,
	north.ScopeDraftsWrite,
	north.ScopePostsSchedule,
	north.ScopeDMRead,
	north.ScopeDMWrite,
	north.ScopeDMReceiptsWrite,
	north.ScopeDMDelete,
	north.ScopeDMConversationsWrite,
	north.ScopeDMRequestsWrite,
}

// OAuthCredential is the refreshable public-API credential stored by nth.
type OAuthCredential struct {
	ClientID string       `json:"clientId"`
	Token    oauth2.Token `json:"token"`
	Scopes   []string     `json:"scopes,omitempty"`
}

func (c OAuthCredential) Valid() bool {
	return strings.TrimSpace(c.ClientID) != "" &&
		(strings.TrimSpace(c.Token.AccessToken) != "" || strings.TrimSpace(c.Token.RefreshToken) != "")
}

// MissingScopes returns permissions required by this version of nth that were
// not recorded when the credential was issued.
func (c OAuthCredential) MissingScopes() []string {
	granted := make(map[string]struct{}, len(c.Scopes))
	for _, scope := range c.Scopes {
		granted[strings.TrimSpace(scope)] = struct{}{}
	}
	missing := make([]string, 0, len(oauthScopes))
	for _, scope := range oauthScopes {
		name := string(scope)
		if _, ok := granted[name]; !ok {
			missing = append(missing, name)
		}
	}

	return missing
}

// NeedsAuthorization reports whether nth now requires permissions that are
// absent from the saved authorization metadata.
func (c OAuthCredential) NeedsAuthorization() bool {
	return c.Valid() && len(c.MissingScopes()) > 0
}

// RequiredOAuthScopes returns the permissions requested by this version of nth.
func RequiredOAuthScopes() []string {
	return scopeNames(oauthScopes)
}

// OAuthPrompt contains the values displayed while device authorization is in progress.
type OAuthPrompt struct {
	VerificationURI         string
	VerificationURIComplete string
	UserCode                string
}

// OAuthSession is one in-progress device authorization.
type OAuthSession interface {
	Prompt() OAuthPrompt
	Wait(context.Context) (OAuthCredential, error)
}

type OAuthStartFunc func(context.Context) (OAuthSession, error)

type deviceOAuthSession struct {
	clientID      string
	config        *oauth2.Config
	authorization *oauth2.DeviceAuthResponse
}

func (s *deviceOAuthSession) Prompt() OAuthPrompt {
	return OAuthPrompt{
		VerificationURI:         s.authorization.VerificationURI,
		VerificationURIComplete: s.authorization.VerificationURIComplete,
		UserCode:                s.authorization.UserCode,
	}
}

func (s *deviceOAuthSession) Wait(ctx context.Context) (OAuthCredential, error) {
	token, err := s.config.DeviceAccessToken(ctx, s.authorization)
	if err != nil {
		return OAuthCredential{}, fmt.Errorf("complete north sign-in: %w", err)
	}

	return OAuthCredential{
		ClientID: s.clientID,
		Token:    *token,
		Scopes:   grantedScopes(token, s.config.Scopes),
	}, nil
}

// StartOAuth begins north's device authorization flow.
func (m *Manager) StartOAuth(ctx context.Context) (OAuthSession, error) {
	clientID := m.oauthClientID()
	buildConfig := m.oauthConfig
	if buildConfig == nil {
		buildConfig = newOAuthConfig
	}
	config := buildConfig(clientID)
	authorization, err := config.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("start north sign-in: %w", err)
	}

	return &deviceOAuthSession{
		clientID:      clientID,
		config:        config,
		authorization: authorization,
	}, nil
}

func (m *Manager) oauthClientID() string {
	if m.getenv != nil {
		if override := strings.TrimSpace(m.getenv(clientIDEnvironment)); override != "" {
			return override
		}
	}

	return defaultOAuthClientID
}

// SaveOAuthToken replaces the stored token after OAuth refresh.
func (m *Manager) SaveOAuthToken(clientID string, token *oauth2.Token) error {
	credential := OAuthCredential{ClientID: strings.TrimSpace(clientID)}
	if token != nil {
		credential.Token = *token
	}
	if !credential.Valid() {
		return errors.New("save OAuth token: invalid credential")
	}
	existing, available, err := m.loadOAuth()
	if err != nil {
		return err
	}
	if available && existing.ClientID == credential.ClientID {
		credential.Scopes = append([]string(nil), existing.Scopes...)
	}

	return m.saveOAuth(credential)
}

func (m *Manager) loadOAuth() (OAuthCredential, bool, error) {
	encoded, err := m.vault.Get(oauthEntry)
	switch {
	case err == nil:
		var credential OAuthCredential
		if err := json.Unmarshal([]byte(encoded), &credential); err != nil {
			return OAuthCredential{}, false, fmt.Errorf("decode OAuth token from keyring: %w", err)
		}

		return credential, credential.Valid(), nil
	case errors.Is(err, errNotFound):
		return OAuthCredential{}, false, nil
	default:
		return OAuthCredential{}, false, fmt.Errorf("read OAuth token from keyring: %w", err)
	}
}

func (m *Manager) saveOAuth(credential OAuthCredential) error {
	encoded, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("encode OAuth token: %w", err)
	}
	if err := m.vault.Set(oauthEntry, string(encoded)); err != nil {
		return fmt.Errorf("save OAuth token in keyring: %w", err)
	}

	return nil
}

func newOAuthConfig(clientID string) *oauth2.Config {
	return north.NewOAuthConfig(clientID, "", oauthScopes...)
}

// OAuthConfig returns the OAuth configuration used by nth.
func OAuthConfig(clientID string) *oauth2.Config {
	return newOAuthConfig(clientID)
}

func scopeNames(scopes []north.Scope) []string {
	names := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		names = append(names, string(scope))
	}

	return names
}

func grantedScopes(token *oauth2.Token, requested []string) []string {
	if token != nil {
		switch granted := token.Extra("scope").(type) {
		case string:
			if scopes := strings.Fields(granted); len(scopes) > 0 {
				return scopes
			}
		case []string:
			if len(granted) > 0 {
				return append([]string(nil), granted...)
			}
		}
	}

	return append([]string(nil), requested...)
}
