package service

import "net/http"

// HTTP Methods
const (
	Options = iota
	Head
	Post
	Get
	Put
	Patch
	Delete
	Connect
	Trace
)

// IsMethod returns true if the int value matches one of the iota values for a
// HTTP method
func IsMethod(m int) bool {
	switch m {
	default:
		return false
	case Options:
		return true
	case Head:
		return true
	case Post:
		return true
	case Get:
		return true
	case Put:
		return true
	case Patch:
		return true
	case Delete:
		return true
	case Connect:
		return true
	case Trace:
		return true
	}
}

// GetMethodName returns the upper-cased method, i.e. GET for a given method
// int value
func GetMethodName(m int) string {
	switch m {
	default:
		return ""
	case Options:
		return "OPTIONS"
	case Head:
		return "HEAD"
	case Post:
		return "POST"
	case Get:
		return "GET"
	case Put:
		return "PUT"
	case Patch:
		return "PATCH"
	case Delete:
		return "DELETE"
	case Connect:
		return "CONNECT"
	case Trace:
		return "TRACE"
	}
}

// GetMethodID returns an int value for a valid HTTP method name (upper-cased)
func GetMethodID(method string) int {
	switch method {
	default:
		return 0
	case "OPTIONS":
		return Options
	case "HEAD":
		return Head
	case "POST":
		return Post
	case "GET":
		return Get
	case "PUT":
		return Put
	case "PATCH":
		return Patch
	case "DELETE":
		return Delete
	case "CONNECT":
		return Connect
	case "TRACE":
		return Trace
	}
}

// GetHTTPMethod returns the method ID for the method in a HTTP request
func GetHTTPMethod(req *http.Request) int {
	return GetMethodID(req.Method)
}
