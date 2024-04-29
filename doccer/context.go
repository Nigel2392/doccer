package doccer

import (
	"html/template"

	"github.com/Nigel2392/doccer/doccer/filesystem"
)

// MenuItem represents a menu item
type MenuItem struct {
	Name string `yaml:"name"`
	URL  string `yaml:"path"`
}

// Menu represents a menu
type Menu struct {
	Logo  string      `yaml:"logo"`
	Items []*MenuItem `yaml:"items"`
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

// Object represents the documentation object
func (c *Context) Object() filesystem.Object {
	return makeContextObject(c.object, c)
}
