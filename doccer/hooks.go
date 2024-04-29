package doccer

import (
	"html/template"

	"github.com/Nigel2392/doccer/doccer/hooks"
)

type (
	Renderer interface {
		Render(*Context) template.HTML
	}
	DoccerHook        func(*Doccer) error
	ConstructMenuHook func(*Doccer, *Menu)
)

func init() {
	hooks.Register(
		"construct_menu", -1,
		func(d *Doccer, m *Menu) {
			var projectRootItem = &MenuItem{
				Name:      d.config.Project.Name,
				URL:       d.config.Server.BaseURL,
				Classname: "navbar-title",
			}

			m.Items = append([]*MenuItem{projectRootItem}, m.Items...)
		},
	)
}
