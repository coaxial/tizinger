package mocks

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/coaxial/tizinger/utils/logger"
)

// Server spins up a new HTTP server to return canned responses. The handler
// function processes the requests and eventually returns a fixture/canned
// data.
func Server(handler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

// LoadFixture is a helper for loading canned responses to use in handler
// functions.
func LoadFixture(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Fatalf("Could not load %q: %v", path, err)
	}

	return content
}
