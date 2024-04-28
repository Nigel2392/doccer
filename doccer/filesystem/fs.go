package filesystem

import (
	"errors"
	"strings"
)

var (
	// ErrNoTemplates is returned when there are no templates in a directory
	ErrNoTemplates = errors.New("no templates found in directory")
)

// isIndexFile returns true if the file is an index file
func IsIndexFile(name string) bool {
	return strings.HasPrefix(name, "index.") ||
		strings.HasPrefix(strings.ToLower(name), "readme.")
}

// Object represents a documentation object
// It might be a directory or file.
type Object interface {
	String() string
	GetName() string
	IsDirectory() bool
	ServeURL() string
	URL() string
}
