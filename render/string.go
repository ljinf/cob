package render

import (
	"fmt"
	"github.com/ljinfu/cob/internal/bytesconv"
	"net/http"
)

type String struct {
	Format string
	Data   []interface{}
}

func (s *String) Render(w http.ResponseWriter,code int) error {
	s.WriteContentType(w)
	w.WriteHeader(code)
	if len(s.Data) > 0 {
		_, err := fmt.Fprintf(w, s.Format, s.Data...)
		return err
	}
	_, err := w.Write(bytesconv.StringToByte(s.Format))
	return err
}

func (s *String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/plain;charset=utf-8")
}
