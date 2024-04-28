package doccer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	text_template "text/template"
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

// isIndexFile returns true if the file is an index file
func isIndexFile(name string) bool {
	return strings.HasPrefix(name, "index.") ||
		strings.HasPrefix(strings.ToLower(name), "readme.")
}

func assetFile(relative string) string {
	if relative == "" {
		panic("asset file path is required")
	}

	if !IsLocal(relative) {
		return relative
	}

	if filepath.IsAbs(relative) {
		return relative
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return relative
	}

	var assetPath = filepath.Join(workingDir, "doccer", relative)
	if _, err := os.Stat(assetPath); err == nil {
		return assetPath
	}

	var executable = os.Args[0]
	return filepath.Join(filepath.Dir(executable), "assets", relative)
}

func ObjectURL(baseURL string, obj Object, isServing bool) string {
	if obj == nil {
		return baseURL
	}

	if isServing {
		return path.Join(baseURL, obj.ServeURL())
	}
	return path.Join(baseURL, obj.URL())
}

func buildMapFunc[T *TemplateDirectory | *Template](context *Context, tree map[string]interface{}) func(string, T) bool {
	return func(key string, v T) bool {
		var value = any(v)
		if template, ok := value.(*Template); ok {
			var (
				extension = path.Ext(template.Name)
				basename  = template.Name[0 : len(template.Name)-len(extension)]
			)

			basename = clean_filename(basename)

			tree[basename] = &contextObject{
				Object:  template,
				context: context,
			}

		} else if directory, ok := value.(*TemplateDirectory); ok {
			var (
				newTree  = make(map[string]interface{})
				fnDirs   = buildMapFunc[*TemplateDirectory](context, newTree)
				fnTpls   = buildMapFunc[*Template](context, newTree)
				basename = directory.Name
			)

			basename = clean_filename(basename)

			tree[basename] = newTree
			directory.Subdirectories.ForEach(fnDirs)
			directory.Templates.ForEach(fnTpls)
		}

		return true
	}
}

func addTemplateContext(context *Context, t *Template) {
	if context.Title == "" {
		context.Title = t.GetName()
	}
	context.object = t
	var (
		renderfn = GetFileRenderer(t.GetName())
		tpl      = text_template.New("content")
	)

	tpl = tpl.Funcs(context.GetFuncs())
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
	err = renderfn(&b2, b.Bytes(), context.Config.Instance)
	if err != nil {
		context.Content = template.HTML(fmt.Sprintf("Error: %s", err))
		return
	}

	context.Content = template.HTML(b2.String())

}
