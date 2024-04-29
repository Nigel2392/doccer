package filesystem

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
)

// Template represents a documentation template
type Template struct {
	// Base filesystem object
	FSBase `json:",inline"`

	// Template configuration
	Config `json:",inline"`

	// Template content
	Content string `json:"content"`
}

// Format the template for %v
func (t *Template) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%sTemplate: %s", t.depthString(), t.Name)
}

// IsDirectory returns true if the object is a directory
func (d *Template) IsDirectory() bool {
	return false
}

// NewTemplate creates a new template
func NewTemplate(rootDir *TemplateDirectory, name, root, filepath, output, relative string, depth int) (*Template, error) {

	if !strings.HasSuffix(output, ".html") {
		var (
			ext  = path.Ext(output)
			base = output[:len(output)-len(ext)]
		)

		output = fmt.Sprintf("%s.html", base)
	}

	var template = &Template{
		FSBase: FSBase{
			Name:          name,
			Path:          filepath,
			Root:          root,
			Output:        output,
			Relative:      relative,
			Depth:         depth,
			RootDirectory: rootDir,
		},
	}

	template.Config = NewConfig(
		&template.FSBase,
	)

	return template, template.loadContent()
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

	content = bytes.TrimSpace(content)

	var (
		lines = bytes.Split(
			bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n")),
			[]byte("\n"),
		)
		contentIndex = 0
	)

loop:
	for i, line := range lines {
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("//")) {
			contentIndex = i
			break loop
		}

		var (
			trimmed = bytes.TrimPrefix(line, []byte("//"))
			parts   = bytes.SplitN(trimmed, []byte(":"), 2)
		)

		if len(parts) != 2 {
			contentIndex = i
			break loop
		}

		var (
			key   = strings.TrimSpace(string(parts[0]))
			value = strings.TrimSpace(string(parts[1]))
		)

		switch strings.ToLower(key) {
		case "title":
			t.Title = value
		case "next":
			t.Next = strings.Split(value, "/")
		case "previous":
			t.Previous = strings.Split(value, "/")
		default:
			contentIndex = i
			break loop
		}
	}

	t.Content = string(bytes.Join(
		lines[contentIndex:], []byte("\n"),
	))

	return nil
}
