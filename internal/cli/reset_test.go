package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Hayao0819/nth/internal/services/auth"
)

func TestResetCommandDeletesSavedSettings(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{resetResult: auth.ResetResult{
		Environment: []string{"NORTH_API_KEY", "NORTH_CLIENT_ID"},
	}}
	command := newCommandWith("test", dependencies{credentials: credentials})
	var output, errorOutput bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&errorOutput)
	command.SetArgs([]string{"reset"})

	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if credentials.resets != 1 {
		t.Fatalf("reset calls = %d", credentials.resets)
	}
	if !strings.Contains(output.String(), "Removed nth settings") {
		t.Fatalf("output = %q", output.String())
	}
	if warning := errorOutput.String(); !strings.Contains(warning, "NORTH_API_KEY") || !strings.Contains(warning, "NORTH_CLIENT_ID") {
		t.Fatalf("warning = %q", warning)
	}
}

func TestResetCommandReportsKeyringFailure(t *testing.T) {
	t.Parallel()

	want := errors.New("keyring is locked")
	credentials := &fakeCredentials{resetErr: want}
	command := newCommandWith("test", dependencies{credentials: credentials})
	command.SetArgs([]string{"reset"})

	if err := command.Execute(); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}
