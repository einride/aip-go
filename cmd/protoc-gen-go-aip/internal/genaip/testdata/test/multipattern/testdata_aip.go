// Code generated by protoc-gen-go-aip. DO NOT EDIT.
//
// versions:
// 	protoc-gen-go-aip development
// 	protoc (unknown)
// source: test/multipattern/testdata.proto

package multipattern

import (
	fmt "fmt"
	resourcename "go.einride.tech/aip/resourcename"
	strings "strings"
)

type BookMultiPatternResourceName interface {
	fmt.Stringer
	MarshalString() (string, error)
	ContainsWildcard() bool
}

func ParseBookMultiPatternResourceName(name string) (BookMultiPatternResourceName, error) {
	switch {
	case resourcename.Match("shelves/{shelf}/books/{book}", name):
		var result ShelvesBookResourceName
		return &result, result.UnmarshalString(name)
	case resourcename.Match("publishers/{publisher}/books/{book}", name):
		var result PublishersBookResourceName
		return &result, result.UnmarshalString(name)
	default:
		return nil, fmt.Errorf("no matching pattern")
	}
}

type ShelvesBookResourceName struct {
	Shelf string
	Book  string
}

func (n ShelfResourceName) ShelvesBookResourceName(
	book string,
) ShelvesBookResourceName {
	return ShelvesBookResourceName{
		Shelf: n.Shelf,
		Book:  book,
	}
}

func (n ShelvesBookResourceName) Validate() error {
	if n.Shelf == "" {
		return fmt.Errorf("shelf: empty")
	}
	if strings.IndexByte(n.Shelf, '/') != -1 {
		return fmt.Errorf("shelf: contains illegal character '/'")
	}
	if n.Book == "" {
		return fmt.Errorf("book: empty")
	}
	if strings.IndexByte(n.Book, '/') != -1 {
		return fmt.Errorf("book: contains illegal character '/'")
	}
	return nil
}

func (n ShelvesBookResourceName) ContainsWildcard() bool {
	return false || n.Shelf == "-" || n.Book == "-"
}

func (n ShelvesBookResourceName) String() string {
	return resourcename.Sprint(
		"shelves/{shelf}/books/{book}",
		n.Shelf,
		n.Book,
	)
}

func (n ShelvesBookResourceName) MarshalString() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	return n.String(), nil
}

func (n *ShelvesBookResourceName) UnmarshalString(name string) error {
	return resourcename.Sscan(
		name,
		"shelves/{shelf}/books/{book}",
		&n.Shelf,
		&n.Book,
	)
}

func (n ShelvesBookResourceName) ShelfResourceName() ShelfResourceName {
	return ShelfResourceName{
		Shelf: n.Shelf,
	}
}

type PublishersBookResourceName struct {
	Publisher string
	Book      string
}

func (n PublishersBookResourceName) Validate() error {
	if n.Publisher == "" {
		return fmt.Errorf("publisher: empty")
	}
	if strings.IndexByte(n.Publisher, '/') != -1 {
		return fmt.Errorf("publisher: contains illegal character '/'")
	}
	if n.Book == "" {
		return fmt.Errorf("book: empty")
	}
	if strings.IndexByte(n.Book, '/') != -1 {
		return fmt.Errorf("book: contains illegal character '/'")
	}
	return nil
}

func (n PublishersBookResourceName) ContainsWildcard() bool {
	return false || n.Publisher == "-" || n.Book == "-"
}

func (n PublishersBookResourceName) String() string {
	return resourcename.Sprint(
		"publishers/{publisher}/books/{book}",
		n.Publisher,
		n.Book,
	)
}

func (n PublishersBookResourceName) MarshalString() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	return n.String(), nil
}

func (n *PublishersBookResourceName) UnmarshalString(name string) error {
	return resourcename.Sscan(
		name,
		"publishers/{publisher}/books/{book}",
		&n.Publisher,
		&n.Book,
	)
}

type ShelfMultiPatternResourceName interface {
	fmt.Stringer
	MarshalString() (string, error)
	ContainsWildcard() bool
}

func ParseShelfMultiPatternResourceName(name string) (ShelfMultiPatternResourceName, error) {
	switch {
	case resourcename.Match("shelves/{shelf}", name):
		var result ShelfResourceName
		return &result, result.UnmarshalString(name)
	case resourcename.Match("libraries/{library}/shelves/{shelf}", name):
		var result LibrariesShelfResourceName
		return &result, result.UnmarshalString(name)
	case resourcename.Match("rooms/{room}/shelves/{shelf}", name):
		var result RoomsShelfResourceName
		return &result, result.UnmarshalString(name)
	default:
		return nil, fmt.Errorf("no matching pattern")
	}
}

type ShelfResourceName struct {
	Shelf string
}

func (n ShelfResourceName) Validate() error {
	if n.Shelf == "" {
		return fmt.Errorf("shelf: empty")
	}
	if strings.IndexByte(n.Shelf, '/') != -1 {
		return fmt.Errorf("shelf: contains illegal character '/'")
	}
	return nil
}

func (n ShelfResourceName) ContainsWildcard() bool {
	return false || n.Shelf == "-"
}

func (n ShelfResourceName) String() string {
	return resourcename.Sprint(
		"shelves/{shelf}",
		n.Shelf,
	)
}

func (n ShelfResourceName) MarshalString() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	return n.String(), nil
}

func (n *ShelfResourceName) UnmarshalString(name string) error {
	return resourcename.Sscan(
		name,
		"shelves/{shelf}",
		&n.Shelf,
	)
}

type LibrariesShelfResourceName struct {
	Library string
	Shelf   string
}

func (n LibrariesShelfResourceName) Validate() error {
	if n.Library == "" {
		return fmt.Errorf("library: empty")
	}
	if strings.IndexByte(n.Library, '/') != -1 {
		return fmt.Errorf("library: contains illegal character '/'")
	}
	if n.Shelf == "" {
		return fmt.Errorf("shelf: empty")
	}
	if strings.IndexByte(n.Shelf, '/') != -1 {
		return fmt.Errorf("shelf: contains illegal character '/'")
	}
	return nil
}

func (n LibrariesShelfResourceName) ContainsWildcard() bool {
	return false || n.Library == "-" || n.Shelf == "-"
}

func (n LibrariesShelfResourceName) String() string {
	return resourcename.Sprint(
		"libraries/{library}/shelves/{shelf}",
		n.Library,
		n.Shelf,
	)
}

func (n LibrariesShelfResourceName) MarshalString() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	return n.String(), nil
}

func (n *LibrariesShelfResourceName) UnmarshalString(name string) error {
	return resourcename.Sscan(
		name,
		"libraries/{library}/shelves/{shelf}",
		&n.Library,
		&n.Shelf,
	)
}

type RoomsShelfResourceName struct {
	Room  string
	Shelf string
}

func (n RoomsShelfResourceName) Validate() error {
	if n.Room == "" {
		return fmt.Errorf("room: empty")
	}
	if strings.IndexByte(n.Room, '/') != -1 {
		return fmt.Errorf("room: contains illegal character '/'")
	}
	if n.Shelf == "" {
		return fmt.Errorf("shelf: empty")
	}
	if strings.IndexByte(n.Shelf, '/') != -1 {
		return fmt.Errorf("shelf: contains illegal character '/'")
	}
	return nil
}

func (n RoomsShelfResourceName) ContainsWildcard() bool {
	return false || n.Room == "-" || n.Shelf == "-"
}

func (n RoomsShelfResourceName) String() string {
	return resourcename.Sprint(
		"rooms/{room}/shelves/{shelf}",
		n.Room,
		n.Shelf,
	)
}

func (n RoomsShelfResourceName) MarshalString() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	return n.String(), nil
}

func (n *RoomsShelfResourceName) UnmarshalString(name string) error {
	return resourcename.Sscan(
		name,
		"rooms/{room}/shelves/{shelf}",
		&n.Room,
		&n.Shelf,
	)
}
