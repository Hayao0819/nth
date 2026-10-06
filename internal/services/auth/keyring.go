package auth

import (
	"errors"

	keyring "github.com/zalando/go-keyring"
)

const keyringService = "nth"

var errNotFound = errors.New("credential not found")

type vault interface {
	Get(string) (string, error)
	Set(string, string) error
	DeleteAll() error
}

type systemKeyring struct{}

func (systemKeyring) Get(name string) (string, error) {
	value, err := keyring.Get(keyringService, name)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", errNotFound
	}

	return value, err
}

func (systemKeyring) Set(name, value string) error {
	return keyring.Set(keyringService, name, value)
}

func (systemKeyring) DeleteAll() error {
	return keyring.DeleteAll(keyringService)
}
