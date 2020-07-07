// Package tidal implements a limited client for the tidal API.
package tidal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coaxial/tizinger/extractor"
	"github.com/coaxial/tizinger/utils/credentials"
	"github.com/coaxial/tizinger/utils/helpers"
	"github.com/coaxial/tizinger/utils/logger"
)

// APIClient implements exporter.Client.
type APIClient struct{}

// baseURL can be overridden while testing to avoid live calls.
var baseURL = "https://api.tidalhifi.com/v1"

// jar is the cookie jar for the tidal client.
var jar http.CookieJar

// tidalClient is the client making requests to the Tidal API. It is defined
// here so that this client instance is reused.
var tidalClient = &http.Client{}

// userData represents the data returned upon logging in that is necessary to
// compose authenticated requests.
type userData struct {
	SessionID   string
	CountryCode string
	UserID      int
}

// tidalUserData is the instance holding user data after logging in.
var tidalUserData userData

// addTidalData adds the necessary headers to the request
func addTidalData(req *http.Request) {
	// Set a default country code, to be overridden by the user's value if
	// the user is logged in. This enables requests before being logged in.
	cc := "US"
	if tidalUserData.CountryCode != "" {
		cc = tidalUserData.CountryCode
	}

	// Only add the session ID if we have one (i.e. are logged in)
	if tidalUserData.SessionID != "" {
		req.Header.Add("X-Tidal-SessionId", tidalUserData.SessionID)
	}
	// This is what the webclient sends. Not strictly necessary, but helps
	// blend in.
	req.Header.Add("Origin", "https://listen.tidal.com")

	q := req.URL.Query()
	// The token is shared amongst users, but is necessary with each
	// request.
	q.Add("token", tidalToken)
	// The countryCode is also mandatory. Each user has one assigned and it
	// seems to be the country they signed up in (not the one they're in
	// when making the request).
	q.Add("countryCode", cc)
	req.URL.RawQuery = q.Encode()
}

// CreatePlaylist creates playlists on Tidal.
func (ac APIClient) CreatePlaylist(name string, tracks extractor.Tracklist) (err error) {
	err = setToken()
	if err != nil {
		logger.Error.Printf("could not fetch tokens: %v", err)
		return err
	}

	// The credentials file can have more than one Tidal account.
	accounts, err := credentials.Tidal()
	if err != nil {
		logger.Error.Printf("error fetching Tidal account information: %v", err)
		return err
	}

	var trackIDs []int
	// search for tracks now, it only need to be done once for all users as
	// the track IDs on Tidal don't depend on the user. This makes things
	// a bit faster when creating playlists on several accounts.
	for i, t := range tracks {
		logger.Info.Printf("searching for track %d/%d: %q by %q", i+1, len(tracks), t.Title, t.Artist)
		ID, err := search(t.Title, t.Artist, t.Album)
		if err != nil {
			logger.Error.Printf("error when searching for track %q %q %q", t.Title, t.Artist, t.Album)
			return err
		}
		// -1 means track not found.
		if ID != -1 {
			trackIDs = append(trackIDs, ID)
		}
	}

	// There can be more than one account, playlists are created and
	// populated for each.
	for i, a := range accounts {
		logger.Info.Printf("processing account %q (%d/%d)", a.Username, i+1, len(accounts))
		err = login(a.Username, a.Password)
		if err != nil {
			logger.Error.Printf("error logging in: %v", err)
			return err
		}
		playlistID, err := createEmptyPlaylist(tidalUserData.UserID, name, "")
		if err != nil {
			logger.Error.Printf("error creating empty playlist: %v", err)
			return err
		}
		countAdded, err := populatePlaylist(trackIDs, playlistID)
		if err != nil {
			logger.Error.Printf("error populating playlist: %v", err)
			return err
		}
		logger.Info.Printf("added %d/%d tracks to playlist %q", countAdded, len(trackIDs), playlistID)
		logger.Info.Printf("done with account %q (%d/%d)", a.Username, i+1, len(accounts))
	}
	return err
}

// queryTidal prepares and sends queries to the Tidal API. uri is where to send
// the request, payload is the JSON to send either in the body for
// http.MethodGet or as a form for http.MethodPost. method is the HTTP method
// to use, tidalJSON is a pointer to the struct to which the response will be
// unmarshalled.
func queryTidal(
	uri string, // where to send the request
	headers map[string]string, // extra headers besides the Tidal headers
	query map[string]string, // query string elements
	payload url.Values, // form data
	method string, // HTTP method
	tidalJSON interface{}, // variable to unmarshal the response in
) (err error) {
	logger.Trace.Printf("preparing %q request to %q", method, uri)
	req, err := http.NewRequest(method, uri, strings.NewReader(payload.Encode()))
	if err != nil {
		logger.Error.Printf("error building request: %v", err)
		return err
	}
	addTidalData(req)
	// POST with a payload means we're posting a form.
	if method == http.MethodPost && len(payload) > 0 {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// Merge-in the caller-supplied querystring elements.
	for k, v := range query {
		q := req.URL.Query()
		q.Add(k, v)
		req.URL.RawQuery = q.Encode()
	}

	// Merge-in the caller-supplied headers.
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Log request payload and query string for debugging
	body, err := req.GetBody()
	if err != nil {
		logger.Error.Printf("error %v", err)
		return err
	}
	debugBody, _ := ioutil.ReadAll(body)
	debugyBodyString := string(debugBody)
	qs := req.URL.RawQuery
	h := req.Header
	// Remove password from logs
	re := regexp.MustCompile(`(?P<firstHalf>.*"password":")(?P<passwordValue>.*?)(?P<SecondHalf>".*)`)
	debugyBodyString = re.ReplaceAllString(debugyBodyString, `$1<redacted>$3`)
	logger.Trace.Printf("request: body: %#v", string(debugyBodyString))
	logger.Trace.Printf("qs: %q", string(qs))
	// Dump headers
	var hstr strings.Builder
	for k, v := range h {
		// Don't log the session ID as it can be used to impersonate
		// user
		if k == "X-Tidal-SessionId" {
			fmt.Fprintf(&hstr, `%q: %q, `, k, `<redacted>`)
		} else {
			fmt.Fprintf(&hstr, `%q: %q, `, k, v)
		}
	}
	logger.Trace.Printf("headers: %s", hstr.String())

	logger.Info.Printf("sending %q request to %q", method, uri)
	// Using a global client so that we can reuse connections etc.
	resp, err := tidalClient.Do(req)
	if err != nil {
		logger.Error.Printf("error making request: %v", err)
		return err
	}
	// The Content-Length headers is sometimes missing and shows as -1
	// length. This is up to the server and there isn't much that can be
	// done about it.
	logger.Info.Printf("received response %q, %d bytes", resp.Header.Get("Content-Type"), resp.ContentLength)
	contents, err := ioutil.ReadAll(resp.Body)
	// The request succeeds only for HTTP 200 OK or HTTP 201 Created (for
	// playlist creation)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
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
	logger.Error.Printf("tidal API responded with HTTP %d: %q", resp.StatusCode, contents)
	return fmt.Errorf("tidal API responded with HTTP %d: %q", resp.StatusCode, contents)
}

// manifestURL is the tokens manifest's location. Tidal seems to rotate them
// (rarely).
// curtesy of https://github.com/yaronzz/Tidal-Media-Downloader
var manifestURL = "https://cdn.jsdelivr.net/gh/yaronzz/Tidal-Media-Downloader@latest/Else/tokens.json"

// tidalToken is the API token required for every request.
var tidalToken string

// setToken gets the currently valid token to send along with API requests.
// The purpose it to avoid hard-coding tokens so that the calls don't fail when
// Tidal rotates the token like they did in June 2020.
func setToken() (err error) {
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
		return err
	}

	tokens, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Error.Printf("error reading response: %v", err)
		return err
	}

	var JSONTokens tokensResponse
	err = json.Unmarshal(tokens, &JSONTokens)
	if err != nil {
		logger.Error.Printf("error unmarshalling token: %v", err)
		return err
	}
	// TokenPhone is probably equally good.
	tidalToken = JSONTokens.Token
	logger.Info.Printf("successfully set API token to %q", tidalToken)
	return err
}

// login performs a login with the Tidal API for a given username and password.
func login(username string, password string) (err error) {
	logger.Trace.Printf("preparing to log user %q in", username)
	endpoint := "/login/username"
	payload := url.Values{
		"username": {username},
		"password": {password},
	}
	uri := baseURL + endpoint

	err = queryTidal(uri, nil, nil, payload, http.MethodPost, &tidalUserData)
	if err != nil {
		logger.Error.Printf("error logging in: %q", err)
		return err
	}

	logger.Info.Printf("successfully logged use %q in", username)
	return err
}

// createEmptyPlaylist creates a new, empty playlist with the supplied title
// and description for user userID on Tidal.
func createEmptyPlaylist(userID int, title string, description string) (UUID string, err error) {
	logger.Trace.Printf(
		"creating playlist (title: %q, description: %q) for user %q",
		title, description, strconv.Itoa(userID),
	)
	endpoint := "/users/" + strconv.Itoa(userID) + "/playlists"
	payload := url.Values{
		"title":       {title},
		"description": {description},
	}
	uri := baseURL + endpoint

	var playlistJSON playlist
	err = queryTidal(uri, nil, nil, payload, http.MethodPost, &playlistJSON)
	if err != nil {
		logger.Error.Printf("error creating empty playlist: %q", err)
		return UUID, err
	}

	UUID = playlistJSON.UUID

	logger.Info.Printf("successfully created empty playlist (%q)", UUID)
	logger.Trace.Printf("playlist UUID: %q, title: %q, desc: %q", UUID, playlistJSON.Title, playlistJSON.Description)

	return UUID, err
}

// search will search for "<track> <artist>" on Tidal and return the track's
// Tidal ID. The ID is -1 if there are no results for that search.
func search(track string, artist string, album string) (trackID int, err error) {
	const trackNotFound = -1
	trackID = trackNotFound
	endpoint := "/search/tracks"
	uri := baseURL + endpoint
	searchTerms := fmt.Sprintf("%s %s", track, artist)
	payload := url.Values{
		"limit":               {"1"}, // We're only interested in the first match
		"offset":              {"0"},
		"types":               {"TRACKS"}, // Only search for tracks
		"includeContributors": {"true"},   // Not sure what this is, but the Tidal client apps have it set to true
	}
	query := map[string]string{"query": searchTerms}
	var searchJSON searchResponse

	logger.Info.Printf("search for track %q from artist %q on album %q", track, artist, album)
	err = queryTidal(uri, nil, query, payload, http.MethodGet, &searchJSON)
	if err != nil {
		logger.Error.Printf("error looking for track %q: %v", track+" "+artist, err)
		return trackID, err
	}
	// Check if the request returned any matches.
	if len(searchJSON.Results) == 0 {
		logger.Warning.Printf("no matching track found for track %q", track+" "+artist)
		return trackID, err
	}
	trackID = searchJSON.Results[0].ID
	logger.Info.Printf("found matching track with ID %q", trackID)
	return trackID, err
}

// populatePlaylist adds the tracks with trackID to the playlist with
// playlistID.
func populatePlaylist(trackIDs []int, playlistID string) (countAdded int, err error) {
	// Remove duplicate tracks from list
	uniqIDs := helpers.Uniq(trackIDs)
	endpoint := "/playlists/" + playlistID + "/items"
	uri := baseURL + endpoint

	logger.Info.Printf("adding %d unique tracks to playlist %q", len(uniqIDs), playlistID)
	// TODO: There might be a way to add tracks in bulk since the payload
	// key trackIds is plural, maybe as an array of trackIDs?
	for i, ID := range uniqIDs {
		// The If-None-Match header value is the playlist's LastUpdated
		// timestamp and a millisecond Unix timestamp. It changes after
		// every track add, so refresh it every time.
		inmVal, err := getLastUpdated(playlistID)
		if err != nil {
			logger.Error.Printf("error updating last updated value: %v", err)
			return countAdded, err
		}
		logger.Info.Printf("adding track %d (%d/%d)", ID, i+1, len(uniqIDs))
		payload := url.Values{
			"onArtifactNotFound": {"FAIL"},
			"onDupes":            {"FAIL"},
			"trackIds":           {strconv.Itoa(ID)},
		}
		// The API refuses to add the track if the If-None-Match header
		// is incorrect!
		inmHeader := map[string]string{"If-None-Match": strconv.FormatInt(inmVal, 10)}
		var populateResult populatePlaylistResult
		err = queryTidal(uri, inmHeader, nil, payload, http.MethodPost, &populateResult)
		if err != nil {
			logger.Error.Printf("error adding track %d to playlist %q:%v", ID, playlistID, err)
			return countAdded, err
		}
		logger.Info.Printf("successfully added track %d (%d/%d) to playlist %q", ID, i+1, len(uniqIDs), playlistID)
		countAdded++
	}
	logger.Info.Printf("successfully added %d/%d tracks to playlist %q", countAdded, len(uniqIDs), playlistID)
	return countAdded, err
}

// getLastUpdated gets the LastUpdated ms timestamp for the playlist matching
// playlistID
func getLastUpdated(playlistID string) (lu int64, err error) {
	endpoint := "/playlists/" + playlistID
	uri := baseURL + endpoint
	var getPlaylistResult playlist

	logger.Trace.Printf("getting last updated for playlist %q", playlistID)
	err = queryTidal(uri, nil, nil, nil, http.MethodGet, &getPlaylistResult)
	if err != nil {
		logger.Error.Printf("error getting playlist metadata: %v", err)
		return lu, err
	}
	lu = getPlaylistResult.LastUpdated.UnixNano() / int64(time.Millisecond)
	logger.Trace.Printf("last updated at %q", lu)
	return lu, err
}
