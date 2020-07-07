package tidal

import (
	"net/http"
	"strings"
	"testing"

	"github.com/coaxial/tizinger/utils/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestFetchingTokens(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, tokensJSON := mocks.LoadFixture("../fixtures/tidal/tokens.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(tokensJSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	want := "mockToken"

	err := SetToken(server.URL)

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
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()
	want := userData{SessionID: "mock-session-id", CountryCode: "MK", UserID: 133713373}

	err := login("mockuser@example.org", "secret")

	assert.Nil(t, err, "should not have errored")
	assert.Equal(t, want, tidalUserData, "should have populated user data")
}

func TestComposeHeadersNilSessionID(t *testing.T) {
	want := []struct {
		header string
		value  string
	}{
		{"Origin", "https://listen.tidal.com"},
		{"X-Tidal-Session-ID", ""},
	}

	tidalToken = "mockToken"
	mockReq, _ := http.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(""))
	addTidalData(mockReq)

	for _, w := range want {
		assert.Equal(t, w.value, mockReq.Header.Get(w.header), "should set the headers")
	}
	assert.Equal(t, "mockToken", mockReq.URL.Query().Get("token"), "should add token to request")
}

func TestComposeHeaders(t *testing.T) {
	tidalUserData.SessionID = "mock-session-id"
	tidalUserData.CountryCode = "MK"
	tidalToken = "mock-token"
	mockReq, _ := http.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(""))
	addTidalData(mockReq)

	assert.Equal(t, "mock-session-id", mockReq.Header.Get("X-Tidal-SessionId"), "should set the headers")
	assert.Equal(t, "MK", mockReq.URL.Query().Get("countryCode"), "should set the country code")
	assert.Equal(t, "mock-token", mockReq.URL.Query().Get("token"), "should set the token")
}

func TestCreateEmptyPlaylist(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, JSON := mocks.LoadFixture("../fixtures/tidal/playlist-create_response.json")
		resp.WriteHeader(http.StatusCreated)
		resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(JSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()

	UUID, err := createEmptyPlaylist(1337, "mock playlist name", "mock playlist description")
	want := struct {
		UUID string
	}{
		"mock-playlist-uuid",
	}

	assert.Equal(t, want.UUID, UUID, "should have returned the created playlist's UUID")
	assert.Nil(t, err, "should not have errored")
}

func TestSearch(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, JSON := mocks.LoadFixture("../fixtures/tidal/search-track_result_response.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(JSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()

	got, err := search("mock track", "mock artist", "mock album")
	want := 132616868

	assert.Equal(t, want, got, "should have returned the track's ID")
	assert.Nil(t, err, "should not have errored")
}

func TestSearchNoResult(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, JSON := mocks.LoadFixture("../fixtures/tidal/search-track_noresult_response.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(JSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()

	got, err := search("mock track", "mock artist", "mock album")
	want := -1

	assert.Equal(t, want, got, "should not have found a track")
	assert.Nil(t, err, "should not have errored")
}

func TestQueryNok(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusInternalServerError)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()

	got, err := search("mock track", "mock artist", "mock album")

	assert.Error(t, err, "should have errored")
	assert.Equal(t, got, -1, "should not have found a track")
}

func TestPopulatePlaylist(t *testing.T) {
	addTrackHandler := func(resp http.ResponseWriter, req *http.Request) {
		length, JSON := mocks.LoadFixture("../fixtures/tidal/playlist-add_success_response.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(JSON)
	}
	getLastUpdatedHandler := func(resp http.ResponseWriter, req *http.Request) {
		length, JSON := mocks.LoadFixture("../fixtures/tidal/playlist-get_response.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(JSON)
	}
	r := mux.NewRouter()
	r.HandleFunc("/playlists/mockUUID/items", addTrackHandler)
	r.HandleFunc("/playlists/mockUUID", getLastUpdatedHandler)
	server := mocks.Server(r)
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()

	tests := []struct {
		input []int
		want  int
		msg   string
	}{
		{
			[]int{42, 666, 1337},
			3,
			"should have returned the number of unique tracks added",
		}, {
			[]int{666, 42, 666, 1337},
			3,
			"should have returned the number of unique tracks added",
		},
	}
	playlist := "mockUUID"

	for _, test := range tests {
		got, err := populatePlaylist(test.input, playlist)
		assert.Nil(t, err, "should not have errored")
		assert.Equal(t, test.want, got, test.msg)
	}
}

func TestGetLastUpdated(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		length, JSON := mocks.LoadFixture("../fixtures/tidal/playlist-get_response.json")
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		resp.Header().Set("Content-Length", string(length))
		resp.Write(JSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	originalURL = baseURL
	baseURL = server.URL
	defer func() { baseURL = originalURL }()

	want := int64(1595684220666)
	got, err := getLastUpdated("mock-playlist-id")
	assert.Nil(t, err, "should not have errored")
	assert.Equal(t, want, got, "should have returned the int64 timestamp")
}
