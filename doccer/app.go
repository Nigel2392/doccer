package doccer

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// FooterMenu represents the footer menu
var FooterMenu = &Menu{
	Items: []*MenuItem{
		{Name: "View Doccer on GitHub", URL: "https://github.com/Nigel2392/doccer"},
	},
}

// Raise a panic with a message
// Prefixes the message with "doccer: "
func raise(message string) {
	panic(
		fmt.Sprintf("doccer: %s", message),
	)
}

// Object represents a documentation object
// It might be a directory or file.
type Object interface {
	String() string
	GetName() string
	IsDirectory() bool
	ServeURL() string
	URL() string
}

// Walker represents an object which can traverse through a path.
// It is used to find a specific object in the documentation tree.
type Walker interface {
	Walk(parts []string) (object Object, ok bool)
}

type Doccer struct {
	// Config file path
	configPath string

	// Config
	config *Config
}

// NewDoccer creates a new doccer instance
func NewDoccer(configPath string) (*Doccer, error) {
	var doccer = &Doccer{
		configPath: configPath,
	}

	// Load the config
	yamlConfig, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the config
	var config = NewConfig(doccer)
	err = yaml.Unmarshal(yamlConfig, config)
	if err != nil {
		return nil, err
	}

	doccer.config = config
	err = doccer.config.Init()
	if err != nil {
		return nil, err
	}

	return doccer, nil
}

// GetMenu returns the menu for the documentation
func (d *Doccer) BuildMenu(isServing bool) *Menu {
	var menu = &Menu{
		Items: make([]*MenuItem, 0),
	}

	if d.config.Menu != nil && len(d.config.Menu.Items) > 0 {
		return d.buildMenu(d.config.Menu, d.config.RootDirectory, isServing)
	}

	if d.config.Menu != nil && d.config.Menu.Logo != "" {
		menu.Logo = d.config.Menu.Logo
	}

	if len(d.config.Menu.Items) == 0 {
		d.config.RootDirectory.Subdirectories.ForEach(func(key string, v *TemplateDirectory) bool {
			menu.Items = append(menu.Items, &MenuItem{
				Name: v.Name,
				URL:  ObjectURL(d.config.Server.BaseURL, v, isServing),
			})
			return true
		})

		d.config.RootDirectory.Templates.ForEach(func(key string, v *Template) bool {
			menu.Items = append(menu.Items, &MenuItem{
				Name: v.Name,
				URL:  ObjectURL(d.config.Server.BaseURL, v, isServing),
			})
			return true
		})
	}

	return menu
}

func (d *Doccer) buildMenu(m *Menu, dir *TemplateDirectory, isServing bool) *Menu {
	var menu = &Menu{
		Items: make([]*MenuItem, 0),
	}

	menu.Logo = m.Logo
	for i, item := range m.Items {
		var parts = strings.Split(item.URL, "/")
		if len(parts) == 1 && parts[0] == "" {
			parts = []string{}
		}

		var url string = item.URL
		if !strings.HasPrefix(url, d.config.Server.BaseURL) {
			url = path.Join(d.config.Server.BaseURL, url)
		}
		if IsLocal(item.URL) {
			var obj, ok = dir.Walk(parts)
			if !ok {
				raise(fmt.Sprintf("menu item not found: %s", item.URL))
			}

			if item.Name == "" {
				item.Name = obj.GetName()
			}

			url = ObjectURL(d.config.Server.BaseURL, obj, isServing)
		} else {
			if item.Name == "" {
				item.Name = item.URL
			}
		}

		if url == "" {
			raise(fmt.Sprintf("menu item %d has no URL: %s", i, item.Name))
		}

		menu.Items = append(menu.Items, &MenuItem{
			Name: item.Name,
			URL:  url,
		})
	}

	return menu
}

// GetContext returns the context for the documentation
func (d *Doccer) GetContext(isServing bool) *Context {
	var context = &Context{
		isServing: isServing,
		Ctx:       d.config.Context,
		Tree:      make(map[string]interface{}),
		Menu:      d.BuildMenu(isServing),
		Footer:    FooterMenu,
		Config:    d.config,
	}

	var fnDirs = buildMapFunc[*TemplateDirectory](context, context.Tree)
	var fnTpls = buildMapFunc[*Template](context, context.Tree)

	d.config.RootDirectory.Subdirectories.ForEach(fnDirs)
	d.config.RootDirectory.Templates.ForEach(fnTpls)

	return context
}

func (d *Doccer) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"GetCurrentDate": func() string {
			return time.Now().Format("2006-01-02")
		},

		"GetTime": func() time.Time {
			return time.Now()
		},

		"Asset": func(name string) template.HTML {

			if IsLocal(d.config.Server.StaticUrl) {
				var p = path.Join(d.config.Server.StaticUrl, name)
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
				return template.HTML(p)
			}

			if strings.HasPrefix(name, "/") && strings.HasSuffix(d.config.Server.StaticUrl, "/") {
				name = strings.TrimPrefix(name, "/")
			}

			if !strings.HasSuffix(d.config.Server.StaticUrl, "/") && !strings.HasPrefix(name, "/") {
				name = "/" + name
			}

			return template.HTML(d.config.Server.StaticUrl + name + "?raw=true")

		},
	}
}

func (d *Doccer) Build() error {
	// Build the templates
	var (
		err  error
		last Object
	)
	d.config.RootDirectory.ForEach(func(obj Object) bool {
		last = obj

		var outputDir string

		// Skip directories
		if obj.IsDirectory() {
			var dir = obj.(*TemplateDirectory)
			err = os.MkdirAll(dir.Output, 0755)
			if err != nil {
				err = fmt.Errorf("error creating directory %s: %s", dir.Output, err)
				return false
			}

			outputDir = filepath.Join(dir.Output, "index.html")
		} else {
			var t = obj.(*Template)
			outputDir = t.Output
		}

		var b bytes.Buffer
		err = d.renderObject(&b, obj)
		if err != nil {
			err = fmt.Errorf("error rendering %s: %s", obj.GetName(), err)
			return false
		}

		// Write the template to the output directory
		err = os.WriteFile(outputDir, b.Bytes(), 0644)
		if err != nil {
			err = fmt.Errorf("error writing %s: %s", outputDir, err)
			return false
		}
		return err == nil
	})
	if err != nil {
		return fmt.Errorf("error building %s: %s", last.GetName(), err)
	}
	return nil
}

// Serve serves the documentation
func (d *Doccer) Serve() error {
	var (
		serverConfig = d.config.Server
		addr         = fmt.Sprintf("%s:%d", serverConfig.Hostname, serverConfig.Port)
	)
	if serverConfig.PrivateKey != "" && serverConfig.Certificate != "" {
		fmt.Printf("Serving documentation on https://%s\n", addr)
		return http.ListenAndServeTLS(addr, serverConfig.Certificate, serverConfig.PrivateKey, d)
	}
	fmt.Printf("Serving documentation on http://%s\n", addr)
	return http.ListenAndServe(addr, d)
}

// ServeHTTP serves the documentation as a handler
func (d *Doccer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		parts []string
		path  = r.URL.Path
	)

	if IsLocal(d.config.Server.StaticUrl) && strings.HasPrefix(path, d.config.Server.StaticUrl) {
		path = filepath.Clean(strings.TrimPrefix(path, d.config.Server.StaticUrl))
		path = assetFile(path)
		http.ServeFile(w, r, path)
		return
	}

	if strings.HasSuffix(path, "favicon.ico") {
		var icon = assetFile("favicon.ico")
		http.ServeFile(w, r, icon)
		return
	}

	var hasBasePrefix = strings.HasPrefix(path, d.config.Server.BaseURL)
	if !hasBasePrefix && path != "/" {
		http.NotFound(w, r)
		return
	} else if !hasBasePrefix && path == "/" {
		http.Redirect(w, r, d.config.Server.BaseURL, http.StatusFound)
		return
	}
	path = strings.TrimPrefix(path, d.config.Server.BaseURL)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		parts = []string{}
	} else {
		parts = strings.Split(path, "/")
	}

	// Walk the directory
	var obj, ok = d.config.RootDirectory.Walk(parts)
	if !ok {
		fmt.Println("Not found", parts)
		http.NotFound(w, r)
		return
	}

	var err = d.renderObject(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (d *Doccer) renderObject(w io.Writer, obj Object) error {
	var _, isServing = w.(http.ResponseWriter)
	var context = d.GetContext(isServing)

	// Serve the object
	if obj.IsDirectory() {
		var dir = obj.(*TemplateDirectory)

		// dir.Subdirectories.ForEach(func(key string, v *TemplateDirectory) bool {
		// 	fmt.Fprintf(w, "<p><a href=\"/%s\">", v.Relative)
		// 	fmt.Fprintf(w, ".%s%s", string(filepath.Separator), v.GetName())
		// 	fmt.Fprintf(w, "</a></p>\n")
		// 	return true
		// })

		if dir.Index != nil {
			addTemplateContext(
				context, dir.Index,
			)
		} else {
			var tpl = &Template{
				Name:     "index.html",
				Path:     "index.html",
				Root:     dir.Root,
				Output:   "index.html",
				Relative: "index.html",
				Depth:    dir.Depth,
			}
			var b bytes.Buffer
			dir.Subdirectories.ForEach(func(key string, v *TemplateDirectory) bool {
				fmt.Fprintf(&b, "<p><a href=\"/%s\">", ObjectURL(d.config.Server.BaseURL, v, isServing))
				fmt.Fprintf(&b, ".%s%s", string(filepath.Separator), v.GetName())
				fmt.Fprintf(&b, "</a></p>\n")
				return true
			})

			dir.Templates.ForEach(func(key string, v *Template) bool {
				fmt.Fprintf(&b, "<p><a href=\"/%s\">", ObjectURL(d.config.Server.BaseURL, v, isServing))
				fmt.Fprintf(&b, ".%s%s", string(filepath.Separator), v.GetName())
				fmt.Fprintf(&b, "</a></p>\n")
				return true
			})

			tpl.Content = b.Bytes()
			addTemplateContext(
				context, tpl,
			)
		}

		err := d.config.Tpl.ExecuteTemplate(w, "base", context)
		if err != nil {
			return err
		}

	} else {
		var t = obj.(*Template)

		addTemplateContext(
			context, t,
		)

		err := d.config.Tpl.ExecuteTemplate(w, "base", context)
		if err != nil {
			return err
		}
	}

	return nil
}
