package bufiox

import (
	"bufio"
	"io"
)

const bufferSizeBytes = 64 * 1024

// Scanner is a drop-in replacement for bufio.Scanner that will
// handle log lines that are longer than the default limit in
// bufio (64KB).
type Scanner struct {
	r     *bufio.Reader
	token []byte

	err error
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r: bufio.NewReaderSize(r, bufferSizeBytes),
	}
}

func (s *Scanner) Scan() bool {
	// Clear the in-memory token buffer, but retain its allocated memory.
	s.token = s.token[:0]

	for done := false; !done; {
		line, isPrefix, err := s.r.ReadLine()
		if err != nil {
			if err != io.EOF {
				s.err = err
			}
			return false
		}
		s.token = append(s.token, line...)

		done = !isPrefix
	}

	return true
}

func (s *Scanner) Bytes() []byte {
	return s.token
}

func (s *Scanner) Text() string {
	return string(s.token)
}

func (s *Scanner) Err() error {
	return s.err
}
