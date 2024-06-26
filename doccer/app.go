package doccer

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "embed"

	"github.com/Nigel2392/doccer/doccer/filesystem"
	"github.com/Nigel2392/doccer/doccer/hooks"
	"github.com/Nigel2392/typeutils/terminal"
	"gopkg.in/yaml.v3"
)

const DOCCER_DIR = ".doccer"
const MAX_MENU_ITEMS_DEPTH = 1

// FooterMenu represents the footer menu
var FooterMenu = &Menu{
	Items: []MenuItem{
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

type Doccer struct {
	// Config file path
	configPath string

	// Config
	config *Config

	// Registered features
	features map[string]Feature

	embedFS *DoccerFS
}

// NewDoccer creates a new doccer instance
func NewDoccer(embedFS fs.FS, configPath string) (*Doccer, error) {
	var doccer = &Doccer{
		configPath: configPath,
		features:   make(map[string]Feature),
		embedFS:    &DoccerFS{embedFS},
	}
	doccer.config = NewConfig(doccer)

	return doccer, nil
}

// FS returns the filesystem
func (d *Doccer) FS() *DoccerFS {
	return d.embedFS
}

// ParseArgs parses the arguments for the command
func (d *Doccer) ParseArgs(args []string) (err error) {
	if len(args) == 0 {
		return
	}

	var (
		fs  = flag.NewFlagSet("doccer", flag.ExitOnError)
		h   = hooks.Get[ParseArgHook]("parse_args")
		fns = make([]ParseFlagFn, 0, len(h))
	)

	for _, hook := range h {
		if fn := hook(d, fs); fn != nil {
			fns = append(fns, fn)
		}
	}

	err = fs.Parse(args)
	if err != nil {
		return
	}

	for _, fn := range fns {
		err = fn(d, fs)
		if err != nil {
			return
		}
	}

	return nil
}

// Load the configuration
func (d *Doccer) Load() error {
	// Load the config
	yamlConfig, err := os.ReadFile(d.configPath)
	if err != nil {
		return ErrNoConfig
	}

	// Unmarshal the config
	err = yaml.Unmarshal(yamlConfig, d.config)
	if err != nil {
		return err
	}

	err = d.config.Init()
	if err != nil {
		return err
	}

	var h = hooks.Get[FeatureHook]("register_features")
	for _, hook := range h {
		var feature = hook(d, d.config)
		if feature != nil {
			d.features[feature.ID()] = feature
		}
	}

	for _, feature := range d.config.Features {
		var f, ok = d.features[feature]
		if !ok {
			return fmt.Errorf("feature not found: %s", feature)
		}

		err = f.Init(d, d.config)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMenu returns the menu for the documentation
func (d *Doccer) BuildMenu(isServing bool) *Menu {
	var menu = &Menu{
		Items: make([]MenuItem, 0),
	}

	if d.config.Menu != nil && len(d.config.Menu.Items) > 0 {
		return d.buildMenu(d.config.Menu, d.config.RootDirectory, isServing)
	}

	if d.config.Menu != nil && d.config.Menu.Logo != "" {
		menu.Logo = d.config.Menu.Logo
	}

	if len(d.config.Menu.Items) == 0 {
		d.config.RootDirectory.Subdirectories.ForEach(func(key string, v *filesystem.TemplateDirectory) bool {
			menu.Items = append(menu.Items, MenuItem{
				Name: v.GetTitle(),
				URL:  ObjectURL(d.config.Server.BaseURL, v, isServing),
			})
			return true
		})

		d.config.RootDirectory.Templates.ForEach(func(key string, v *filesystem.Template) bool {
			menu.Items = append(menu.Items, MenuItem{
				Name: v.GetTitle(),
				URL:  ObjectURL(d.config.Server.BaseURL, v, isServing),
			})
			return true
		})
	}

	var h = hooks.Get[ConstructMenuHook]("construct_menu")
	for _, hook := range h {
		hook(d, menu)
	}

	return menu
}

func (d *Doccer) buildMenu(m *Menu, dir *filesystem.TemplateDirectory, isServing bool) *Menu {
	var menu = &Menu{
		Items: make([]MenuItem, 0),
	}

	menu.Logo = m.Logo
	menu.Items = d.buildMenuItems(m.Items, dir, isServing, 0)

	var h = hooks.Get[ConstructMenuHook]("construct_menu")
	for _, hook := range h {
		hook(d, menu)
	}

	return menu
}

func (d *Doccer) buildMenuItems(m []MenuItem, dir *filesystem.TemplateDirectory, isServing bool, depth int) []MenuItem {
	var items = make([]MenuItem, 0)
	for i, item := range m {
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
				item.Name = obj.GetTitle()
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

		if len(item.Items) > 0 {
			if depth > MAX_MENU_ITEMS_DEPTH {
				raise(fmt.Sprintf("menu item %s has too many levels: %d > %d", item.Name, i, MAX_MENU_ITEMS_DEPTH))
			}
			item.Items = d.buildMenuItems(item.Items, dir, isServing, depth+1)
		}

		items = append(items, MenuItem{
			Name:       item.Name,
			URL:        url,
			Items:      item.Items,
			Icon:       item.Icon,
			Classname:  item.Classname,
			Attributes: item.Attributes,
		})
	}

	return items
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

	context.Tree["root"] = &contextObject{
		Object:  d.config.RootDirectory,
		context: context,
	}

	var fnDirs = buildMapFunc[*filesystem.TemplateDirectory](context, context.Tree)
	var fnTpls = buildMapFunc[*filesystem.Template](context, context.Tree)

	d.config.RootDirectory.Subdirectories.ForEach(fnDirs)
	d.config.RootDirectory.Templates.ForEach(fnTpls)

	return context
}

var WidthRegex, _ = regexp.Compile(`width="([a-zA-Z0-9]+)"`)
var HeightRegex, _ = regexp.Compile(`height="([a-zA-Z0-9]+)"`)

func (d *Doccer) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"JSON": func(v interface{}) template.HTML {
			var b, err = json.MarshalIndent(v, "", "  ")
			if err != nil {
				return template.HTML(fmt.Sprintf("Error: %s", err))
			}
			return template.HTML(b)
		},
		"encodeZipped": func(v interface{}, elementId string) template.HTML {
			var b, err = json.Marshal(v)
			if err != nil {
				return template.HTML(fmt.Sprintf("Error: %s", err))
			}
			var buf bytes.Buffer
			var z = gzip.NewWriter(&buf)
			_, err = z.Write(b)
			if err != nil {
				return template.HTML(fmt.Sprintf("Error: %s", err))
			}
			err = z.Close()
			if err != nil {
				return template.HTML(fmt.Sprintf("Error: %s", err))
			}
			var data = buf.Bytes()
			return template.HTML(fmt.Sprintf("<script id=\"%s\" type=\"application/base64\">%s</script>", elementId, base64.StdEncoding.EncodeToString(data)))
		},
		"RenderHook": func(c *Context, hook string) template.HTML {
			var h = hooks.Get[RendererHook](hook)
			var html = make([]string, 0)
			for _, hook := range h {
				var renderer = hook(c)
				html = append(html, renderer.Render(c))
			}
			return template.HTML(strings.Join(html, "\n"))
		},
		"Env": func(key string) string {
			return os.Getenv(key)
		},
		"GetCurrentDate": func() string {
			return time.Now().Format("2006-01-02")
		},
		"GetTime": func() time.Time {
			return time.Now()
		},
		// Returns an icon which can be used in markdown files.
		// The regular "Icon" function can be used; but is not as good as it might mess up headings.
		"MarkdownIcon": func(name string, alt ...string) string {
			var altStr = name
			if len(alt) > 1 {
				altStr = strings.Join(alt, " ")
			} else {
				altStr = alt[0]
			}
			return fmt.Sprintf("![%s](%s)", altStr, d.AssetURL(path.Join(
				"static/bootstrap-icons",
				fmt.Sprintf("%s.svg", name),
			)))
		},
		"Icon": func(name string, sizing ...string) template.HTML {
			if name == "" {
				return template.HTML("")
			}

			var w, h string
			if len(sizing) == 1 {
				var wh = strings.Split(sizing[0], "x")
				if len(wh) == 2 {
					w = wh[0]
					h = wh[1]
				} else {
					w = sizing[0]
					h = sizing[0]
				}
			} else if len(sizing) > 1 {
				return template.HTML(fmt.Sprintf("Error: Icon sizing has too many arguments: %v", sizing))
			} else {
				w = "24"
				h = "24"
			}

			var svgPath = fmt.Sprintf("static/bootstrap-icons/%s.svg", name)
			var svg, err = fs.ReadFile(d.embedFS, svgPath)
			if err != nil {
				return template.HTML(fmt.Sprintf("Error rendering SVG: %s", err))
			}

			var svgStr = string(svg)
			svgStr = WidthRegex.ReplaceAllString(svgStr, fmt.Sprintf(`width="%s"`, w))
			svgStr = HeightRegex.ReplaceAllString(svgStr, fmt.Sprintf(`height="%s"`, h))

			return template.HTML(svgStr)
		},
		"Asset": func(name string) template.HTML {
			return template.HTML(d.AssetURL(name))
		},
	}
}

// Asset returns the asset path
func (d *Doccer) AssetURL(name string) string {

	if IsLocal(d.config.Server.StaticUrl) {
		var p = path.Join(d.config.Server.StaticUrl, name)
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return p
	}

	if strings.HasPrefix(name, "/") && strings.HasSuffix(d.config.Server.StaticUrl, "/") {
		name = strings.TrimPrefix(name, "/")
	}

	if !strings.HasSuffix(d.config.Server.StaticUrl, "/") && !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	return d.config.Server.StaticUrl + name + "?raw=true"
}

func (d *Doccer) Build() error {
	// Build the templates
	var (
		err  error
		last filesystem.Object
	)

	// Run all build hooks
	var h = hooks.Get[DoccerHook]("before_build")
	for _, hook := range h {
		err = hook(d)
		if err != nil {
			return err
		}
	}

	d.config.RootDirectory.ForEach(func(obj filesystem.Object) bool {
		last = obj

		var outputDir string

		// Skip directories
		if obj.IsDirectory() {
			var dir = obj.(*filesystem.TemplateDirectory)
			err = os.MkdirAll(dir.Output, 0755)
			if err != nil {
				err = fmt.Errorf("error creating directory %s: %s", dir.Output, err)
				return false
			}

			outputDir = filepath.Join(dir.Output, "index.html")
		} else {
			var t = obj.(*filesystem.Template)
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

	// Run all build hooks
	var ldHooks = hooks.Get[LoadHook]("after_build")
	for _, hook := range ldHooks {
		err = hook(d, d.config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Doccer) Init() error {
	var err = os.MkdirAll(DOCCER_DIR, 0755)
	if err != nil {
		return err
	}

	var createConfig = true
	if _, err := os.Stat(d.configPath); err == nil {
		createConfig = false

		ans, _ := terminal.RepeatAsk(
			"Config file already exists. Do you want to overwrite it? (y/n)",
			[]string{
				"y", "yes",
				"n", "no",
			},
			false,
		)

		if ans == "y" || ans == "yes" {
			createConfig = true
		}
	}

	if createConfig {
		var h = hooks.Get[LoadHook]("init_new_config")
		for _, hook := range h {
			err = hook(d, d.config)
			if err != nil {
				return err
			}
		}

		var b, err = yaml.Marshal(d.config)
		if err != nil {
			return err
		}

		err = os.WriteFile(d.configPath, b, 0644)
		if err != nil {
			return err
		}
	}

	// Create the templates directory
	err = os.MkdirAll(filepath.Join(DOCCER_DIR, "templates"), 0755)
	if err != nil {
		return err
	}

	// Create the static directory
	err = os.MkdirAll(filepath.Join(DOCCER_DIR, "static"), 0755)
	if err != nil {
		return err
	}

	// Run all init hooks
	var h = hooks.Get[DoccerHook]("init_new_project")
	for _, hook := range h {
		err = hook(d)
		if err != nil {
			return err
		}
	}

	// Copy the static files
	err = CopyDirectory(d.embedFS.FS, "assets/static", filepath.Join(DOCCER_DIR, "static"))
	if err != nil {
		return err
	}

	// Copy the templates
	return CopyDirectory(
		d.embedFS.FS,
		"assets/templates",
		filepath.Join(DOCCER_DIR, "templates"),
	)
}

// Serve serves the documentation
func (d *Doccer) Serve() error {
	var (
		serverConfig = d.config.Server
		addr         = fmt.Sprintf("%s:%d", serverConfig.Hostname, serverConfig.Port)
	)

	var h = hooks.Get[DoccerHook]("before_serve")
	for _, hook := range h {
		if err := hook(d); err != nil {
			return err
		}
	}

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
		http.ServeFile(w, r, path)
		return
	}

	var (
		baseUrl       = strings.TrimSuffix(d.config.Server.BaseURL, "/")
		hasBasePrefix = strings.HasPrefix(path, baseUrl)
	)
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

func (d *Doccer) renderObject(w io.Writer, obj filesystem.Object) error {
	var _, isServing = w.(http.ResponseWriter)
	var context = d.GetContext(isServing)

	var h = hooks.Get[func(*Doccer, *Context, filesystem.Object) error]("pre_render_object")
	for _, hook := range h {
		var err = hook(d, context, obj)
		if err != nil {
			return err
		}
	}

	if t, ok := obj.(*filesystem.Template); ok && !t.IsTextFile() {
		return t.Render(w, d.TemplateFuncs(), context)
	}

	// Serve the object
	if obj.IsDirectory() {
		var dir = obj.(*filesystem.TemplateDirectory)
		if dir.Index != nil {
			addTemplateContext(
				context, dir.Index,
			)
		} else {
			var tpl = &filesystem.Template{
				FSBase: filesystem.FSBase{
					Name:          "index.html",
					Path:          "index.html",
					Root:          dir.Root,
					Output:        "index.html",
					Relative:      "index.html",
					Depth:         dir.Depth,
					RootDirectory: d.config.RootDirectory,
				},
			}
			tpl.Config = filesystem.NewConfig(
				&tpl.FSBase,
			)

			var b = new(strings.Builder)
			dir.Subdirectories.ForEach(func(key string, v *filesystem.TemplateDirectory) bool {
				//var o = &contextObject{
				//	Object:  v,
				//	context: context,
				//}
				fmt.Fprintf(b, "<p><a href=\"%s\">", ObjectURL(d.config.Server.BaseURL, v, isServing))
				fmt.Fprint(b, v.GetTitle())
				fmt.Fprintf(b, "</a></p>\n")
				return true
			})

			dir.Templates.ForEach(func(key string, v *filesystem.Template) bool {
				//var o = &contextObject{
				//	Object:  v,
				//	context: context,
				//}
				fmt.Fprintf(b, "<p><a href=\"%s\">", ObjectURL(d.config.Server.BaseURL, v, isServing))
				fmt.Fprint(b, v.GetTitle())
				fmt.Fprintf(b, "</a></p>\n")
				return true
			})

			tpl.Content = b.String()
			addTemplateContext(
				context, tpl,
			)
		}

	} else {
		var t = obj.(*filesystem.Template)

		addTemplateContext(
			context, t,
		)
	}

	var err = d.config.Tpl.ExecuteTemplate(w, "base", context)
	if err != nil {
		return err
	}

	return nil
}
