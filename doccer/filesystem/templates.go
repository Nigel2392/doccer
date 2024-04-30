package filesystem

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	text_template "text/template"

	"github.com/Nigel2392/doccer/doccer/render"
)

// Template represents a documentation template
type Template struct {
	// Base filesystem object
	FSBase `json:",inline"`

	// Template configuration
	Config `json:",inline"`

	// Template content
	Content       string `json:"content"`
	isTextFile    bool
	canBeTemplate bool
	loaded        bool
}

// Format the template for %v
func (t *Template) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%sTemplate: %s", t.depthString(), t.Name)
}

// IsDirectory returns true if the object is a directory
func (d *Template) IsDirectory() bool {
	return false
}

// IsTextFile returns true if the file is a text file
func (t *Template) IsTextFile() bool {
	return t.isTextFile
}

// NewSimpleTemplate creates a new template from just a name, path, output and directory.
func NewDirectoryChild(dir *TemplateDirectory, name string, content []byte) (*Template, error) {
	var (
		relative = path.Join(dir.Relative, name)
		template = &Template{
			FSBase: FSBase{
				Name:          name,
				Path:          path.Join(dir.Path, name),
				Output:        filepath.Join(dir.Output, name),
				Root:          dir.Root,
				Relative:      relative,
				Depth:         dir.Depth + 1,
				RootDirectory: dir,
			},
		}
		config = NewConfig(
			&template.FSBase,
		)
	)

	template.Config = config

	return template, template.loadContent(content)
}

// NewTemplate creates a new template
func NewTemplate(rootDir *TemplateDirectory, name, root, filepath, output, relative string, depth int) (*Template, error) {
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

	// Load the template content
	var content, err = os.ReadFile(template.Path)
	if err != nil {
		return nil, err
	}

	return template, template.loadContent(content)
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
func (t *Template) loadContent(content []byte) error {
	// Check if the file is a text file
	t.isTextFile = isTextFile(t.Name, content)
	t.canBeTemplate = isValidUTF8(content)

	// Setup the output file.
	// This can only be done now - we need to know if the file is a text file
	if t.isTextFile && !strings.HasSuffix(t.Output, ".html") {
		var (
			ext  = path.Ext(t.Output)
			base = t.Output[:len(t.Output)-len(ext)]
		)

		t.Output = fmt.Sprintf("%s.html", base)
	}

	if t.isTextFile {

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
	} else {
		t.Content = string(content)
	}

	return nil
}

// Render the template
func (t *Template) Render(w io.Writer, funcs template.FuncMap, context interface{}) error {
	var renderfn = render.For(t.GetName())

	if !t.loaded && t.canBeTemplate {
		var tpl = text_template.New("content")

		tpl = tpl.Funcs(funcs)
		tpl, err := tpl.Parse(t.Content)
		if err != nil {
			return err
		}

		var b bytes.Buffer
		err = tpl.ExecuteTemplate(&b, "content", context)
		if err != nil {
			return err
		}

		t.Content = b.String()
		t.loaded = true
	}

	return renderfn(w, []byte(t.Content))
}
