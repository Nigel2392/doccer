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

type FSBase struct {
	// Absolute directory path
	Path string

	// Documentation root directory
	Root string

	// Output directory
	Output string

	// Relative output directory path
	Relative string

	// Directory name
	Name string

	// Depth in the directory tree
	Depth int
}
