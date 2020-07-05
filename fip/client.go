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
	// last records the timestamp for the last track received. It is used
	// as the offset value for the next request, so that we can move
	// timestampFrom forward when requests over 100 tracks are split.
	// Without this, we'd get the same trackCount chunk of tracks over and
	// over.
	last := timestampFrom

	// The API will return HTTP 400 if more than 100 tracks are requested.
	const maxTrackCountPerReq = 100

	remainingTracks := trackCount

	// So we must split the number of requests into 100 tracks chunks.
	if trackCount <= maxTrackCountPerReq {
		// last is not required since there is only a single request to
		// make
		list, _, err := getTracks(last, trackCount)
		if err != nil {
			logger.Error.Printf("error getting tracks: %v", err)
			return trackList, err
		}
		remainingTracks = 0
		trackList = append(trackList, list...)
	} else {
		list, last, err := getTracks(last, maxTrackCountPerReq)
		if err != nil {
			logger.Error.Printf("error getting tracks: %v", err)
			return trackList, err
		}
		remainingTracks = trackCount - maxTrackCountPerReq
		trackList = append(trackList, list...)
		return fip.Playlist(last, remainingTracks)
	}

	return trackList, err
}

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
		"Preparing to fetch playlist history (last %d tracks) "+
			"from timestamp %d (%s) for station ID %d",
		first,
		from,
		time.Unix(from, 0),
		station,
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
func unmarshalResponse(response *http.Response) (history historyResponse, err error) {
	responseData, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		errMsg := fmt.Sprintf("Error while reading response data: %v", err)
		logger.Error.Print(errMsg)
		return history, err
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
	logger.Trace.Println(JSON)

	for _, v := range JSON.Data.TimelineCursor.Edges {
		var track = extractor.Track{
			Title:  v.Node.Title,
			Artist: v.Node.Artist,
			Album:  v.Node.Album,
		}
		trackList = append(trackList, track)
	}

	if len(trackList) == 0 {
		errMsg := fmt.Sprintf("Empty playlist. Unmarshalled reponse: %v", JSON)
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
