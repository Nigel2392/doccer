package filesystem

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"
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

// isTextFile returns true if the file is a text file
func isTextFile(content []byte) bool {
	fileScanner := bufio.NewScanner(
		bytes.NewReader(content),
	)
	fileScanner.Split(bufio.ScanLines)
	fileScanner.Scan()
	return utf8.ValidString(fileScanner.Text())
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
	Path string `json:"path"`

	// Documentation root directory
	Root string `json:"root"`

	// Output directory
	Output string `json:"output"`

	// Relative output directory path
	Relative string `json:"relative"`

	// Directory name
	Name string `json:"name"`

	// Depth in the directory tree
	Depth int `json:"depth"`

	// The root directory, nil if it is the root
	RootDirectory *TemplateDirectory `json:"-"`
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
	//if d, ok := next.(*TemplateDirectory); ok {
	//	return d.Index
	//}
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
