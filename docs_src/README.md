// Title: Doccer DocGen
// Next: configuration.md

<center>
    <img src="{{ Asset "static/logo.svg" }}" alt="Doccer">
</center>

{{ MarkdownIcon "terminal" "" }} {{ .Config.Project.Name }}
=======================

A simple documentation generator for your project.

Easily build static HTML pages from your markdown files.

## {{ MarkdownIcon "download" "" }} Installation

```bash
go install {{ index .Ctx.InstallURL }}@latest
```

## {{ MarkdownIcon "question-lg" "" }} Usage

```bash
doccer init  # Initialize a new skeleton for the documentation.
doccer serve # Serve the documentation with a local server.
doccer build # Build the documentation.
```

