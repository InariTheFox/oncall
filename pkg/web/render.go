package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/sync/singleflight"
)

const (
	DEV  = "development"
	PROD = "production"
)

var (
	Env = DEV
)

func Renderer(dir, leftDelim, rightDelim string) Middleware {
	var devEnvGr singleflight.Group
	fs := os.DirFS(dir)
	t, err := compileTemplates(fs, leftDelim, rightDelim)
	if err != nil {
		panic(fmt.Sprintf("Renderer: %w", err))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := FromContext(r.Context())
			ctx.template = t

			if Env == DEV {
				tt, err, _ := devEnvGr.Do("dev", func() (any, error) {
					return compileTemplates(fs, leftDelim, rightDelim)
				})
				if err != nil {
					panic("Context.HTML:" + err.Error())
				}
				ctx.template = tt.(*template.Template)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func compileTemplates(filesystem fs.FS, leftDelim, rightDelim string) (*template.Template, error) {
	t := template.New("")
	t.Delims(leftDelim, rightDelim)
	err := fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return nil // skip unreadable or erroneous filesystem items
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".html" && ext != ".tmpl" {
			return nil
		}
		data, err := fs.ReadFile(filesystem, path)
		if err != nil {
			return err
		}
		basename := path[:len(path)-len(ext)]
		_, err = t.New(basename).Parse(string(data))
		return err
	})
	return t, err
}
