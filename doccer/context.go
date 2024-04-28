package doccer

import (
	"html/template"
	"maps"
	"reflect"
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

type contextObject struct {
	// The current object
	Object Object

	// The current context
	context *Context
}

func (c *contextObject) GetName() string {
	return c.Object.GetName()
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

	// Title of the current documentation page
	Title string

	// Current object being rendered
	object Object

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

	// Pagination
	Previous Object
	Next     Object
}

// Object represents the documentation object
func (c *Context) Object() Object {
	if _, ok := c.object.(*contextObject); ok {
		return c.object
	}

	return &contextObject{
		Object:  c.object,
		context: c,
	}
}

// GetFuncs returns the template functions
func (c *Context) GetFuncs() template.FuncMap {
	var m = template.FuncMap{
		"Set": func(fieldname string, value any) string {
			var rValue = reflect.ValueOf(c)
			if rValue.Kind() == reflect.Ptr {
				rValue = rValue.Elem()
			}
			var rField = rValue.FieldByName(fieldname)
			if rField.IsValid() {
				rField.Set(reflect.ValueOf(value))
			} else {
				c.Ctx[fieldname] = value
			}
			return ""
		},
		"Get": func(fieldname string) any {
			var rValue = reflect.ValueOf(c)
			if rValue.Kind() == reflect.Ptr {
				rValue = rValue.Elem()
			}
			var rField = rValue.FieldByName(fieldname)
			if rField.IsValid() {
				return rField.Interface()
			}
			return c.Ctx[fieldname]
		},
	}
	var funcs = c.Config.Instance.TemplateFuncs()

	maps.Copy(m, funcs)

	return m
}
