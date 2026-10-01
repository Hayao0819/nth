package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
)

const northOrigin = "https://north.rip/api/"

var (
	errBrowserProfileNotFound = errors.New("browser profile was not found")
	errNorthCookiesNotFound   = errors.New("north.rip cookies were not found")
)

// Profile identifies one browser cookie store.
type Profile struct {
	Browser string `json:"browser"`
	Name    string `json:"profile,omitempty"`
	Path    string `json:"path,omitempty"`
	Default bool   `json:"-"`
}

func (p Profile) Valid() bool {
	return strings.TrimSpace(p.Browser) != ""
}

func (p Profile) Same(other Profile) bool {
	return strings.EqualFold(p.Browser, other.Browser) &&
		p.Name == other.Name && cleanPath(p.Path) == cleanPath(other.Path)
}

// Label is the browser and profile name shown in setup.
func (p Profile) Label() string {
	browser := browserName(p.Browser)
	if p.Name == "" {
		return browser
	}

	label := browser + " · " + p.Name
	if p.Default {
		label += " (default)"
	}

	return label
}

type browserSource interface {
	Profiles(context.Context) []Profile
	CookieHeader(context.Context, Profile) (string, error)
}

type localBrowsers struct{}

func (localBrowsers) Profiles(ctx context.Context) []Profile {
	stores := kooky.FindAllCookieStores(ctx)
	defer closeStores(stores)

	profiles := make([]Profile, 0, len(stores))
	for _, store := range stores {
		if store == nil {
			continue
		}
		profile := Profile{
			Browser: store.Browser(),
			Name:    store.Profile(),
			Path:    store.FilePath(),
			Default: store.IsDefaultProfile(),
		}
		if profile.Path != "" {
			if _, err := os.Stat(profile.Path); err != nil {
				continue
			}
		}
		if !profile.Valid() || containsProfile(profiles, profile) {
			continue
		}
		profiles = append(profiles, profile)
	}
	sortProfiles(profiles)

	return profiles
}

func (localBrowsers) CookieHeader(ctx context.Context, selected Profile) (string, error) {
	stores := kooky.FindAllCookieStores(ctx)
	defer closeStores(stores)

	matches := matchingStores(stores, selected)
	if len(matches) == 0 {
		return "", fmt.Errorf("%s: %w", selected.Label(), errBrowserProfileNotFound)
	}

	domain := kooky.FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.EqualFold(strings.TrimPrefix(cookie.Domain, "."), "north.rip")
	})
	var readErrors []error
	for _, store := range matches {
		cookies, err := store.TraverseCookies(domain, kooky.Valid).ReadAllCookies(ctx)
		if err != nil {
			readErrors = append(readErrors, err)
			continue
		}
		header, err := cookieHeader(cookies)
		if err == nil {
			return header, nil
		}
		readErrors = append(readErrors, err)
	}

	return "", fmt.Errorf("read %s cookies: %w", selected.Label(), errors.Join(readErrors...))
}

func cookieHeader(cookies []*kooky.Cookie) (string, error) {
	target, err := url.Parse(northOrigin)
	if err != nil {
		return "", err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	for _, cookie := range cookies {
		if cookie == nil || !strings.EqualFold(strings.TrimPrefix(cookie.Domain, "."), target.Hostname()) {
			continue
		}
		copy := cookie.Cookie
		jar.SetCookies(target, []*http.Cookie{&copy})
	}

	request := &http.Request{Header: make(http.Header)}
	for _, cookie := range jar.Cookies(target) {
		request.AddCookie(cookie)
	}
	header := request.Header.Get("Cookie")
	if header == "" {
		return "", errNorthCookiesNotFound
	}

	return header, nil
}

func matchingStores(stores []kooky.CookieStore, selected Profile) []kooky.CookieStore {
	var matches []kooky.CookieStore
	requirePath := selected.Path != ""
	if requirePath {
		if _, err := os.Stat(selected.Path); os.IsNotExist(err) {
			requirePath = false
		}
	}
	for _, store := range stores {
		if store == nil || !strings.EqualFold(store.Browser(), selected.Browser) {
			continue
		}
		if selected.Name != "" && store.Profile() != selected.Name {
			continue
		}
		if requirePath && cleanPath(store.FilePath()) != cleanPath(selected.Path) {
			continue
		}
		matches = append(matches, store)
	}
	slices.SortStableFunc(matches, func(left, right kooky.CookieStore) int {
		if left.IsDefaultProfile() != right.IsDefaultProfile() {
			if left.IsDefaultProfile() {
				return -1
			}
			return 1
		}

		return strings.Compare(cleanPath(left.FilePath()), cleanPath(right.FilePath()))
	})

	return matches
}

func closeStores(stores []kooky.CookieStore) {
	for _, store := range stores {
		if store != nil {
			_ = store.Close()
		}
	}
}

func containsProfile(profiles []Profile, candidate Profile) bool {
	return slices.ContainsFunc(profiles, candidate.Same)
}

func sortProfiles(profiles []Profile) {
	slices.SortStableFunc(profiles, func(left, right Profile) int {
		if byBrowser := strings.Compare(strings.ToLower(left.Browser), strings.ToLower(right.Browser)); byBrowser != 0 {
			return byBrowser
		}
		if left.Default != right.Default {
			if left.Default {
				return -1
			}
			return 1
		}
		if byName := strings.Compare(strings.ToLower(left.Name), strings.ToLower(right.Name)); byName != 0 {
			return byName
		}

		return strings.Compare(cleanPath(left.Path), cleanPath(right.Path))
	})
}

func cleanPath(path string) string {
	if path == "" {
		return ""
	}

	return filepath.Clean(path)
}

func browserName(browser string) string {
	names := map[string]string{
		"brave":    "Brave",
		"browsh":   "Browsh",
		"chrome":   "Google Chrome",
		"chromium": "Chromium",
		"edge":     "Microsoft Edge",
		"firefox":  "Firefox",
		"ie":       "Internet Explorer",
		"opera":    "Opera",
		"safari":   "Safari",
	}
	if name := names[strings.ToLower(browser)]; name != "" {
		return name
	}

	return browser
}
