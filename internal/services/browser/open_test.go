package browser

import "testing"

func TestValidateURL(t *testing.T) {
	t.Parallel()

	for _, rawURL := range []string{"https://north.rip/alice", "http://localhost:3000/post/1"} {
		if err := validateURL(rawURL); err != nil {
			t.Errorf("validateURL(%q): %v", rawURL, err)
		}
	}
	for _, rawURL := range []string{"", "north.rip/alice", "file:///tmp/post", "https://user:secret@north.rip"} {
		if err := validateURL(rawURL); err == nil {
			t.Errorf("validateURL(%q) succeeded", rawURL)
		}
	}
}

func TestLaunchers(t *testing.T) {
	t.Parallel()

	if got := launchers("darwin", "https://north.rip"); len(got) != 1 || got[0].name != "open" {
		t.Fatalf("darwin launchers = %#v", got)
	}
	if got := launchers("windows", "https://north.rip"); len(got) != 1 || got[0].name != "rundll32" {
		t.Fatalf("windows launchers = %#v", got)
	}
	if got := launchers("linux", "https://north.rip"); len(got) != 2 || got[0].name != "xdg-open" || got[1].name != "gio" {
		t.Fatalf("linux launchers = %#v", got)
	}
	if got := launchers("plan9", "https://north.rip"); len(got) != 0 {
		t.Fatalf("unsupported launchers = %#v", got)
	}
}
