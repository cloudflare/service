# Service

The goal of this package is to quickly surface consistent and supportable JSON web services for small bits of business logic.

It is useful for two key scenarios:

1. Wrapping small pieces of business logic in a web service
2. Achieving consistency in the response structure

Features:
* Parsing of JSON inputs
* Rendering of JSON
* Errors rendered as JSON
* Pagination struct for consistent pagination by API consumers
* Automatic HTTP `OPTIONS`
* Automatic HTTP `HEAD`
* `glog`-style logging interface
* Sentry support if `os.Getenv("SENTRY_DSN")` is set
* Middleware capability (via [Negroni](https://github.com/codegangsta/negroni))
* `/_debug/profile/info.html` for web based profiling
* `/_debug/pprof` for pprof profiling
* `/_heartbeat` basic version info
* `/_version` endpoint that services can override with their own (i.e. to provide DB migration version information in addition to process version information)

## External dependencies

```go
go get github.com/codegangsta/negroni
go get github.com/getsentry/raven-go
go get github.com/mistifyio/negroni-pprof
go get github.com/unrolled/render
go get github.com/wblakecaldwell/profiler
go get github.com/coreos/go-oidc/jose
go get github.com/coreos/go-oidc/key
go get github.com/coreos/go-oidc/oidc
```

## Usage

```bash
go get github.com/cloudflare/service
```

Create a new project, and in your `main.go`:

```go
package main

import (
	"flag"

	"github.com/cloudflare/service"
)

func main() {
	// service/log requires that we call flag.Parse()
	flag.Parse()

	// Creates a new webservice, this holds all of the controllers
	ws := service.NewWebService()

	// Add a controller to the web service, a controller handles one path
	ws.AddWebController(HelloController())

	// Run the web service
	ws.Run(":8080")
}

```

In `example.go`:

```go
package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/cloudflare/service"
	"github.com/cloudflare/service/render"
)

// HelloController handles the /hello/{name} route
func HelloController() service.WebController {
	wc := service.NewWebController("/hello/{name}")

	// Add a handler for GET
	wc.AddMethodHandler(service.Get, HelloControllerGet)

	return wc
}

// HelloControllerGet handles GET for /hello/{name}
func HelloControllerGet(w http.ResponseWriter, req *http.Request) {
	// Get the path var for this controller
	name := mux.Vars(req)["name"]

	type HelloWorld struct {
		Hello string `json:"hello"`
	}

	hello := HelloWorld{Hello: name}

	render.JSON(w, http.StatusOK, hello)
}
```

Build and run the service, access via `http://localhost:8080/hello/world`.

Access `http://localhost:8080/` for a list of the endpoints available and their methods, call `http://localhost:8080/hello/world` to see the above endpoint working.

## Heartbeat and Versioning

To enable `/heartbeat` to echo the git hash of the current build and the timestamp of the current build you will need to use a Makefile to build your executable. You will need to adjust your main.go to support this:

Your `main.go`:

```go
package main

import "github.com/cloudflare/service"

var buildTag = "dev"
var buildDate = "0001-01-01T00:00:00Z"

func main() {
	service.BuildTag = buildTag
	service.BuildDate = buildDate
}
```

Your `Makefile`:

```Makefile
.PHONY: all build install

all: install

BUILD_TAG = $(shell git log --pretty=format:'%h' -n 1)
BUILD_DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	@go build -ldflags "-X main.buildTag=$(BUILD_TAG) -X main.buildDate=$(BUILD_DATE)"
	@echo "build complete"

install:
	@go install -ldflags "-X main.buildTag=$(BUILD_TAG) -X main.buildDate=$(BUILD_DATE)"
```

Using Make to build using `make build` or to install using `make install`. Those will build and install in the same way that `go build` or `go install` will, but simply wrap the `go` command to ensure that the build information is added.
