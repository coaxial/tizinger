package fip

// originalURL keeps track of the fip.fr endpoint hitting the live API
var originalURL = endpointURL

// SetEndpointURL is for testing, so that a mock server can be used instead of
// the live one, and arbitrary responses or failures can be served as needed.
func SetEndpointURL(url string) {
	endpointURL = url
}

// ResetEndpointURL resets the endpointURL to its original value, hitting the
// live fip.fr API
func ResetEndpointURL() {
	endpointURL = originalURL
}
