// Package fip extracts playlist data from fip.fr
package fip

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/coaxial/tizinger/extractor"
	"github.com/coaxial/tizinger/utils/logger"
)

// APIClient implements the extractor.Client interface for fip.fr
type APIClient struct{}

// endpointURL is the URL where the API endpoint is located. It can be
// overridden when testing to serve canned responses instead.
var endpointURL = "https://www.fip.fr/latest/api/graphql"

// Playlist returns the playlist history from `timestampFrom`, which is a Unix
// epoch in seconds. trackCount is the number of tracks to fetch. There seems
// to be around 320 tracks played per 24h.
func (fip APIClient) Playlist(timestampFrom int64, trackCount int) (trackList []extractor.Track, err error) {
	trackList, _, err = appendTracks(timestampFrom, trackCount, trackList)
	return trackList, err
}

// getTracks prepares, sends, and parses the request to the API. It returns the
// `count` number of tracks played up until `ts` along with the `last`
// timestamp of the last track in the list. `last` is required when splitting
// requests, so that we're not requesting the same `count` tracks over and over
// again but rather moving back in time.
func getTracks(ts int64, count int) (tracks extractor.Tracklist, last int64, err error) {
	req, err := buildRequest(ts, count)
	if err != nil {
		return tracks, last, err
	}

	client := buildClient()
	response, err := makeRequest(req, client)
	if err != nil {
		return tracks, last, err
	}

	fipHistoryJSON, err := unmarshalResponse(response)
	if err != nil {
		return tracks, last, err
	}

	tracks, err = buildTracklist(fipHistoryJSON)
	if err != nil {
		return tracks, last, err
	}

	last, err = extractEndCursor(&fipHistoryJSON)

	return tracks, last, err
}

// appendTracks splits the requests into 100 tracks chunks. Because the API
// will only process requests for 100 tracks maximum, it is necessary to make
// more than one request when requesting more.
func appendTracks(
	// ts is the timestamp to fetch backwards from.
	ts int64,
	// count is the numbers of tracks to fetch.
	count int,
	// prevChunk is the tracklist we already have from previous calls.
	prevChunk extractor.Tracklist,
) (
	// allChunks is prevChunks and the new chunk fetched for that iteration.
	allChunks extractor.Tracklist,
	// last is the timestamp for the last track in this itration's chunk,
	// so that we fetch the remaining tracks from that point in time rather
	// than from the original timestamp: we'd end up with identical chunks
	// every time otherwise.
	last int64,
	err error,
) {
	// We need to keep track of how many tracks remain. For now it's the
	// total wanted number of tracks since we haven't done anything yet.
	// Declaring remaining here makes it in scope for the if base case too.
	remaining := count
	// maxCount is the maximum number of tracks the API will return in one
	// request.
	const maxCount = 100 // tracks

	// This is the base case.
	if count <= maxCount {
		logger.Info.Printf("requesting less than %d tracks, doing it in one call", maxCount)
		chunk, last, err := getTracks(ts, count)
		if err != nil {
			logger.Error.Printf("error fetching tracks: %v", err)
			return allChunks, last, err
		}
		logger.Info.Printf("received %d tracks after requesting %d, %d more to get", len(chunk), count, remaining)
		allChunks = append(prevChunk, chunk...)

		return allChunks, last, err
	}
	logger.Info.Printf("requesting over %d tracks, splitting calls", maxCount)
	// About to get maxCount tracks so there will be that many tracks less
	// remaining to fetch.
	remaining -= maxCount
	chunk, last, err := getTracks(ts, maxCount)
	if err != nil {
		logger.Error.Printf("error fetching tracks: %v", err)
		return allChunks, last, err
	}
	logger.Info.Printf("received %d tracks after requesting %d, %d more to get", len(chunk), count, remaining)
	// We need all the tracks we already got from previous requests, plus
	// the tracks we just got.
	allChunks = append(prevChunk, chunk...)

	// Do it all again for the remaining tracks.
	return appendTracks(last, remaining, allChunks)
}

// buildRequest assembles the query string and headers. from is the timestamp
// from which to start looking back, first is how many tracks are requested.
func buildRequest(from int64, first int) (*http.Request, error) {
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

	timestamp := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(from, 10)))
	var stationIDs = map[string]int{
		"fip":            7,
		"fipRock":        64,
		"fipJazz":        65,
		"fipGroove":      66,
		"fipPop":         78,
		"fipElectro":     74,
		"fipMonde":       69,
		"fipReggae":      71,
		"fipToutNouveau": 70,
	}
	station := stationIDs["fip"]

	logger.Info.Printf(
		"preparing to fetch playlist history (last %d tracks) "+
			"from timestamp %d (%s) for station ID %d (FIP Paris)",
		first,
		from,
		time.Unix(from, 0),
		station,
	)

	req, err := http.NewRequest("GET", endpointURL, nil)
	if err != nil {
		errMsg := fmt.Sprintf("error while building new request: %v", err)
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
			station,
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
	logger.Info.Printf("initiating GET %s", req.URL)
	response, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf(
			"error while performing GET %s: %v",
			req.URL,
			err,
		)
		logger.Error.Print(errMsg)
		return nil, err
	}

	logger.Info.Print(
		fmt.Sprintf(
			"received response %q, %d bytes",
			response.Header.Get("content-type"),
			response.ContentLength,
		),
	)

	return response, nil
}

// unmarshalResponse parses the API response and unmarshals it to JSON.
func unmarshalResponse(response *http.Response) (history historyResponse, err error) {
	responseData, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		errMsg := fmt.Sprintf("error while reading response data: %v", err)
		logger.Error.Print(errMsg)
		return history, err
	}
	// A response that isn't HTTP 200 OK will still make err nil, so a
	// check needs to be done to see whether it succeeded or failed.
	if response.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf(
			"request failed, got HTTP %d (%v)",
			response.StatusCode,
			string(responseData),
		)
		logger.Error.Printf(errMsg)
		return history, errors.New(errMsg)
	}

	// We most likely have the playlist data, time to unmarshal it and pick
	// the fields we want.
	err = json.Unmarshal(responseData, &history)
	if err != nil {
		logger.Error.Printf("error unmarshalling history response: %v", err)
		return history, err
	}
	return history, err
}

// buildTracklist picks the relevant metadata from the API response and puts it
// into a []extractor.Track
func buildTracklist(JSON historyResponse) (trackList []extractor.Track, err error) {
	for _, v := range JSON.Data.TimelineCursor.Edges {
		var track = extractor.Track{
			Title:  v.Node.Title,
			Artist: v.Node.Artist,
			Album:  v.Node.Album,
		}
		trackList = append(trackList, track)
	}

	if len(trackList) == 0 {
		errMsg := fmt.Sprintf("empty playlist. Unmarshalled reponse: %v", JSON)
		logger.Error.Print(errMsg)
		return trackList, errors.New(errMsg)
	}
	return trackList, err
}

// extractEndCursor returns the timestamp for the last received track.
func extractEndCursor(JSON *historyResponse) (timestamp int64, err error) {
	ec := JSON.Data.TimelineCursor.PageInfo.EndCursor
	logger.Trace.Printf("converting %q to int64 timestamp", ec)
	endCursorByte, err := base64.StdEncoding.DecodeString(ec)
	timestamp, err = strconv.ParseInt(string(endCursorByte), 0, 64)
	if err != nil {
		logger.Error.Printf("error decoding endCursor %q to timestamp: %v", ec, err)
		return timestamp, err
	}
	logger.Trace.Printf("converted to %d", timestamp)
	return timestamp, err
}
