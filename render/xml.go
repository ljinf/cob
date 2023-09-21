package render

import (
	"encoding/xml"
	"net/http"
)

type Xml struct {
	Data interface{}
}

func (x *Xml) Render(w http.ResponseWriter, code int) error {
	x.WriteContentType(w)
	w.WriteHeader(code)
	return xml.NewEncoder(w).Encode(x.Data)
}

func (x *Xml) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/xml;charset=utf-8")
}
