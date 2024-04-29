package doccer

import (
	"encoding/json"
	"html/template"

	"github.com/Nigel2392/doccer/doccer/filesystem"
)

// MenuItem represents a menu item
type MenuItem struct {
	Name       string            `yaml:"name"`
	URL        string            `yaml:"path"`
	Classname  string            `yaml:"classname"`
	Icon       string            `yaml:"icon"`
	Attributes map[string]string `yaml:"attributes"`
	Items      []MenuItem        `yaml:"items"`
}

func (m MenuItem) Copy() MenuItem {
	var items = make([]MenuItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = item.Copy()
	}
	return MenuItem{
		Name:       m.Name,
		URL:        m.URL,
		Classname:  m.Classname,
		Icon:       m.Icon,
		Attributes: m.Attributes,
		Items:      items,
	}
}

// Menu represents a menu
type Menu struct {
	Logo  string     `yaml:"logo"`
	Items []MenuItem `yaml:"items"`
}

func makeContextObject(object filesystem.Object, context *Context) filesystem.Object {
	if object == nil {
		return nil
	}
	if _, ok := object.(*contextObject); ok {
		return object
	}
	return &contextObject{
		Object:  object,
		context: context,
	}
}

type contextObject struct {
	// The current object
	Object filesystem.Object

	// The current context
	context *Context
}

func (c *contextObject) GetName() string {
	return c.Object.GetName()
}

func (c *contextObject) GetTitle() string {
	return c.Object.GetTitle()
}

func (c *contextObject) GetNext() filesystem.Object {
	var next = c.Object.GetNext()
	if next == nil {
		return nil
	}
	return makeContextObject(next, c.context)
}

func (c *contextObject) GetPrevious() filesystem.Object {
	var previous = c.Object.GetPrevious()
	if previous == nil {
		return nil
	}
	return makeContextObject(previous, c.context)
}

func (c *contextObject) IsDirectory() bool {
	return c.Object.IsDirectory()
}

func (c *contextObject) URL() string {
	return ObjectURL(c.context.Config.Server.BaseURL, c.Object, c.context.isServing)
}

func (c *contextObject) ServeURL() string {
	return ObjectURL(c.context.Config.Server.BaseURL, c.Object, true)
}

func (c *contextObject) String() string {
	return c.Object.String()
}

func (c *contextObject) MarshalJSON() ([]byte, error) {
	var obj = map[string]interface{}{
		"name":   c.GetName(),
		"title":  c.GetTitle(),
		"url":    c.URL(),
		"is_dir": c.IsDirectory(),
	}

	if c.IsDirectory() {
		var dir = c.Object.(*filesystem.TemplateDirectory)
		if dir.Index != nil {
			obj["content"] = dir.Index.Content
		} else {
			obj["content"] = ""
		}
	} else {
		obj["content"] = c.Object.(*filesystem.Template).Content
	}
	return json.Marshal(obj)
}

// Context represents the context for the documentation
type Context struct {

	// Flag to indicate if the server is serving over HTTP
	isServing bool

	// Current object being rendered
	object filesystem.Object

	// The current configuration
	Config *Config

	// The current content
	Content template.HTML

	// Context from the config
	Ctx map[string]interface{}

	// Menu items
	Menu *Menu

	// Footer
	Footer *Menu

	// The directory tree
	Tree map[string]interface{}
}

// IsServing returns true if the server is serving over HTTP
func (c *Context) IsServing() bool {
	return c.isServing
}

// Object represents the documentation object
func (c *Context) Object() filesystem.Object {
	return makeContextObject(c.object, c)
}

// FlatObjectList returns a flat list of objects
func (c *Context) FlatObjectList() []filesystem.Object {
	var list []filesystem.Object
	var fn = func(obj filesystem.Object) bool {
		list = append(list, makeContextObject(obj, c))
		return true
	}
	c.Config.RootDirectory.ForEach(fn)
	return list
}
