package mocks

import "net/http"

func ExampleServer() {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, historyJSON := LoadFixture("../fixtures/fip/history_response.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(historyJSON)
	}
	server := Server(http.HandlerFunc(handler))
	defer server.Close()

	// Override and defer resetting the endpoint URL so that the mock
	// server is queried in the tests, rather than the actual server:
	// SetEndpointURL(server.URL)
	// defer ResetEndpointURL()
}
