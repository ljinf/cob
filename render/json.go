package render

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	Data interface{}
}

func (j *Json) Render(w http.ResponseWriter, code int) error {
	j.WriteContentType(w)
	w.WriteHeader(code)
	dataByte, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataByte)
	return err
}

func (j *Json) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/json;charset=utf-8")
}
