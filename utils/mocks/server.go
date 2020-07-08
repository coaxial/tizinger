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
func Server(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

// LoadFixture is a helper for loading canned responses to use in handler
// functions.
func LoadFixture(path string) (length int, content []byte) {
	content, err := ioutil.ReadFile(path)
	length = len(content)
	if err != nil {
		logger.Error.Fatalf("Could not load %q: %v", path, err)
	}

	return length, content
}
