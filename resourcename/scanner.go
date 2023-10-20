package resourcename

import (
	"strings"
)

// Scanner scans a resource name.
type Scanner struct {
	name                     string
	start, end               int
	serviceStart, serviceEnd int
	full                     bool
}

// Init initializes the scanner.
func (s *Scanner) Init(name string) {
	s.name = name
	s.start, s.end = 0, 0
	s.full = false
}

// Scan to the next segment.
func (s *Scanner) Scan() bool {
	if s.name == "/" {
		return false
	}
	switch s.end {
	case len(s.name):
		return false
	case 0:
		// Special case for full resource names.
		if strings.HasPrefix(s.name, "//") {
			s.full = true
			s.start, s.end = 2, 2
			nextSlash := strings.IndexByte(s.name[s.start:], '/')
			if nextSlash == -1 {
				s.serviceStart, s.serviceEnd = s.start, len(s.name)
				s.start, s.end = len(s.name), len(s.name)
				return false
			}
			s.serviceStart, s.serviceEnd = s.start, s.start+nextSlash
			s.start, s.end = s.start+nextSlash+1, s.start+nextSlash+1
		}
	default:
		s.start = s.end + 1 // start past latest slash
	}
	if nextSlash := strings.IndexByte(s.name[s.start:], '/'); nextSlash == -1 {
		s.end = len(s.name)
	} else {
		s.end = s.start + nextSlash
	}
	return true
}

// Start returns the start index (inclusive) of the current segment.
func (s *Scanner) Start() int {
	return s.start
}

// End returns the end index (exclusive) of the current segment.
func (s *Scanner) End() int {
	return s.end
}

// Segment returns the current segment.
func (s *Scanner) Segment() Segment {
	return Segment(s.name[s.start:s.end])
}

// Full returns true if the scanner has detected a full resource name.
func (s *Scanner) Full() bool {
	return s.full
}

// ServiceName returns the service name, when the scanner has detected a full resource name.
func (s *Scanner) ServiceName() string {
	return s.name[s.serviceStart:s.serviceEnd]
}
