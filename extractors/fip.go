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

// Node contains the playlist individual tracks information from the API
type Node struct {
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

// Edges is a wrapper key from API response
type Edges struct {
	Node   Node   `json:"node"`
	Cursor string `json:"cursor"`
}

// PageInfo contains pagination info
type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

// TimelineCursor is a wrapper for pagination info and track info
type TimelineCursor struct {
	Edges    []Edges  `json:"edges"`
	PageInfo PageInfo `json:"pageInfo"`
}

// Data a wrapper for TimelineCursor
type Data struct {
	TimelineCursor TimelineCursor `json:"timelineCursor"`
}

// FipHistoryResponse is the API response to the History call
type FipHistoryResponse struct {
	Data Data `json:"data"`
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
	from := time.Now().Unix()
	timestamp := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(from, 10)))
	const fip = 7
	stationID := fip

	// Build URL with http.NewRequest so that the query string can be easily built
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
			"{\"first\":%d,\"after\":\"%s\",\"stationID\":%d}",
			first,
			timestamp,
			stationID,
		),
	)
	query.Add(
		"extensions",
		"{\"persistedQuery\":{\"version\":1,\"sha256Hash\":"+
			"\"f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270\"}}",
	)
	req.URL.RawQuery = query.Encode()

	// Build client
	client := &http.Client{}
	logger.Info.Printf("Initiating GET %s", req.URL)
	// Make request
	response, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Error while performing GET: %v", err)
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
	responseData, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		errMsg := fmt.Sprintf("Error while reading response data: %v", err)
		logger.Error.Print(errMsg)
		return nil, err
	}
	// A response that isn't HTTP 200 OK will still make err nil, so a
	// check needs to be done to make sure it succeeded
	if response.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf(
			"Request failed, got HTTP %d (%v)",
			response.StatusCode,
			string(responseData),
		)
		logger.Error.Printf(errMsg)
		return nil, errors.New(errMsg)
	}

	// We most likely have the playlist data, time to unmarshal it and pick
	// the fields we want.
	var responseObject FipHistoryResponse
	json.Unmarshal(responseData, &responseObject)
	var trackList []playlist.Track

	for _, v := range responseObject.Data.TimelineCursor.Edges {
		var track = playlist.Track{
			Title:  v.Node.Title,
			Artist: v.Node.Artist,
			Album:  v.Node.Album,
		}
		trackList = append(trackList, track)
	}

	if len(trackList) == 0 {
		errMsg := fmt.Sprintf("Empty playlist. Unmarshalled reponse: %v", responseObject)
		logger.Error.Print(errMsg)
		return nil, errors.New(errMsg)
	}
	return trackList, nil
}
