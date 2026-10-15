//go:build windows

package app

import (
	"io"
	"os"
)

func terminalInput() io.Reader {
	return os.Stdin
}
