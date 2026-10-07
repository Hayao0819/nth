package auth

import "strings"

// Method identifies the credential most recently configured by setup.
type Method string

const (
	MethodBrowser  Method = "browser"
	MethodAPIToken Method = "api-token"
	MethodOAuth    Method = "oauth"
)

func (m Method) Valid() bool {
	return m == MethodBrowser || m == MethodAPIToken || m == MethodOAuth
}

// Settings contains the values collected by initial setup.
type Settings struct {
	Method  Method
	Token   string
	OAuth   OAuthCredential
	Browser Profile
	Images  bool

	apiTokenFromEnvironment bool
}

func (s Settings) HasAPIToken() bool {
	return strings.TrimSpace(s.Token) != ""
}

func (s Settings) HasBrowser() bool {
	return s.Browser.Valid()
}

func (s Settings) HasOAuth() bool {
	return s.OAuth.Valid()
}

// PreferOAuth reports whether the selected OAuth login should be used instead
// of an API token stored in the keyring.
func (s Settings) PreferOAuth() bool {
	return s.Method == MethodOAuth && s.HasOAuth() && !s.apiTokenFromEnvironment
}

// UsesOAuth reports whether the public API client will use the saved OAuth
// credential after applying API-token precedence.
func (s Settings) UsesOAuth() bool {
	return s.HasOAuth() && (!s.HasAPIToken() || s.PreferOAuth())
}

// OAuthNeedsAuthorization reports whether the active OAuth credential lacks
// permissions required by this version of nth.
func (s Settings) OAuthNeedsAuthorization() bool {
	return s.UsesOAuth() && s.OAuth.NeedsAuthorization()
}

// HasEnvironmentToken reports whether NORTH_API_KEY supplied the API token.
func (s Settings) HasEnvironmentToken() bool {
	return s.apiTokenFromEnvironment
}

// Complete reports whether at least one authentication method is configured.
func (s Settings) Complete() bool {
	return s.HasAPIToken() || s.HasOAuth() || s.HasBrowser()
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

// ResetResult reports settings outside nth's system keyring.
type ResetResult struct {
	Environment []string
}
