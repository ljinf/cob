package binding

import (
	"encoding/xml"
	"errors"
	"net/http"
)

type xmlBinding struct {
}

func (x *xmlBinding) Name() string {
	return "xml"
}

func (x *xmlBinding) Bind(r *http.Request, obj interface{}) error {
	if r.Body == nil {
		return errors.New("request body is nil")
	}
	decoder := xml.NewDecoder(r.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}
