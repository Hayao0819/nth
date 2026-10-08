package browser

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

type launcher struct {
	name string
	args []string
}

func Open(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if err := validateURL(rawURL); err != nil {
		return err
	}

	var failures []error
	found := false
	for _, candidate := range launchers(runtime.GOOS, rawURL) {
		path, err := exec.LookPath(candidate.name)
		if err != nil {
			continue
		}
		found = true
		if err := exec.Command(path, candidate.args...).Run(); err == nil {
			return nil
		} else {
			failures = append(failures, fmt.Errorf("%s: %w", candidate.name, err))
		}
	}

	if !found {
		return errors.New("no browser launcher was found")
	}

	return fmt.Errorf("browser launch failed: %w", errors.Join(failures...))
}

func validateURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Scheme != "https" && parsed.Scheme != "http" {
		return errors.New("invalid web URL")
	}

	return nil
}

func launchers(goos, rawURL string) []launcher {
	switch goos {
	case "darwin":
		return []launcher{{name: "open", args: []string{rawURL}}}
	case "windows":
		return []launcher{{name: "rundll32", args: []string{"url.dll,FileProtocolHandler", rawURL}}}
	case "linux", "freebsd", "openbsd", "netbsd", "dragonfly":
		return []launcher{
			{name: "xdg-open", args: []string{rawURL}},
			{name: "gio", args: []string{"open", rawURL}},
		}
	default:
		return nil
	}
}
