package ui

import "testing"

func TestThemeHasVisualHierarchy(t *testing.T) {
	t.Parallel()

	theme := NewTheme()
	for name, style := range map[string]bool{
		"brand":       theme.Brand.GetBold(),
		"title":       theme.Title.GetBold(),
		"section":     theme.Section.GetBold(),
		"modal title": theme.ModalTitle.GetBold(),
		"post author": theme.Name.GetBold(),
	} {
		if !style {
			t.Errorf("%s is not bold", name)
		}
	}
	if theme.ModalTitle.GetBackground() == nil {
		t.Error("modal title has no background")
	}
	if theme.Dialog.GetBackground() == nil {
		t.Error("dialog has no background")
	}
}
