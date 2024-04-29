// Title: Setting up Templates
// Previous: configuration.md

# Setting up Templates

Doccer uses Go's `html/template` package to render the HTML pages.

The templates are stored in the `templates` directory, typically found in `./.doccer/templates`.

Each markdown file will be embedded into the `base.tmpl` template.

Markdown files will have access to the following variables by default:

- `.Config` - The configuration object.

  - `.Project`           - The project metadata.

    - `.Name`            - Project name
    - `.Version`         - Project version
    - `.Repository`      - Repository URL
    - `.InputDirectory`  - Documentation root directory
    - `.OutputDirectory` - Output directory

  - `.Server`            - The server configuration.

    - `.Hostname`        - Hostname to use for the server
    - `.Port`            - Port to use for the server
    - `.BaseURL`         - Base URL for the server
    - `.StaticUrl`       - Static URL for assets
    - `.StaticRoot`      - Static root directory for assets
    - `.PrivateKey`      - Private key for the server
    - `.Certificate`     - Certificate for the server
    
- `.Object`           - The current object being worked on.
  - `.GetName`        - A function to get the name of the object.
  - `.GetTitle`       - A function to get the title of the object.
  - `.GetNext`        - A function to get the next object.
  - `.GetPrevious`    - A function to get the previous object.

- `.Menu`   - The menu items defined in the configuration file.
  (Otherwise automatically generated).

- `.Ctx`    - The custom context variables defined in the configuration file.
  This is a map of string to interface.

- `Asset`   - A function to prefix your staticfiles correctly.

Your markdown files can individually configure themselves. Think of changing titles, setting up related pages etc.

These configurations are made at the top of the markdown file.

An example:
    
```markdown
// Title: This is my README.md
// Next: my_folder/next_page.md

# This is my README.md
...

```
