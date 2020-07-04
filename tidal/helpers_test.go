package tidal

import (
	"github.com/coaxial/tizinger/utils/logger"
)

// originalURL keeps track of the live API endpoint
var originalURL = baseURL

// SetBaseURL is for testing, so that a mock server can be used instead of
// the live one, and arbitrary responses or failures can be served as needed.
func SetBaseURL(url string) {
	baseURL = url
	logger.Trace.Printf("base URL overridden to %q", url)
}

// ResetBaseURL resets the endpointURL to its original value, hitting the
// live API
func ResetBaseURL() {
	baseURL = originalURL
	logger.Trace.Printf("base URL reset to %q", originalURL)
}

// SetToken is a wrapper for the fetchToken method so that the token
// manifest URL can be overridden.
func SetToken(url string) (err error) {
	originalURL := manifestURL
	logger.Trace.Printf("tokens manifest URL overridden to %q", url)
	manifestURL = url
	err = setToken()
	defer func() {
		manifestURL = originalURL
		logger.Trace.Printf("tokens manifest URL reset to %q", originalURL)
	}()
	return err
}
