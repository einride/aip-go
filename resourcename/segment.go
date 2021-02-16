package resourcename

import "strings"

// RevisionSeparator is the separator character used to separate resource IDs from revision IDs.
const RevisionSeparator = '@'

// Segment is a segment of a resource name or a resource name pattern.
//
// EBNF
//
//  Segment  = Literal | Variable ;
//  Variable = "{" Literal "}" ;
type Segment string

// IsVariable reports whether the segment is a variable segment.
func (s Segment) IsVariable() bool {
	return len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}'
}

// Literal returns the literal value of the segment.
// For variables, the literal value is the name of the variable.
func (s Segment) Literal() Literal {
	switch {
	case s.IsVariable():
		return Literal(s[1 : len(s)-1])
	default:
		return Literal(s)
	}
}

// IsWildcard reports whether the segment is a wildcard.
func (s Segment) IsWildcard() bool {
	return s == "-"
}

// Literal is the literal part of a resource name segment.
//
// EBNF
//
//  Literal  = RESOURCE_ID | RevisionLiteral ;
//  RevisionLiteral = RESOURCE_ID "@" REVISION_ID ;
type Literal string

// ResourceID returns the literal's resource ID.
func (l Literal) ResourceID() string {
	if !l.HasRevision() {
		return string(l)
	}
	return string(l[:strings.IndexByte(string(l), RevisionSeparator)])
}

// RevisionID returns the literal's revision ID.
func (l Literal) RevisionID() string {
	if !l.HasRevision() {
		return ""
	}
	return string(l[strings.IndexByte(string(l), RevisionSeparator)+1:])
}

// HasRevision returns true if the literal has a valid revision.
func (l Literal) HasRevision() bool {
	revisionSeparatorIndex := strings.IndexByte(string(l), RevisionSeparator)
	if revisionSeparatorIndex < 1 || revisionSeparatorIndex >= len(l)-1 {
		return false // must have content on each side of the revision marker
	}
	if strings.IndexByte(string(l[revisionSeparatorIndex+1:]), RevisionSeparator) != -1 {
		return false // multiple revision markers means no valid revision
	}
	return true
}
