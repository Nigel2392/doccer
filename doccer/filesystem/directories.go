package filesystem

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
	FSBase

	// Index template
	Index *Template

	// Subdirectories
	Subdirectories *orderedmap.Map[string, *TemplateDirectory]

	// Templates in the directory
	Templates *orderedmap.Map[string, *Template]
}

// NewTemplateDirectory creates a new template directory
func NewTemplateDirectory(rootDir *TemplateDirectory, name, root, path, output, relative string, depth int) (*TemplateDirectory, error) {

	if relative == "" && name != "" {
		relative = name
	}

	var dir = &TemplateDirectory{
		FSBase: FSBase{
			Name:          name,
			Root:          root,
			Path:          path,
			Output:        output,
			Relative:      relative,
			Depth:         depth,
			RootDirectory: rootDir,
		},
		Subdirectories: orderedmap.New[string, *TemplateDirectory](),
		Templates:      orderedmap.New[string, *Template](),
	}

	if rootDir == nil {
		rootDir = dir
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
			var subDir, err = NewTemplateDirectory(rootDir, d.Name(), root, fPath, oPath, rPath, dir.Depth+1)
			if err != nil {
				return nil, err
			}

			dir.Subdirectories.Set(subDir.Name, subDir)
		} else {
			var template, err = NewTemplate(rootDir, d.Name(), root, fPath, oPath, rPath, dir.Depth+1)
			if err != nil {
				return nil, err
			}

			if IsIndexFile(template.Name) {
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
	return d.Relative
}

// Format the directory for %v
func (d *TemplateDirectory) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%sTemplateDirectory{%s, %s}", d.depthString(), d.Name, d.Relative)
}

// GetName returns the name of the directory
func (d *TemplateDirectory) GetName() string {
	return d.Name
}

// GetTitle returns the title of the directory
func (t *TemplateDirectory) GetTitle() string {
	return t.Index.Title
}

// GetNext returns the next object in the directory
func (d *TemplateDirectory) GetNext() Object {
	return d.Index.GetNext()
}

// GetPrevious returns the previous object in the directory
func (d *TemplateDirectory) GetPrevious() Object {
	return d.Index.GetPrevious()
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
	if parts == nil {
		return nil, false
	}

	if len(parts) == 0 {
		return d, true
	}

	if len(parts) == 1 && d.Index != nil && parts[0] == d.Index.Name {
		return d, true
	}

	var (
		part     = parts[0]
		next, ok = d.Subdirectories.GetOK(part)
	)
	if ok {
		return next.Walk(parts[1:])
	}

	n, ok := d.Templates.GetOK(part)
	if ok {
		return n, true
	}

	return nil, false
}

func (d *TemplateDirectory) URL() string {
	var url string
	if strings.HasPrefix(d.Relative, "/") {
		url = d.Relative
	} else {
		url = fmt.Sprintf("/%s", d.Relative)
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url
}

func (d *TemplateDirectory) ServeURL() string {
	return d.URL()
}
