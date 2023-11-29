package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

type Template struct {
	htmlTpl *template.Template
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
		},
	)
	w.Header().Set("Content-Type", "text/html; charset-utf-8")
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}
	// this creates an overhead if we are using large pages, remove buffer and copy statement to improve non-error-use-cases
	io.Copy(w, &buf)
}

func ParseFS(fs fs.FS, patterns ...string) (Template, error) {
	htmlTpl := template.New(patterns[0])
	htmlTpl = htmlTpl.Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", fmt.Errorf("csrfField NOT implemented")
		},
	})
	htmlTpl, err := htmlTpl.ParseFS(fs, patterns...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}
	return Template{
		htmlTpl: htmlTpl,
	}, nil
}

func Must(tpl Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return tpl
}
