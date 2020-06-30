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

	_, err := FetchToken(server.URL)

	assert.Nil(t, err, "should not have errored")
	assert.Equal(t, want, tidalToken, "should have gotten a mock token")
}

func TestLogin(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, JSON := mocks.LoadFixture("../fixtures/tidal/login_response.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(JSON)
	}
	server := mocks.Server(handler)
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()
	want := userData{SessionID: "mock-session-id", CountryCode: "MK", UserID: 133713373}

	ok, err := login("mockuser@example.org", "secret")

	assert.Nil(t, err, "should not have errored")
	assert.True(t, ok, "should have succeeded")
	assert.Equal(t, want, tidalUserData, "should have populated user data")
}
