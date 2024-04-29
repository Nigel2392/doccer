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
	GetTitle() string
	GetNext() Object
	GetPrevious() Object
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

	// The root directory, nil if it is the root
	RootDirectory *TemplateDirectory
}

type Config struct {
	Title    string   // Title of the object
	Next     []string // Path to the next object
	Previous []string // Path to the previous object
	t        *FSBase
}

func NewConfig(t *FSBase) Config {
	return Config{t: t}
}

func (d *Config) depthString() string {
	return strings.Repeat("  ", d.t.Depth)
}

// Format the template as a string
func (t *Config) String() string {
	return t.t.Relative
}

// GetName returns the name of the template
func (t *Config) GetName() string {
	return t.t.Name
}

// GetTitle returns the title of the template
func (t *Config) GetTitle() string {
	if t.Title != "" {
		return t.Title
	}
	return t.t.Name
}

// GetNext returns the next object for the template
func (d *Config) GetNext() Object {
	var next, ok = d.t.RootDirectory.Walk(d.Next)
	if !ok {
		return nil
	}
	return next
}

// GetPrevious returns the previous object for the template
func (d *Config) GetPrevious() Object {
	var prev, ok = d.t.RootDirectory.Walk(d.Previous)
	if !ok {
		return nil
	}
	return prev
}
