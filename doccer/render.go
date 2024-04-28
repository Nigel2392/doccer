package doccer

import (
	"io"
	"path"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var renderMap = make(map[string]func(io.Writer, []byte, *Doccer) error)

func init() {
	renderMap["html"] = renderRaw
	renderMap["md"] = renderMarkdown
	renderMap["markdown"] = renderMarkdown
}

func GetFileRenderer(filename string) func(io.Writer, []byte, *Doccer) error {
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

func renderRaw(w io.Writer, content []byte, d *Doccer) error {
	content = append([]byte("<pre>"), content...)
	content = append(content, []byte("</pre>")...)
	_, err := w.Write(content)
	return err
}

func renderMarkdown(w io.Writer, content []byte, d *Doccer) error {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	return md.Convert(content, w)
}
