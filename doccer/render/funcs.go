package render

import (
	"io"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func renderRaw(w io.Writer, content []byte) error {
	_, err := w.Write(content)
	return err
}

//	func renderPre(w io.Writer, content []byte) error {
//		_, err := w.Write([]byte("<pre>"))
//		if err != nil {
//			return err
//		}
//		_, err = w.Write(content)
//		if err != nil {
//			return err
//		}
//		_, err = w.Write([]byte("</pre>"))
//		return err
//	}

func renderMarkdown(w io.Writer, content []byte) error {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
		),
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
