package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

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
