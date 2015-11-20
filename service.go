package service

import (
	"fmt"
	"net/http"
	gopprof "net/http/pprof"
	"os"
	"sort"

	"github.com/codegangsta/negroni"
	raven "github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	pprof "github.com/mistifyio/negroni-pprof"
	"github.com/wblakecaldwell/profiler"

	"github.com/cloudflare/service/log"
	"github.com/cloudflare/service/render"
)

const (
	root string = `/`

	// VersionRoute is the path to the version information endpoint
	VersionRoute string = `/_version`
)

// EndPoint describes an endpoint that exists on this web service
type EndPoint struct {
	URL     string `json:"href"`
	Methods string `json:"methods"`
}

// EndPoints is a slice of all endpoints on this web service
type EndPoints []EndPoint

func (slice EndPoints) Len() int {
	return len(slice)
}

func (slice EndPoints) Less(i, j int) bool {
	return slice[i].URL < slice[j].URL
}

func (slice EndPoints) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// WebService represents a web server with a collection of controllers
type WebService struct {
	controllers []WebController
}

// NewWebService provides a way to create a new blank WebService
func NewWebService() WebService {
	ws := WebService{}

	// Heartbeat controller (echoes the default version info)
	heartbeatController := NewWebController("/_heartbeat")
	heartbeatController.AddMethodHandler(Get,
		func(w http.ResponseWriter, r *http.Request) {
			v := Version{}
			v.Hydrate()
			render.JSON(w, http.StatusOK, v)
		},
	)
	ws.AddWebController(heartbeatController)

	return ws
}

// AddWebController allows callees to add their controller.
// Note: The order in which the controllers are added is the order in which the
// routes will be applied.
func (ws *WebService) AddWebController(wc WebController) {
	ws.controllers = append(ws.controllers, wc)
}

// BuildRouter collects all of the controllers, wires up the routes and returns
// the resulting router
func (ws *WebService) BuildRouter() *mux.Router {
	// Router
	//
	// StrictSlash forces the routes to be applied literally...
	// i.e. Route /foo/ with requests /foo will redirect to /foo/
	// and route /bar with requests to /bar/ will redirect to /bar
	r := mux.NewRouter().StrictSlash(true)

	// Controllers
	rootSeen := false
	versionSeen := false
	links := EndPoints{}
	for _, wc := range ws.controllers {
		if !rootSeen && wc.Route == root {
			rootSeen = true
		}

		if !versionSeen && wc.Route == VersionRoute {
			versionSeen = true
		}

		// Add the handler for a route, and rate-limit it using throttle
		r.Handle(
			wc.Route,
			http.HandlerFunc(GetHandler(wc)),
		)

		links = append(links, EndPoint{URL: wc.Route, Methods: wc.GetAllowedMethods()})
	}

	// Profiling handlers
	// XXX: should we add them using the public api too?
	r.HandleFunc("/_profiler/info.html", profiler.MemStatsHTMLHandler)
	links = append(links, EndPoint{URL: "/_profiler/info.html", Methods: "GET"})
	r.HandleFunc("/_profiler/info", profiler.ProfilingInfoJSONHandler)
	r.HandleFunc("/_profiler/start", profiler.StartProfilingHandler)
	r.HandleFunc("/_profiler/stop", profiler.StopProfilingHandler)

	r.HandleFunc("/_debug/pprof/", http.HandlerFunc(gopprof.Index))
	links = append(links, EndPoint{URL: "/_debug/pprof", Methods: "GET"})
	r.HandleFunc("/_debug/pprof/cmdline", http.HandlerFunc(gopprof.Cmdline))
	r.HandleFunc("/_debug/pprof/profile", http.HandlerFunc(gopprof.Profile))
	r.HandleFunc("/_debug/pprof/symbol", http.HandlerFunc(gopprof.Symbol))

	if !versionSeen {
		// If detailed version info is not provided, we echo the default
		// This allows services to provide their own extended version info, i.e.
		// database versioning as well as process versioning
		r.HandleFunc(VersionRoute, func(w http.ResponseWriter, r *http.Request) {
			v := Version{}
			v.Hydrate()
			render.JSON(w, http.StatusOK, v)
		})
		links = append(links, EndPoint{URL: VersionRoute, Methods: "GET"})
	}

	// The last routes are the NotFound routes as we want to return JSON.
	//
	// This handles / on it's own, and we should only do this if no other
	// route already registered /
	if !rootSeen {
		sort.Sort(links)
		r.HandleFunc(root, func(w http.ResponseWriter, r *http.Request) {
			render.JSON(w, http.StatusOK, links)
		})
	}

	// This is a wildcard route and will greedily consume all remaining routes
	r.HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		render.Error(
			w,
			http.StatusNotFound,
			fmt.Errorf("/%s not found", mux.Vars(r)["path"]),
		)
	})

	return r
}

// Run collects all of the controllers, wires up the routes and starts the server
func (ws *WebService) Run(addr string) {
	n := negroni.New()

	// Middleware for net/http/pprof
	n.Use(pprof.Pprof())

	// Send errors to sentry if the SENTRY_DSN environment variable is set
	r := ws.BuildRouter()
	hfn := r.ServeHTTP
	if os.Getenv("SENTRY_DSN") != "" {
		hfn = raven.RecoveryHandler(hfn)
	}

	// Apply mux routes
	n.UseHandlerFunc(hfn)

	// Wrap ListenAndServe and start the server
	log.Fatal(http.ListenAndServe(addr, n))
}
