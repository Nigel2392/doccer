package doccer

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"

	"github.com/Nigel2392/doccer/doccer/hooks"
)

type (
	Feature interface {
		ID() string
		Init(*Doccer, *Config) error
	}
	Renderer interface {
		Render(*Context) string
	}
	FeatureHook       func(*Doccer, *Config) Feature
	DoccerHook        func(*Doccer) error
	LoadHook          func(*Doccer, *Config) error
	ConstructMenuHook func(*Doccer, *Menu)
	RendererHook      func(*Context) Renderer
	ParseFlagFn       func(*Doccer, *flag.FlagSet) error
	ParseArgHook      func(d *Doccer, fs *flag.FlagSet) ParseFlagFn
)

type templatePaths []string // TemplatePath represents a path to a template

func TemplatePath(paths ...string) templatePaths {
	return paths
}

func (t templatePaths) Render(c *Context) string {
	var tpl = template.New("feature_template")
	tpl.Funcs(c.Config.Instance.TemplateFuncs())
	tpl, err := tpl.ParseFS(c.Config.Instance.embedFS, t...)
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

type Hook struct {
	HookName string
	Priority int
	Handlers []any
}

type HookFeature struct {
	Identifier string
	Hooks      []Hook
}

func (h *HookFeature) ID() string {
	return h.Identifier
}

func (h *HookFeature) Init(d *Doccer, c *Config) error {
	for _, hook := range h.Hooks {
		hooks.Register(
			hook.HookName,
			hook.Priority,
			hook.Handlers...,
		)
	}
	return nil
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
				"templates/hooks/navbar_menu_item.tmpl",
			)
		},
	)

	hooks.Register(
		"register_features", 0,
		func(d *Doccer, c *Config) Feature {
			return &HookFeature{
				Identifier: "search",
				Hooks: []Hook{
					{HookName: "render_navbar_content", Priority: -10, Handlers: []any{
						func(c *Context) Renderer {
							return TemplatePath(
								"templates/hooks/navbar_search.tmpl",
							)
						},
					}},
					//	{HookName: "after_build", Priority: -10, Handlers: []any{
					//		func(d *Doccer, c *Config) error {
					//
					//			return nil
					//		},
					//	}},
				},
			}
		},
	)

}
