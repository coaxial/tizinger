package mocks

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/coaxial/tizinger/utils/logger"
)

// Server mocks the response.
func Server(handler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

// LoadFixture is for loading canned responses to use in handler functions.
func LoadFixture(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Fatalf("Could not load %s: %v", path, err)
	}

	return content
}
