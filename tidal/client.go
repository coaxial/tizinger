// Package tidal implements a limited client for the tidal API.
package tidal

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/coaxial/tizinger/playlist"
	"github.com/coaxial/tizinger/utils/logger"
)

// APIClient implements exporter.Client
type APIClient struct{}

// baseURL can be overridden while testing to avoid live calls.
var baseURL = "https://api.tidalhifi.com/v1"

// CreatePlaylist creates playlists on Tidal.
func (ac APIClient) CreatePlaylist(name string, tracks playlist.Tracklist) (ok bool, err error) {
	return true, nil
}

// func login() (ok bool, err error) {
// 	endpoint := "/login/username"
// 	var payload = map[string]string{
// 		"username": "username",
// 		"password": "password",
// 	}
// 	return true,nil
// }

// manifestURL is the tokens manifest's location.
// curtesy of https://github.com/yaronzz/Tidal-Media-Downloader
var manifestURL = "https://cdn.jsdelivr.net/gh/yaronzz/Tidal-Media-Downloader@latest/Else/tokens.json"

// fetchToken gets the currently valid token to send along with API requests.
// The purpose it to avoid hard-coding tokens so that the calls don't fail when
// Tidal rotates the token like they did in June 2020.
func fetchToken() (token string, err error) {
	type tokensResponse struct {
		Token      string `json:"token"`
		TokenPhone string `json:"token_phone"`
	}

	logger.Trace.Printf("getting tokens manifest at %q", manifestURL)
	resp, err := http.Get(manifestURL)
	logger.Trace.Printf(
		"Received response %q, %d bytes",
		resp.Header.Get("content-type"),
		resp.ContentLength,
	)
	if err != nil {
		logger.Error.Printf("error fetching API tokens from %q: %v", manifestURL, err)
		return token, err
	}

	tokens, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Error.Printf("error reading response: %v", err)
		return token, err
	}

	var JSONTokens tokensResponse
	json.Unmarshal(tokens, &JSONTokens)
	return JSONTokens.Token, err
}
