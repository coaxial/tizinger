package Extractors

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	playlist "github.com/coaxial/tizinger/playlist"
)

type FipExtractor struct{}

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
type Edges struct {
	Node   Node   `json:"node"`
	Cursor string `json:"cursor"`
}
type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}
type TimelineCursor struct {
	Edges    []Edges  `json:"edges"`
	PageInfo PageInfo `json:"pageInfo"`
}
type Data struct {
	TimelineCursor TimelineCursor `json:"timelineCursor"`
}
type FipHistoryResponse struct {
	Data Data `json:"data"`
}

func (extractor FipExtractor) Playlist(timestampFrom int64) ([]playlist.Track, error) {
	// FIP uses graphql to serve its tracks history. To get the history for
	// any given date and time, issue a GET to
	// www.fip.fr/latest/api/graphql with a query string containing the
	// following:
	// operationName=History
	// variables={
	//   "first": <int>,
	//   "after": <base64 encoded seconds epoch>,
	//   "stationgId": <int>
	// }
	// extensions={
	//   "persistedQuery: {
	//     "version": 1,
	//     "sha256Hash": "f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270"
	//   }
	// }
	// It returns a FipHistoryResponse containing `first` number of tracks
	// played since `after` timestamp
	endpoint := "https://www.fip.fr/latest/api/graphql"
	first := 10
	from := time.Now().Unix()
	timestamp := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(from, 10)))
	stationId := 7 // 7 is FIP

	// Build URL with http.NewRequest so that the query string can be easily built
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	query := req.URL.Query()
	query.Add("operationName", "History")
	query.Add(
		"variables",
		fmt.Sprintf(
			"{\"first\":%d,\"after\":\"%s\",\"stationId\":%d}",
			first,
			timestamp,
			stationId,
		),
	)
	query.Add(
		"extensions",
		"{\"persistedQuery\":{\"version\":1,\"sha256Hash\":"+
			"\"f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270\"}}",
	)
	req.URL.RawQuery = query.Encode()

	// Build the client to make the request
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// A response that isn't HTTP 200 OK will still make err nil, so a
	// check needs to be done to make sure it succeeded
	if response.StatusCode != http.StatusOK {
		fmt.Printf("%#v", response)
		log.Fatal("Query didn't succeed")
		return nil, err
	}

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var responseObject FipHistoryResponse
	json.Unmarshal(responseData, &responseObject)
	var trackList []playlist.Track

	for _, v := range responseObject.Data.TimelineCursor.Edges {
		var track = playlist.Track{Title: v.Node.Title, Artist: v.Node.Artist, Album: v.Node.Album}
		trackList = append(trackList, track)
	}

	return trackList, nil
}
