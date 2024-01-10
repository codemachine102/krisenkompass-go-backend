package multitemplate

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"text/template"
)

var (
	layouts       = make(map[string]*template.Template)
	pages         = make(map[string]*template.Template)
	templatesPath = "../../templates/"
)

func AddLayout(name string, includes ...string) {
	// TODO: Remove slash prefix

	for i := 0; i < len(includes); i++ {
		includes[i] = templatesPath + includes[i]
	}

	layouts[name] = template.Must(template.ParseFiles(includes...))
}

func AddTemplate(name string, layout string, includes ...string) {
	// TODO: Remove slash prefix

	l, ok := layouts[layout]
	if !ok {
		panic("Layout not defined.")
	}

	for i := 0; i < len(includes); i++ {
		includes[i] = templatesPath + includes[i]
	}
	pages[name], _ = template.Must(l.Clone()).ParseFiles(includes...)
}

func RenderHTML(w http.ResponseWriter, name string, data interface{}) {
	_, ok := pages[name]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	buf := &bytes.Buffer{}
	err := Render(buf, name, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		buf.WriteTo(w)
	}
}

func Render(w io.Writer, name string, data interface{}) error {
	_, ok := pages[name]
	if !ok {
		return errors.New("not found") // TODO: Add error types.
	}

	err := pages[name].Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}
