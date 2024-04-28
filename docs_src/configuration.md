// Title: Configuring Doccer
// Next: templates

# Configuring Doccer

Doccer can be configured using a configuration file.

The configuration file is named `doccer.yaml` and should be placed in the root of the project.


Typically the configuration file will have the following mappings:

The `project` section contains metadata about the project.

```yaml
project:
  name: "Getting Started"
  version: "1.0.0"
  repository: "https://github.com/Nigel2392/doccer"
  input: "./docs_src"
  output: "./docs"
```

The `server` section contains the configuration for the local server.

It also handles the base URL for the documentation and the static URL for the assets (even when externally hosted).

```yaml

server:
  base_url: "/doccer/"
  static_url: "https://github.com/Nigel2392/doccer/blob/main/assets"
  hostname: "localhost"
  port: 8080
  # private_key: "path/to/private_key"
  # certificate: "path/to/public_key"
```

The `menu` section contains the configuration for the menu items.

Menu items must have a `name` and a `path`.

The path is relative to the `input` directory.

```yaml
menu:
  items:
    - name: "Configuration"
      path: "configuration.md"
    - name: "Templates"
      path: "templates.md"
```

The `context` section contains custom context variables that can be accessed in the markdown files.

These can be accessed using `{{ index .Ctx.your_custom_key }}`.

```yaml
context:
  # Add custom context variables here
  # These can be accessed in the markdown files using {{ index .Ctx.key }}
  # key: value
```
