package doccer

import (
	"html/template"
	"os"

	"github.com/Nigel2392/doccer/doccer/filesystem"
)

type (
	ServerConfig struct {
		Hostname    string `yaml:"hostname"`    // Hostname to use for the server
		Port        int    `yaml:"port"`        // Port to use for the server
		BaseURL     string `yaml:"base_url"`    // Base URL for the server
		StaticUrl   string `yaml:"static_url"`  // Static URL for assets
		StaticRoot  string `yaml:"static_root"` // Static root directory for assets
		PrivateKey  string `yaml:"private_key"` // Private key for the server
		Certificate string `yaml:"certificate"` // Certificate for the server
	}

	ProjectConfig struct {
		Name            string `yaml:"name"`       // Project name
		Version         string `yaml:"version"`    // Project version
		Repository      string `yaml:"repository"` // Repository URL
		InputDirectory  string `yaml:"input"`      // Documentation root directory
		OutputDirectory string `yaml:"output"`     // Output directory
	}

	Config struct {
		Server  ServerConfig           `yaml:"server"`  // Server configuration
		Project ProjectConfig          `yaml:"project"` // Project configuration
		Context map[string]interface{} `yaml:"context"` // Extra context for generating documentation
		Menu    *Menu                  `yaml:"menu"`    // Menu items

		Instance      *Doccer                       `yaml:"-"` // Doccer instance
		Tpl           *template.Template            `yaml:"-"` // HTML Template
		RootDirectory *filesystem.TemplateDirectory `yaml:"-"` // Root directory
	}
)

func NewConfig(instance *Doccer) *Config {
	return &Config{
		Context:  make(map[string]interface{}),
		Menu:     &Menu{},
		Instance: instance,
	}
}

// UnmarshalYAML unmarshals the config from YAML
func (c *Config) Init() error {
	if c.Project.Name == "" {
		raise("'project' is required")
	}

	if c.Project.Version == "" {
		raise("'version' is required")
	}

	if c.Project.InputDirectory == "" {
		raise("'input' is required")
	}

	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	if c.Server.BaseURL == "" {
		c.Server.BaseURL = "/"
	}

	if c.Project.OutputDirectory == "" {
		c.Project.OutputDirectory = "docs_output"
	}

	if c.Server.StaticUrl == "" {
		c.Server.StaticUrl = "/static"
	}

	if c.Context == nil {
		c.Context = make(map[string]interface{})
	}

	// Create the output directory
	var err = os.MkdirAll(c.Project.OutputDirectory, 0755)
	if err != nil {
		return err
	}

	// Check if the root directory exists
	if _, err := os.Stat(c.Project.InputDirectory); os.IsNotExist(err) {
		raise("input root does not exist")
	}

	var files = []string{
		"templates/footer.tmpl",
		"templates/navbar.tmpl",
		"templates/main.tmpl",
		"templates/head.tmpl",
		"templates/base.tmpl",
		// assetFile(c.Template),
	}

	// Create the template
	var tpl = template.New("base")

	tpl.Funcs(c.Instance.TemplateFuncs())

	tpl, err = tpl.ParseFS(c.Instance.embedFS, files...)
	if err != nil {
		return err
	}

	// Create the root directory
	var (
		inp = c.Project.InputDirectory
		out = c.Project.OutputDirectory
	)
	rootDirectory, err := filesystem.NewTemplateDirectory(nil, "", inp, inp, out, "", 0)
	if err != nil {
		return err
	}

	c.Tpl = tpl
	rootDirectory.Name = c.Project.Name
	rootDirectory.Root = inp
	rootDirectory.Output = out
	c.RootDirectory = rootDirectory

	return nil
}
