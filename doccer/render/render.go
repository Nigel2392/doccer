package render

import (
	"io"
	"path"
	"strings"

	"github.com/Nigel2392/doccer/doccer/hooks"
)

var (
	renderMap = make(map[string]func(io.Writer, []byte) error)
)

func init() {
	// Check if the file is a javascript, css or webassembly file
	hooks.Register("is_text_file", 99, func(name string, content []byte) bool {
		name = strings.ToLower(name)
		return !(strings.HasSuffix(name, ".js") ||
			strings.HasSuffix(name, ".css") ||
			strings.HasSuffix(name, ".wasm") ||
			strings.HasSuffix(name, ".wat"))
	})

	Register("css", renderRaw)
	Register("js", renderRaw)
	Register("wasm", renderRaw)
	Register("wat", renderRaw)
	Register("html", renderRaw)

	Register("md", renderMarkdown)
	Register("markdown", renderMarkdown)
}

func Register(filetype string, render func(io.Writer, []byte) error) {
	renderMap[filetype] = render
}

func For(filename string) func(io.Writer, []byte) error {
	var ext = path.Ext(filename)
	if ext == "" {
		return renderRaw
	}

	ext = ext[1:]
	if render, ok := renderMap[ext]; ok {
		return render
	}

	return renderRaw
}
