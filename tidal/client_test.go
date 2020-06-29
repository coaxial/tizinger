package tidal

import (
	"net/http"
	"testing"

	"github.com/coaxial/tizinger/utils/mocks"
	"github.com/stretchr/testify/assert"
)

func TestFetchTokensSuccess(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, tokensJSON := mocks.LoadFixture("../fixtures/tidal/tokens.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(tokensJSON)
	}
	server := mocks.Server(handler)
	defer server.Close()
	want := "mockToken"

	got, err := FetchToken(server.URL)

	assert.Nil(t, err, "should not have errored")
	assert.Equal(t, want, got, "should have gotten a mock token")
}
