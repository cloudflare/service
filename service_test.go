package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type errorResponse struct {
	Error string `json:"error"`
}

var (
	server *httptest.Server
	base   string
)

func startServer(ws WebService) {
	server = httptest.NewServer(ws.BuildRouter())
	base = server.URL
}

func createDefaultWS() WebService {
	VersionRoute = "/customVersionRoute"
	HeartbeatRoute = "/customHeartbeatRoute"
	BuildDate = "buildDate"
	BuildTag = "buildTag"

	return NewWebService()
}

func TestHasHeartbeatRouteByDefault(t *testing.T) {
	startServer(createDefaultWS())

	request, _ := http.NewRequest("GET", base+HeartbeatRoute, nil)
	res, _ := http.DefaultClient.Do(request)

	assertIsDefaultVersionResponse(t, res)
}

func TestHasDefaultVersionRouteIfNoneIsRegistered(t *testing.T) {
	startServer(createDefaultWS())

	request, _ := http.NewRequest("GET", base+VersionRoute, nil)
	res, _ := http.DefaultClient.Do(request)

	assertIsDefaultVersionResponse(t, res)
}

func TestAutomaticallyProvidesHeadAndOptions(t *testing.T) {
	ws := createDefaultWS()
	route := "/dummyRoute"
	ws.AddWebController(basicControllerForMethods(route, []int{Get, Post}))
	startServer(ws)

	for _, method := range []string{"OPTIONS", "HEAD"} {
		request, _ := http.NewRequest(method, base+route, nil)
		res, _ := http.DefaultClient.Do(request)

		assertResponseAllowsMethods(t, res, []string{"GET", "HEAD", "OPTIONS", "POST"})
	}
}

func TestProvides404Response(t *testing.T) {
	var responseError = errorResponse{}
	startServer(createDefaultWS())

	route := "/foobar"
	request, _ := http.NewRequest("GET", base+route, nil)
	res, _ := http.DefaultClient.Do(request)

	assertStatusCodeIs(t, res, http.StatusNotFound)

	json.NewDecoder(res.Body).Decode(&responseError)
	if expected := route + " not found"; responseError.Error != expected {
		t.Errorf("Got unexpected response body. Got: %s allowed: %s", responseError.Error, expected)
	}
}

func TestGivesMethodNotAllowed(t *testing.T) {
	var responseError = errorResponse{}

	ws := createDefaultWS()
	route := "/dummyRoute"
	ws.AddWebController(basicControllerForMethods(route, []int{Get, Post}))
	startServer(ws)

	request, _ := http.NewRequest("PUT", base+route, nil)
	res, _ := http.DefaultClient.Do(request)

	assertStatusCodeIs(t, res, http.StatusMethodNotAllowed)
	allowed := "GET,HEAD,OPTIONS,POST"
	if got := res.Header.Get("Allow"); got != allowed {
		t.Errorf("Allow header should be set. Got: %s allowed: %s", got, allowed)
	}

	json.NewDecoder(res.Body).Decode(&responseError)
	expected := "405 Method Not Allowed. Allowed: " + allowed
	if responseError.Error != expected {
		t.Errorf("Got unexpected response body. Got: %s allowed: %s", responseError.Error, expected)
	}
}

func TestCanOverrideVersionEndpoint(t *testing.T) {
	ws := createDefaultWS()
	ws.AddWebController(basicControllerForMethods(VersionRoute, []int{Get}))
	startServer(ws)

	request, _ := http.NewRequest("GET", base+VersionRoute, nil)
	res, _ := http.DefaultClient.Do(request)

	assertResponseBodyIs(t, res, "dummy for GET")
}

func TestCanOverrideRootEndpoint(t *testing.T) {
	route := "/"
	ws := createDefaultWS()
	ws.AddWebController(basicControllerForMethods(route, []int{Get}))
	startServer(ws)

	request, _ := http.NewRequest("GET", base+route, nil)
	res, _ := http.DefaultClient.Do(request)

	assertResponseBodyIs(t, res, "dummy for GET")
}

func TestResponsesFromAControllerAreOkay(t *testing.T) {
	route := "/foobar"
	ws := createDefaultWS()
	methods := []int{Get, Post}
	ws.AddWebController(basicControllerForMethods(route, methods))
	startServer(ws)

	for _, m := range []int{Get, Post} {
		request, _ := http.NewRequest(GetMethodName(m), base+route, nil)
		res, _ := http.DefaultClient.Do(request)

		assertStatusCodeIs(t, res, http.StatusOK)
		assertResponseBodyIs(t, res, "dummy for "+GetMethodName(m))
	}
}

func TestCannotSetOptionsMethod(t *testing.T) {
	testCannotSetMethodForController(t, Options, "TestCannotSetOptionsMethod")
}

func TestCannotSetHeadMethod(t *testing.T) {
	testCannotSetMethodForController(t, Options, "TestCannotSetHeadMethod")
}

func testCannotSetMethodForController(t *testing.T, method int, testName string) {
	// Only run the failing part in a different subprocess
	if os.Getenv("BE_CRASHER") == "1" {
		basicControllerForMethods("/foo", []int{method})
		return
	}

	// Start the actual test in a different subprocess
	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	stdout, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Check the log fatal message
	gotBytes, _ := ioutil.ReadAll(stdout)
	got := string(gotBytes)
	expected := fmt.Sprintf("Cannot set %s, this is provided for you", GetMethodName(method))
	if got[len(got)-len(expected)-1:len(got)-1] != expected {
		t.Fatalf("Unexpected log fatal. Got %s expected %s", got[len(got)-len(expected)-1:len(got)-1], expected)
	}

	// Check that the program exited
	err := cmd.Wait()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Process ran with err %v, want exit status 1", err)
	}
}

func basicControllerForMethods(route string, methods []int) WebController {
	controller := NewWebController(route)
	for _, method := range methods {
		controller.AddMethodHandler(method, dummyHandlerWithResponse("dummy for "+GetMethodName(method)))
	}
	return controller
}

func dummyHandlerWithResponse(output string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(output))
	}
}

func assertResponseBodyIs(t *testing.T, res *http.Response, expected string) {
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if string(content) != expected {
		t.Errorf("Got unexpected response body. Got: %s allowed: %s", string(content), expected)
	}
}

func assertResponseAllowsMethods(t *testing.T, res *http.Response, allowedMethods []string) {
	assertStatusCodeIs(t, res, 200)

	tests := map[string]string{
		"Content-Length": "0",
		"Allow":          strings.Join(allowedMethods, ","),
	}

	for headerKey, expected := range tests {
		val := res.Header.Get(headerKey)
		if val != expected {
			t.Errorf("Header %s was different: got %s expected %s", headerKey, val, expected)
		}
	}
}

func assertIsDefaultVersionResponse(t *testing.T, res *http.Response) {
	var version Version

	assertStatusCodeIs(t, res, 200)

	json.NewDecoder(res.Body).Decode(&version)
	tests := map[string]string{
		BuildDate: version.BuildDate,
		BuildTag:  version.BuildTag,
	}

	for expected, got := range tests {
		if expected != got {
			t.Errorf("Property was different: got %s expected %s", got, expected)
		}
	}
}

func assertStatusCodeIs(t *testing.T, res *http.Response, expected int) {
	if res.StatusCode != expected {
		t.Errorf("Status code was different: got %d expected %d", res.StatusCode, expected)
	}
}
