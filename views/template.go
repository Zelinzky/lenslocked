package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-faster/errors"
	"github.com/gorilla/csrf"

	"lenslocked/appctx"
	"lenslocked/models"
)

type Template struct {
	htmlTpl *template.Template
}

type public interface {
	Public() string
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any, errs ...error) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}
	errMsgs := errMessages(errs...)
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return appctx.User(r.Context())
			},
			"errors": func() []string {
				return errMsgs
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
		"currentUser": func() (*models.User, error) {
			return nil, fmt.Errorf("currentUser NOT implemented")
		},
		"errors": func() []string {
			return nil
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
func errMessages(errs ...error) []string {
	var messages []string
	for _, err := range errs {
		var pubErr public
		if errors.As(err, &pubErr) {
			messages = append(messages, pubErr.Public())
		} else {
			fmt.Println(err)
			messages = append(messages, "Something went wrong.")
		}
	}
	return messages
}
