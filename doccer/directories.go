package doccer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Nigel2392/orderedmap"
)

// TemplateDirectory represents a directory of documentation templates
// This is used to generate a tree- like structure for documentation templates.
type TemplateDirectory struct {
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

	// Index template
	Index *Template

	// Subdirectories
	Subdirectories *orderedmap.Map[string, *TemplateDirectory]

	// Templates in the directory
	Templates *orderedmap.Map[string, *Template]

	// Depth in the directory tree
	Depth int
}

// NewTemplateDirectory creates a new template directory
func NewTemplateDirectory(name, root, path, output, relative string, depth int) (*TemplateDirectory, error) {

	if relative == "" && name != "" {
		relative = name
	}

	var dir = &TemplateDirectory{
		Name:           name,
		Root:           root,
		Path:           path,
		Output:         output,
		Relative:       relative,
		Subdirectories: orderedmap.New[string, *TemplateDirectory](),
		Templates:      orderedmap.New[string, *Template](),
		Depth:          depth,
	}

	var dirs, err = os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(dirs) == 0 && dir.Depth == 0 {
		return nil, ErrNoTemplates
	}

	// Sort based on isdir
	slices.SortFunc(dirs, func(i, j fs.DirEntry) int {
		var d1, d2 = i.IsDir(), j.IsDir()
		if d1 && !d2 {
			return -1
		} else if !d1 && d2 {
			return 1
		} else {
			return 0
		}
	})

	for _, d := range dirs {

		var (
			fPath = filepath.Join(path, d.Name())
			oPath = filepath.Join(output, d.Name())
			rPath = filepath.Join(relative, d.Name())
		)

		if d.IsDir() {
			var subDir, err = NewTemplateDirectory(d.Name(), root, fPath, oPath, rPath, dir.Depth+1)
			if err != nil {
				return nil, err
			}

			dir.Subdirectories.Set(subDir.Name, subDir)
		} else {
			var template, err = NewTemplate(d.Name(), root, fPath, oPath, rPath, dir.Depth+1)
			if err != nil {
				return nil, err
			}

			if isIndexFile(template.Name) {
				dir.Index = template
			} else {
				dir.Templates.Set(template.Name, template)
			}
		}

	}

	return dir, nil
}

func (d *TemplateDirectory) depthString() string {
	return strings.Repeat("  ", d.Depth)
}

// Format the directory as a string
func (d *TemplateDirectory) String() string {
	return fmt.Sprintf("Directory: %s", d.Relative)
}

// Format the directory for %v
func (d *TemplateDirectory) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%sTemplateDirectory{%s, %s}", d.depthString(), d.Name, d.Relative)
}

// GetName returns the name of the directory
func (d *TemplateDirectory) GetName() string {
	return d.Name
}

// IsDirectory returns true if the object is a directory
func (d *TemplateDirectory) IsDirectory() bool {
	return true
}

func (d *TemplateDirectory) ForEach(f func(Object) bool) bool {
	if !f(d) {
		return false
	}

	d.Subdirectories.ForEach(func(key string, v *TemplateDirectory) bool {
		return v.ForEach(f)
	})

	d.Templates.ForEach(func(key string, v *Template) bool {
		return f(v)
	})

	return true
}
func (d *TemplateDirectory) Traverse(fn func(*TemplateDirectory) (o Object, next bool, mayReturn bool)) (Object, bool) {
	var o, next, mayReturn = fn(d)
	if !next || mayReturn {
		return o, mayReturn
	}

	if o != nil {
		if d, ok := o.(*TemplateDirectory); ok {
			return d.Traverse(fn)
		} else if mayReturn {
			return o, true
		}
	}

	return nil, false
}

// Walk walks the directory tree
func (d *TemplateDirectory) Walk(parts []string) (Object, bool) {
	return d.Traverse(func(dir *TemplateDirectory) (Object, bool, bool) {
		if len(parts) == 0 {
			return dir, false, true
		}

		var (
			part     = parts[dir.Depth]
			next, ok = dir.Subdirectories.GetOK(part)
		)
		if ok {
			return next, true, false
		}

		n, ok := dir.Templates.GetOK(part)
		if ok {
			return n, true, true
		}

		return nil, false, false
	})
}

func (d *TemplateDirectory) URL() string {
	if strings.HasPrefix(d.Relative, "/") {
		return d.Relative
	}

	return fmt.Sprintf("/%s", d.Relative)
}

func (d *TemplateDirectory) ServeURL() string {
	return d.URL()
}
