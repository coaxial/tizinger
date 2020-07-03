// Package tidal implements a limited client for the tidal API.
package tidal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/coaxial/tizinger/extractor"
	"github.com/coaxial/tizinger/utils/logger"
)

// APIClient implements exporter.Client.
type APIClient struct{}

// baseURL can be overridden while testing to avoid live calls.
var baseURL = "https://api.tidalhifi.com/v1"

// jar is the cookie jar for the tidal client.
var jar http.CookieJar

// tidalClient is the client making requests to the Tidal API.
var tidalClient = &http.Client{
	Jar: jar,
}

// userData represents the data returned upon logging in that is necessary to
// compose authenticated requests.
type userData struct {
	SessionID   string
	CountryCode string
	UserID      int
}

var tidalUserData userData

// addTidalData adds the necessary headers to the request
func addTidalData(req *http.Request) {
	req.Header.Add("Origin", "https://listen.tidal.com")
	req.Header.Add("X-Tidal-SessionId", tidalUserData.SessionID)
	q := req.URL.Query()
	q.Add("token", tidalToken)
	q.Add("countryCode", tidalUserData.CountryCode)
	req.URL.RawQuery = q.Encode()
}

// CreatePlaylist creates playlists on Tidal.
func (ac APIClient) CreatePlaylist(name string, tracks extractor.Tracklist) (ok bool, err error) {
	ok, err = setToken()
	if err != nil {
		logger.Error.Printf("could not fetch tokens: %v", err)
		return ok, err
	}
	return ok, err
}

// queryTidal prepares and sends queries to the Tidal API. uri is where to send
// the request, payload is the JSON to send either in the body for
// http.MethodGet or as a form for http.MethodPost. method is the HTTP method
// to use, tidalJSON is a pointer to the struct to which the response will be
// unmarshalled.
func queryTidal(uri string, query map[string]string, payload string, method string, tidalJSON interface{}) (err error) {
	logger.Trace.Printf("preparing %q request to %q", method, uri)
	req, err := http.NewRequest(method, uri, strings.NewReader(payload))
	if err != nil {
		logger.Error.Printf("error building request: %v", err)
		return err
	}
	addTidalData(req)

	// Add the caller's querystring elements to the request
	for k, v := range query {
		q := req.URL.Query()
		q.Add(k, v)
		req.URL.RawQuery = q.Encode()
	}

	// Log request payload and query string for debugging
	body, err := req.GetBody()
	if err != nil {
		logger.Error.Printf("error %v", err)
	}
	debugBody, _ := ioutil.ReadAll(body)
	debugyBodyString := string(debugBody)
	qs := req.URL.RawQuery
	// Remove password from logs (when logging in)
	re := regexp.MustCompile(`(?P<firstHalf>.*"password":")(?P<passwordValue>.*?)(?P<SecondHalf>".*)`)
	debugyBodyString = re.ReplaceAllString(debugyBodyString, `$1<redacted>$3`)
	logger.Trace.Printf("request: body: %#v", string(debugyBodyString))
	logger.Trace.Printf("qs: %q", string(qs))

	logger.Info.Printf("sending %q request to %q", method, uri)
	// Using a global client so that we can reuse connections etc.
	resp, err := tidalClient.Do(req)
	logger.Info.Printf("received response %q, %d bytes", resp.Header.Get("Content-Type"), resp.ContentLength)
	contents, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		logger.Error.Printf("error reading response: %v", err)
		return err
	}
	err = json.Unmarshal(contents, &tidalJSON)
	if err != nil {
		logger.Error.Printf("error unmarshalling response: %v", err)
		return err
	}
	return err
}

// manifestURL is the tokens manifest's location.
// curtesy of https://github.com/yaronzz/Tidal-Media-Downloader
var manifestURL = "https://cdn.jsdelivr.net/gh/yaronzz/Tidal-Media-Downloader@latest/Else/tokens.json"

// tidalToken is the API token required for every request.
var tidalToken string

// setToken gets the currently valid token to send along with API requests.
// The purpose it to avoid hard-coding tokens so that the calls don't fail when
// Tidal rotates the token like they did in June 2020.
func setToken() (ok bool, err error) {
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
		return ok, err
	}

	tokens, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Error.Printf("error reading response: %v", err)
		return ok, err
	}

	var JSONTokens tokensResponse
	err = json.Unmarshal(tokens, &JSONTokens)
	if err != nil {
		logger.Error.Printf("error unmarshalling token: %v", err)
		return ok, err
	}
	tidalToken = JSONTokens.Token
	ok = true
	logger.Info.Printf("successfully set API token")
	return ok, err
}

// login performs a login with the Tidal API for a given username and password.
func login(username string, password string) (err error) {
	logger.Trace.Printf("preparing to log user %q in", username)
	endpoint := "/login/username"
	payload := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	uri := baseURL + endpoint

	err = queryTidal(uri, nil, payload, http.MethodPost, &tidalUserData)
	if err != nil {
		logger.Error.Printf("error logging in: %q", err)
		return err
	}

	logger.Info.Printf("successfully logged in")
	return err
}

func createEmptyPlaylist(userID int, name string, desc string) (UUID string, lu tidalTimestamp, err error) {
	logger.Trace.Printf("creating playlist (name: %q, desc: %q) for user %q", name, desc, strconv.Itoa(userID))
	endpoint := "/users/" + strconv.Itoa(userID) + "/playlists"
	payload := fmt.Sprintf(`{"title":%q,"desc":%q}`, name, desc)
	uri := baseURL + endpoint

	var playlistJSON playlist
	err = queryTidal(uri, nil, payload, http.MethodPost, &playlistJSON)
	if err != nil {
		logger.Error.Printf("error creating empty playlist: %q", err)
		return UUID, lu, err
	}

	UUID, lu = playlistJSON.UUID, playlistJSON.LastUpdated

	logger.Info.Printf("successfully created empty playlist (%q)", playlistJSON.UUID)
	logger.Trace.Printf("playlist UUID: %q, title: %q, desc: %q", playlistJSON.UUID, playlistJSON.Title, playlistJSON.Description)

	return UUID, lu, err
}

func search(track string, artist string, album string) (trackID int, err error) {
	logger.Info.Printf("search for track %q from artist %q on album %q", track, artist, album)
	endpoint := "/search/tracks"
	uri := baseURL + endpoint
	searchTerms := fmt.Sprintf("%s %s", track, artist)
	payload := `{"limit":"1","offset":"0","types":"TRACKS","includeContributors":"true"}`
	query := map[string]string{"query": searchTerms}
	var searchJSON searchResponse

	err = queryTidal(uri, query, payload, http.MethodGet, &searchJSON)
	trackID = searchJSON.Results[0].ID
	return trackID, err
}
