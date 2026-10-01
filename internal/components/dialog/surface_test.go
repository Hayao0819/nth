package dialog

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestDialogSurfaceRestoresColorsAfterNestedReset(t *testing.T) {
	t.Parallel()

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("235"))
	nested := lipgloss.NewStyle().Bold(true).Render("name") + " plain"
	rendered := DialogSurface(style, nested)
	base := ansi.Style{}.
		ForegroundColor(style.GetForeground()).
		BackgroundColor(style.GetBackground()).
		String()
	if !strings.Contains(rendered, ansi.ResetStyle+base+" plain") {
		t.Fatalf("dialog colors were not restored after reset: %q", rendered)
	}
}

func TestDialogSurfaceRestoresIndividuallyResetColors(t *testing.T) {
	t.Parallel()

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("235"))
	rendered := DialogSurface(style, "foreground\x1b[39m background\x1b[49m both\x1b[39;49m")
	foreground := ansi.Style{}.ForegroundColor(style.GetForeground()).String()
	background := ansi.Style{}.BackgroundColor(style.GetBackground()).String()
	base := ansi.Style{}.
		ForegroundColor(style.GetForeground()).
		BackgroundColor(style.GetBackground()).
		String()
	for _, want := range []string{
		"\x1b[39m" + foreground,
		"\x1b[49m" + background,
		"\x1b[39;49m" + base,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("dialog color was not restored after %q: %q", want, rendered)
		}
	}
}
