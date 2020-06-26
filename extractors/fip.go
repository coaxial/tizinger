// Package extractors pull out historical playlist data from their source. All
// extractors live under this module, each extractor cattering to one
// particular source.
// The interface for an extractor is to implement a Playlist method that
// returns a []playlist.Track.
package extractors

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/coaxial/tizinger/playlist"
	"github.com/coaxial/tizinger/utils/logger"
)

// node contains the playlist individual tracks information from the API
type node struct {
	// the song's title is under the subtitle key
	Title       string `json:"subtitle"`
	StartTime   int    `json:"start_time"`
	EndTime     int    `json:"end_time"`
	Album       string `json:"album"`
	MusicalKind string `json:"musical_kind"`
	Year        int    `json:"year"`
	// the artist's name is under the title key
	Artist string `json:"title"`
}

// edges is a wrapper key from API response
type edges struct {
	Node   node   `json:"node"`
	Cursor string `json:"cursor"`
}

// pageInfo contains pagination info
type pageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

// timelineCursor is a wrapper for pagination info and track info
type timelineCursor struct {
	Edges    []edges  `json:"edges"`
	PageInfo pageInfo `json:"pageInfo"`
}

// data is a wrapper for TimelineCursor
type data struct {
	TimelineCursor timelineCursor `json:"timelineCursor"`
}

// fipHistoryResponse is the API response to the History call
type fipHistoryResponse struct {
	Data data `json:"data"`
}

// FipExtractor is the extractor dealing with fip.fr play history
type FipExtractor struct {
}

var endpointURL string

func init() {
	endpointURL = "https://www.fip.fr/latest/api/graphql"
}

// SetEndpointURL is for testing, so that a mock server can be used instead of
// the live one, and arbitrary responses or failures can be served as needed.
func (extractor *FipExtractor) SetEndpointURL(url string) {
	endpointURL = url
}

// Playlist returns the playlist history from `timestampFrom`, which is a Unix
// epoch in seconds.
func (extractor FipExtractor) Playlist(timestampFrom int64) ([]playlist.Track, error) {
	// TODO: Fetch 24h worth and/or check all tracks are from the same day
	req, err := buildRequest(timestampFrom)
	if err != nil {
		return nil, err
	}

	client := buildClient()
	response, err := makeRequest(req, client)
	if err != nil {
		return nil, err
	}

	fipHistoryJSON, err := unmarshalResponse(response)
	if err != nil {
		return nil, err
	}

	trackList, err := buildTracklist(fipHistoryJSON)
	if err != nil {
		return nil, err
	}

	return trackList, nil
}

// buildRequest assembles the query string and headers.
func buildRequest(from int64) (*http.Request, error) {
	// FIP uses graphql to serve its tracks history. To get the history for
	// any given date and time, issue a GET to
	// www.fip.fr/latest/api/graphql with a query string containing the
	// following:
	// operationName=History
	// variables={
	//   "first": <int>,
	//   "after": <base64 encoded seconds epoch>,
	//   "stationgID": <int>
	// }
	// extensions={
	//   "persistedQuery: {
	//     "version": 1,
	//     "sha256Hash": "f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270"
	//   }
	// }
	// It returns a FipHistoryResponse containing `first` number of tracks
	// played since `after` timestamp

	first := 10
	timestamp := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(from, 10)))
	const fip = 7
	stationID := fip

	logger.Info.Printf(
		"Preparing to fetch playlist history (last %d tracks) "+
			"from timestamp %d (%s) for station ID %d",
		first,
		from,
		time.Unix(from, 0),
		stationID,
	)

	req, err := http.NewRequest("GET", endpointURL, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error while building new request: %v", err)
		logger.Error.Println(errMsg)
		return nil, err
	}
	query := req.URL.Query()
	query.Add("operationName", "History")
	query.Add(
		"variables",
		fmt.Sprintf(
			`{"first":%d,"after":"%s","stationID":%d}`,
			first,
			timestamp,
			stationID,
		),
	)
	query.Add(
		"extensions",
		`{"persistedQuery":{"version":1,"sha256Hash":"f8f404573583a6a`+
			`9410cd24637f214a0b93038696c1d20f19202111b51fd8270"}}`,
	)
	req.URL.RawQuery = query.Encode()
	return req, nil
}

func buildClient() *http.Client {
	client := &http.Client{}
	return client
}

// makeRequest sends the request to the API
func makeRequest(req *http.Request, client *http.Client) (*http.Response, error) {
	logger.Info.Printf("Initiating GET %s", req.URL)
	response, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf(
			"Error while performing GET %s: %v",
			req.URL,
			err,
		)
		logger.Error.Print(errMsg)
		return nil, err
	}

	logger.Info.Print(
		fmt.Sprintf(
			"Received response %q, %d bytes",
			response.Header.Get("content-type"),
			response.ContentLength,
		),
	)

	return response, nil
}

// unmarshalResponse parses the API response and unmarshals it to JSON.
func unmarshalResponse(response *http.Response) (fipHistoryResponse, error) {
	var responseObject fipHistoryResponse

	responseData, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		errMsg := fmt.Sprintf("Error while reading response data: %v", err)
		logger.Error.Print(errMsg)
		return responseObject, err
	}
	// A response that isn't HTTP 200 OK will still make err nil, so a
	// check needs to be done to see whether it succeeded or failed.
	if response.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf(
			"Request failed, got HTTP %d (%v)",
			response.StatusCode,
			string(responseData),
		)
		logger.Error.Printf(errMsg)
		return responseObject, errors.New(errMsg)
	}

	// We most likely have the playlist data, time to unmarshal it and pick
	// the fields we want.
	json.Unmarshal(responseData, &responseObject)
	return responseObject, nil
}

func buildTracklist(JSON fipHistoryResponse) ([]playlist.Track, error) {
	logger.Trace.Println(JSON)
	var trackList []playlist.Track

	for _, v := range JSON.Data.TimelineCursor.Edges {
		var track = playlist.Track{
			Title:  v.Node.Title,
			Artist: v.Node.Artist,
			Album:  v.Node.Album,
		}
		trackList = append(trackList, track)
	}

	if len(trackList) == 0 {
		errMsg := fmt.Sprintf("Empty playlist. Unmarshalled reponse: %v", JSON)
		logger.Error.Print(errMsg)
		return nil, errors.New(errMsg)
	}
	return trackList, nil
}
