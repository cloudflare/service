package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/cloudflare/service/render"
)

// WebController describes the HTTP method handlers for a given route.
// Create a WebController with service.NewController(route)
type WebController struct {
	Route    string
	handlers map[int]func(w http.ResponseWriter, req *http.Request)
	allowed  string
}

// NewWebController creates a new controller for a given route
func NewWebController(route string) WebController {
	wc := WebController{}

	wc.handlers = make(map[int]func(w http.ResponseWriter, req *http.Request))

	wc.Route = route

	return wc
}

// GetAllowedMethods returns a comma-delimited string of HTTP methods allowed by
// this controller. This is determined by examining which methods have handlers
// assigned to them.
func (wc *WebController) GetAllowedMethods() string {
	if wc.allowed != "" {
		return wc.allowed
	}

	allowed := []string{}

	for k := range wc.handlers {
		allowed = append(allowed, GetMethodName(k))
	}

	wc.allowed = strings.Join(allowed, ",")

	return wc.allowed
}

// AddMethodHandler adds a HTTP handler to a given HTTP method
func (wc *WebController) AddMethodHandler(m int, h func(w http.ResponseWriter, req *http.Request)) {
	if !IsMethod(m) {
		log.Fatalf("Method iota %d not recognised", m)
	}

	if m == Options {
		log.Fatal("Cannot set OPTIONS, this is provided for you")
	}

	if m == Head {
		log.Fatal("Cannot set HEAD, this is provided for you")
	}

	wc.handlers[m] = h
	wc.allowed = ""
}

// GetMethodHandler returns the appropriate method handler for the request or a
// Method Not Allowed handler
func (wc *WebController) GetMethodHandler(m int) func(w http.ResponseWriter, req *http.Request) {
	if m == Options {
		return func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Allow", wc.GetAllowedMethods())
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusOK)
		}
	}

	if m == Head {
		return func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Allow", wc.GetAllowedMethods())
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusOK)
		}
	}

	if h, ok := wc.handlers[m]; ok {
		return h
	}

	return func(w http.ResponseWriter, req *http.Request) {
		allowed := wc.GetAllowedMethods()
		w.Header().Set("Allow", allowed)

		render.Error(
			w,
			http.StatusMethodNotAllowed,
			fmt.Errorf("405 Method Not Allowed. Allowed: %s", allowed),
		)
	}
}

// GetHandler returns a global handler for this route, to be used by the server
// mux
func GetHandler(
	wc WebController,
) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		wc.GetMethodHandler(GetHTTPMethod(req))(w, req)
	}
}
