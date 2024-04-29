package doccer

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	text_template "text/template"

	"github.com/Nigel2392/doccer/doccer/filesystem"
	"github.com/Nigel2392/doccer/doccer/render"
)

var replacer = strings.NewReplacer(
	" ", "_",
	"-", "_",
	".", "_",
)

func clean_filename(filename string) string {
	filename = strings.ToLower(filename)
	return replacer.Replace(filename)
}

func IsLocal(filepath string) bool {
	if strings.HasPrefix(filepath, "http:") ||
		strings.HasPrefix(filepath, "https:") ||
		strings.HasPrefix(filepath, "ftp:") {
		return false
	}

	var u, err = url.Parse(filepath)
	return err != nil || u.Scheme == ""
}

type DoccerFS struct {
	fs.FS
}

func (d *DoccerFS) Open(name string) (f fs.File, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var doccerPath = filepath.Join(cwd, DOCCER_DIR, name)
	if _, err = os.Stat(doccerPath); err == nil {
		f, err = os.Open(doccerPath)
	} else {
		f, err = d.FS.Open(path.Join("assets", name))
	}

	return f, err
}

func ObjectURL(baseURL string, obj filesystem.Object, isServing bool) string {
	if obj == nil {
		return baseURL
	}

	var url string
	if isServing {
		url = obj.ServeURL()
	} else {
		url = obj.URL()
	}

	url = strings.Replace(url, "\\", "/", -1)
	url = path.Join(baseURL, url)

	if !strings.HasSuffix(url, "/") && obj.IsDirectory() {
		url += "/"
	}

	return url
}

func CopyDirectory(fileSys fs.FS, scrDir, dest string) error {
	dirs, err := fs.ReadDir(fileSys, scrDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		var (
			fSrcPath = path.Join(scrDir, dir.Name())
			fDstPath = filepath.Join(dest, dir.Name())
		)

		if dir.IsDir() {
			continue
		}

		fSrc, err := fileSys.Open(fSrcPath)
		if err != nil {
			return err
		}

		if _, err := os.Stat(fDstPath); err == nil {
			fmt.Printf("File %s already exists, skipping\n", fDstPath)
			continue
		}

		fDst, err := os.Create(fDstPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(fDst, fSrc)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildMapFunc[T *filesystem.TemplateDirectory | *filesystem.Template](context *Context, tree map[string]interface{}) func(string, T) bool {
	return func(key string, v T) bool {
		var value = any(v)
		if template, ok := value.(*filesystem.Template); ok {
			var (
				extension = path.Ext(template.Name)
				basename  = template.Name[0 : len(template.Name)-len(extension)]
			)

			basename = clean_filename(basename)

			tree[basename] = &contextObject{
				Object:  template,
				context: context,
			}

		} else if directory, ok := value.(*filesystem.TemplateDirectory); ok {
			var (
				newTree  = make(map[string]interface{})
				fnDirs   = buildMapFunc[*filesystem.TemplateDirectory](context, newTree)
				fnTpls   = buildMapFunc[*filesystem.Template](context, newTree)
				basename = directory.Name
			)

			basename = clean_filename(basename)

			tree[basename] = newTree
			newTree["root"] = &contextObject{
				Object:  directory,
				context: context,
			}
			directory.Subdirectories.ForEach(fnDirs)
			directory.Templates.ForEach(fnTpls)
		}

		return true
	}
}

func addTemplateContext(context *Context, t *filesystem.Template) {
	var (
		renderfn = render.Get(t.GetName())
		tpl      = text_template.New("content")
	)

	tpl = tpl.Funcs(context.Config.Instance.TemplateFuncs())
	tpl, err := tpl.Parse(string(t.Content))
	if err != nil {
		context.Content = template.HTML(fmt.Sprintf("Error: %s", err))
		return
	}

	var b bytes.Buffer
	err = tpl.ExecuteTemplate(&b, "content", context)
	if err != nil {
		context.Content = template.HTML(fmt.Sprintf("Error: %s", err))
		return
	}

	var b2 bytes.Buffer
	err = renderfn(&b2, b.Bytes())
	if err != nil {
		context.Content = template.HTML(fmt.Sprintf("Error: %s", err))
		return
	}

	context.Content = template.HTML(b2.String())
	context.object = t
}
