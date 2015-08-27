package service

import "os"

// BuildTag and BuildDate should be replaced at compile time via Makefile:
//   BUILD_TAG = $(shell git log --pretty=format:'%h' -n 1)
//   BUILD_DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
//   install:
//     @go install -ldflags "-X main.buildTag=$(BUILD_TAG) -X main.buildDate=$(BUILD_DATE)"
//
// Your main.go should have:
//    var buildTag = "dev"
//    var buildDate = "0001-01-01T00:00:00Z"
//
//    func main() {
//        service.BuildTag = buildTag
//        service.BuildDate = buildDate
//    }
//

// BuildTag is the git hash of the build, or "dev" if no hash is provided
var BuildTag = "dev"

// BuildDate is the date that this was compiled, or zeroes if no date is provided
var BuildDate = "0001-01-01T00:00:00Z"

// Version is the base struct returned by the /version endpoint
type Version struct {
	BuildTag  string `json:"build"`
	BuildDate string `json:"buildDate"`
	Command   string `json:"command"`
}

// Hydrate will fill in the Build and Command fields of the Version struct given
func (v *Version) Hydrate() {
	v.BuildTag = BuildTag
	v.BuildDate = BuildDate
	v.Command = os.Args[0]
}
