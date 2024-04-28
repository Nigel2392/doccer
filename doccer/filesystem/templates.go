package filesystem

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// Template represents a documentation template
type Template struct {
	// Template name
	Name string

	// Absolute template path
	Path string

	// Documentation root directory
	Root string

	// Output directory
	Output string

	// Relative output directory path
	Relative string

	// Template content
	Content []byte

	// Depth in the directory tree
	Depth int
}

func (d *Template) depthString() string {
	return strings.Repeat("  ", d.Depth)
}

// Format the template as a string
func (t *Template) String() string {
	return t.Relative
}

// Format the template for %v
func (t *Template) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%sTemplate: %s", t.depthString(), t.Name)
}

// GetName returns the name of the template
func (t *Template) GetName() string {
	return t.Name
}

// IsDirectory returns true if the object is a directory
func (d *Template) IsDirectory() bool {
	return false
}

// NewTemplate creates a new template
func NewTemplate(name, root, filepath, output, relative string, depth int) (*Template, error) {

	if !strings.HasSuffix(output, ".html") {
		var (
			ext  = path.Ext(output)
			base = output[:len(output)-len(ext)]
		)

		output = fmt.Sprintf("%s.html", base)
	}

	var template = &Template{
		Name:     name,
		Path:     filepath,
		Root:     root,
		Output:   output,
		Relative: relative,
		Depth:    depth,
	}

	var err = template.loadContent()

	return template, err
}

func (t *Template) URL() string {
	var (
		ext      = path.Ext(t.Relative)
		relative = t.Relative[:len(t.Relative)-len(ext)]
	)
	return fmt.Sprintf("%s.html", relative)

}

func (t *Template) ServeURL() string {
	var output = strings.Replace(t.Relative, "\\", "/", -1)
	if strings.HasPrefix(output, "/") {
		return output
	}

	return fmt.Sprintf("/%s", output)
}

// loadContent loads the template content from disk
func (t *Template) loadContent() error {
	// Load the template content
	var content, err = os.ReadFile(t.Path)
	if err != nil {
		return err
	}

	t.Content = content
	return nil
}
