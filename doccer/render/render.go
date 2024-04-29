package render

import (
	"io"
	"path"
)

var (
	renderMap = make(map[string]func(io.Writer, []byte) error)
)

func init() {
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
