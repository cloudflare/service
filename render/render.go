package render

import (
	"net/http"

	"github.com/unrolled/render"
)

var r = render.New(
	render.Options{
		IndentJSON: true,
	},
)

// Error will write a given error to the http.ResponseWriter as JSON
// and set the HTTP status.
func Error(w http.ResponseWriter, status int, err error) {
	type ErrorJS struct {
		Message string `json:"error"`
	}

	r.JSON(w, status, ErrorJS{Message: err.Error()})
}

// JSON will write a given interface{} to the http.ResponseWriter as JSON
// and set the HTTP status.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	r.JSON(w, status, v)
}
