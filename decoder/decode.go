package decoder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrContentTypeUndefined is returned when the request does not include the
	// Content-Type header.
	ErrContentTypeUndefined = fmt.Errorf("Content-Type is undefined")

	// ErrDecoderNotImplemented is returned if the Content-Type does not match
	// one of the defined decoders, i.e.
	//    "application/json" => jsonDecode
	//    "application/xml" => undefined and this error is return
	ErrDecoderNotImplemented = fmt.Errorf("Decoding is not yet implement")
)

// Decode will ready the body of the HTTP request and attempt to unmarshall the
// content into the supplied interface. If the content-type of the request is
// not one that matches a known decoder, then an error will be thrown
func Decode(req *http.Request, v interface{}) error {
	contentType := getContentType(req)
	switch contentType {
	case "application/json":
		return jsonDecode(req, v)
	case "":
		return ErrContentTypeUndefined
	default:
		return ErrDecoderNotImplemented
	}
}

func getContentType(req *http.Request) (contentType string) {
	contentType = strings.TrimSpace(req.Header.Get("Content-Type"))

	return
}

func jsonDecode(req *http.Request, v interface{}) error {
	defer req.Body.Close()

	return json.NewDecoder(req.Body).Decode(&v)
}
