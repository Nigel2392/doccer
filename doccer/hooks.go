package doccer

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Nigel2392/doccer/doccer/hooks"
)

type (
	Renderer interface {
		Render(*Context) string
	}
	DoccerHook           func(*Doccer) error
	LoadHook             func(*Doccer, *Config) error
	RegisterTemplateHook func(*Doccer) string
	ConstructMenuHook    func(*Doccer, *Menu)
	RendererHook         func(*Context) Renderer
)

type TemplatePath string // TemplatePath represents a path to a template

func (t TemplatePath) Render(c *Context) string {
	var tpl = template.New("feature_template")
	tpl.Funcs(c.Config.Instance.TemplateFuncs())
	tpl, err := tpl.ParseFS(c.Config.Instance.embedFS, string(t))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var buf = new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(buf, "feature_template", c); err != nil {
		return err.Error()
	}
	return buf.String()
}

func init() {
	// hooks.Register(
	// 	"construct_menu", -1,
	// 	func(d *Doccer, m *Menu) {
	// 		var projectRootItem = MenuItem{
	// 			Name:      d.config.Project.Name,
	// 			URL:       d.config.Server.BaseURL,
	// 			Classname: "navbar-title",
	// 		}
	//
	// 		m.Items = append([]MenuItem{projectRootItem}, m.Items...)
	// 	},
	// )
	hooks.Register(
		"render_navbar_content", 0,
		func(c *Context) Renderer {
			return TemplatePath(
				"templates/hooks/navbar_menu.tmpl",
			)
		},
	)

	hooks.Register(
		"render_navbar_content", -10,
		func(c *Context) Renderer {
			return TemplatePath(
				"templates/hooks/navbar_search.tmpl",
			)
		},
	)
}
