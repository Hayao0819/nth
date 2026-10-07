package list

import (
	"strings"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
)

const (
	NameKey        = "name"
	DescriptionKey = "description"
	PrivateKey     = "private"
)

func NewForm(theme ui.Theme, title, button string, item *north.List) *dialog.Form {
	name, description, private := "", "", false
	if item != nil {
		name, private = item.Name, item.Private
		if item.Description != nil {
			description = *item.Description
		}
	}

	return dialog.NewForm(theme, title, button,
		dialog.Field{Key: NameKey, Label: "Name", Placeholder: "List name", Value: name, Required: true},
		dialog.Field{Key: DescriptionKey, Label: "Description", Placeholder: "Optional description", Value: description},
		dialog.Field{Key: PrivateKey, Label: "Private", Kind: dialog.ToggleField, Checked: private},
	)
}

func CreateRequest(result dialog.FormResult) north.CreateListRequest {
	request := north.CreateListRequest{
		Name:    strings.TrimSpace(result.Values[NameKey]),
		Private: result.Toggles[PrivateKey],
	}
	if description := strings.TrimSpace(result.Values[DescriptionKey]); description != "" {
		request.Description = &description
	}

	return request
}

func UpdateRequest(result dialog.FormResult) north.UpdateListRequest {
	name := strings.TrimSpace(result.Values[NameKey])
	private := result.Toggles[PrivateKey]

	return north.UpdateListRequest{
		Name:        &name,
		Description: north.NewNullableString(strings.TrimSpace(result.Values[DescriptionKey])),
		Private:     &private,
	}
}
