package pageheader

import (
	"strings"
	"testing"

	"github.com/Hayao0819/nth/internal/ui"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderAndBackHitArea(t *testing.T) {
	t.Parallel()

	header := Render(ui.NewTheme(), "Profile", "Loading…", 40)
	lines := strings.Split(header, "\n")
	if len(lines) != Height {
		t.Fatalf("header height = %d, want %d", len(lines), Height)
	}
	if width := ansi.StringWidth(lines[1]); width != 40 {
		t.Fatalf("header line width = %d, want 40: %q", width, lines[1])
	}
	if !BackAt(0, 0) || !BackAt(5, Height-1) || BackAt(6, 1) || BackAt(0, Height) {
		t.Fatal("back-button hit area does not match the rendered header")
	}
	if _, ok := Back()().(BackMsg); !ok {
		t.Fatal("Back did not return BackMsg")
	}
}

func TestRenderWithActionKeepsActionVisibleAndClickable(t *testing.T) {
	t.Parallel()

	theme := ui.NewTheme()
	header := RenderWithAction(theme, "Post", "A status message that may be clipped", "w  Browser ↗", 32)
	plain := ansi.Strip(header)
	if !strings.Contains(plain, "Post") || !strings.Contains(plain, "w  Browser ↗") {
		t.Fatalf("header action is not visible:\n%s", plain)
	}
	if !ActionAt(theme, 31, 1, 32, "w  Browser ↗") || ActionAt(theme, 10, 1, 32, "w  Browser ↗") {
		t.Fatal("header action hit area does not match the rendered action")
	}
	if ActionAt(theme, 31, Height, 32, "w  Browser ↗") {
		t.Fatal("header action extends below the header")
	}
}
