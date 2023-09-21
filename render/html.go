package render

import (
	"github.com/ljinfu/cob/internal/bytesconv"
	"html/template"
	"net/http"
)

type HTMLRender struct {
	Template *template.Template
}

type HTML struct {
	Data       interface{}
	Name       string
	Template   *template.Template
	IsTemplate bool
}

func (h *HTML) Render(w http.ResponseWriter, code int) error {
	h.WriteContentType(w)
	w.WriteHeader(code)
	if h.IsTemplate {
		return h.Template.ExecuteTemplate(w, h.Name, h.Data)
	}
	_, err := w.Write(bytesconv.StringToByte(h.Data.(string)))
	return err
}

func (h *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html;charset=utf-8")
}
