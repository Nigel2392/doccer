// Title: Configuring Doccer
// Next: templates.md
// Previous: README.md

# Configuring Doccer

Doccer can be configured using a configuration file.

The configuration file is named `doccer.yaml` and should be placed in the root of the project.


Typically the configuration file will have the following mappings:

## Project

The `project` section contains metadata about the project.

It can define the following labels:

- `name` - The name of the project.
- `version` - The version of the project.
- `repository` - The repository URL.
- `input` - The input directory for the markdown files.
- `output` - The output directory for the generated HTML files.

```yaml
project:
  name: "Getting Started"
  version: "1.0.0"
  repository: "https://github.com/Nigel2392/doccer"
  input: "./docs_src"
  output: "./docs"
```

## Server

The `server` section contains the configuration for the local server.

It also handles the base URL for the documentation and the static URL for the assets (even when externally hosted).

It can define the following labels:

- `base_url` - The base URL for the documentation.
- `static_url` - The static URL for the assets.
- `hostname` - The hostname to use for the server.
- `port` - The port to use for the server.
- `private_key` - The private key file for the server.
- `certificate` - The certificate file for the server.

```yaml
server:
  base_url: "/doccer/"
  static_url: "https://github.com/Nigel2392/doccer/blob/main/assets"
  hostname: "localhost"
  port: 8080
  # private_key: "path/to/private_key"
  # certificate: "path/to/public_key"
```

## Menu

The `menu` section contains the configuration for the menu items.

Menu items can define the following labels:

 - `name` - The name of the menu item.
 - `path` - The relative path to the markdown file.
 - `classname` - A custom class name for the menu item.
 - `icon` - A Bootstrap icons icon-name.
 - `attributes` - Custom attributes for the menu item. Cannot use: `href`, `class`.
 - `items` - A list of sub-menu items with the same structure.
   Sub-menu items can not define `items` themselves.

The path is relative to the `input` directory.

```yaml
menu:
  items:
    - name: "Configuration"
      path: "configuration.md"
    - name: "Templates"
      path: "templates.md"
      icon: "file-earmark-text" # Bootstrap icon library included.
      attributes:
        data-my-attribute: "value"
      items:
        - name: "Setting up Templates"
          path: "setup_templates.md"
        - name: "Customizing Templates"
          path: "customizing_templates.md"
```

## Custom Context

The `context` section contains custom context variables that can be accessed in the markdown files.

These can be accessed using `{{ "{{ index .Ctx.your_custom_key }}" }}`.

```yaml
context:
  # Add custom context variables here
  # These can be accessed in the markdown files using {{ "{{ index .Ctx.key }}" }}
  # key: value
```
