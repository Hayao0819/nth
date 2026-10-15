//go:build !windows

package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestCommittedTextReaderEncodesOneIMECommit(t *testing.T) {
	t.Parallel()

	reader := &committedTextReader{source: bytes.NewBufferString("あいうえお")}
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b[12354;1;12354:12356:12358:12360:12362u"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCommittedTextReaderPreservesKeyboardInput(t *testing.T) {
	t.Parallel()

	tests := []string{
		"a",
		"abc",
		"あ",
		"\x1b[A",
		"\x1b[200~日本語\x1b[201~",
	}
	for _, input := range tests {
		reader := &committedTextReader{source: bytes.NewBufferString(input)}
		got, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("read %q: %v", input, err)
		}
		if string(got) != input {
			t.Errorf("got %q, want %q", got, input)
		}
	}
}

func TestCommittedTextReaderHandlesSmallBuffers(t *testing.T) {
	t.Parallel()

	reader := &committedTextReader{source: bytes.NewBufferString("日本語入力")}
	var got bytes.Buffer
	buffer := make([]byte, len("日本語入力")+1)
	for {
		read, err := reader.Read(buffer)
		got.Write(buffer[:read])
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	want := string(encodeCommittedText([]byte("日本語入力")))
	if got.String() != want {
		t.Fatalf("got %q, want %q", got.String(), want)
	}
}

func TestCommittedTextReaderProducesOneKeyEvent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reader := &committedTextReader{source: bytes.NewBufferString("あいうえお")}
	program := tea.NewProgram(
		&inputProbe{},
		tea.WithContext(ctx),
		tea.WithInput(reader),
		tea.WithOutput(io.Discard),
		tea.WithoutRenderer(),
		tea.WithoutSignalHandler(),
	)
	final, err := program.Run()
	if err != nil {
		t.Fatal(err)
	}
	probe := final.(*inputProbe)
	if probe.keys != 1 || probe.text != "あいうえお" {
		t.Fatalf("key events = %d, text = %q", probe.keys, probe.text)
	}
}

type inputProbe struct {
	keys int
	text string
}

func (*inputProbe) Init() tea.Cmd {
	return nil
}

func (p *inputProbe) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	p.keys++
	p.text = key.Text

	return p, tea.Quit
}

func (*inputProbe) View() tea.View {
	return tea.NewView("")
}
