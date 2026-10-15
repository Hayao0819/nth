//go:build !windows

package app

import (
	"io"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type terminalInputReader struct {
	*os.File
	reader *committedTextReader
}

func terminalInput() io.Reader {
	return &terminalInputReader{
		File:   os.Stdin,
		reader: &committedTextReader{source: os.Stdin},
	}
}

func (r *terminalInputReader) Read(buffer []byte) (int, error) {
	return r.reader.Read(buffer)
}

type committedTextReader struct {
	source      io.Reader
	pending     []byte
	deferredErr error
}

func (r *committedTextReader) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	if len(r.pending) > 0 {
		read := copy(buffer, r.pending)
		r.pending = r.pending[read:]

		return read, nil
	}
	if r.deferredErr != nil {
		err := r.deferredErr
		r.deferredErr = nil

		return 0, err
	}

	read, err := r.source.Read(buffer)
	if read == 0 {
		return 0, err
	}
	if err != nil {
		r.deferredErr = err
	}
	if !isCommittedText(buffer[:read]) {
		return read, nil
	}

	encoded := encodeCommittedText(buffer[:read])
	read = copy(buffer, encoded)
	r.pending = append(r.pending[:0], encoded[read:]...)

	return read, nil
}

func isCommittedText(input []byte) bool {
	if !utf8.Valid(input) || utf8.RuneCount(input) < 2 {
		return false
	}

	nonASCII := false
	for _, character := range string(input) {
		if unicode.IsControl(character) {
			return false
		}
		if character > unicode.MaxASCII {
			nonASCII = true
		}
	}

	return nonASCII
}

func encodeCommittedText(input []byte) []byte {
	characters := []rune(string(input))
	encoded := make([]byte, 0, len(input)*2)
	encoded = append(encoded, '\x1b', '[')
	encoded = strconv.AppendInt(encoded, int64(characters[0]), 10)
	encoded = append(encoded, ';', '1', ';')
	for index, character := range characters {
		if index > 0 {
			encoded = append(encoded, ':')
		}
		encoded = strconv.AppendInt(encoded, int64(character), 10)
	}

	return append(encoded, 'u')
}
