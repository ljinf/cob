package binding

import "net/http"

type Binding interface {
	Name() string
	Bind(r *http.Request, obj interface{}) error
}

var (
	JSON = jsonBinding{}
	XML  = xmlBinding{}
)
