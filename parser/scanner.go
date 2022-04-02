package parser

import (
	"bufio"
	"io"
)

type Scanner struct {
	buffer []string
	line int
}

func NewScanner(reader io.Reader) *Scanner {
	s := &Scanner{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		s.buffer = append(s.buffer, scanner.Text())
	}
	return s
}

func (s *Scanner) String() string {
	out := ""
	for i,line := range s.buffer {
		if i == s.line - 1 {
			out += "* "	
		} else {
			out += "  "
		}
		out += line
		if i != len(s.buffer) - 1 {
			out += "\n"
		}
	}
	return out
}

func (s *Scanner) Scan() bool {
	if s.line <= len(s.buffer) {
		s.line += 1
	}
	return s.line <= len(s.buffer)
}

func (s *Scanner) Text() string {
	if s.line == 0 || s.EOF() {
		return ""
	}
	return s.buffer[s.line - 1]
}

func (s *Scanner) Backtrack() bool {
	if s.line > 0 {
		s.line -= 1
	}
	return s.line > 0
}

func (s *Scanner) Line() int {
	return s.line
}

func (s *Scanner) EOF() bool {
	return s.line > len(s.buffer)
}
