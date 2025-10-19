package fip

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/coaxial/tizinger/extractor"
	"github.com/coaxial/tizinger/utils/logger"
	"github.com/coaxial/tizinger/utils/mocks"
	"github.com/stretchr/testify/assert"
)

var client APIClient

func TestPlaylistErr(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Header().Set("Content-Type", "application/html")
		length, badReqResp := mocks.LoadFixture("../fixtures/fip/bad_req.json")
		resp.Header().Set("Content-Length", strconv.Itoa(length))
		_, _ = resp.Write(badReqResp)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	SetEndpointURL(server.URL)
	defer ResetEndpointURL()

	actual, err := client.Playlist(0, 10)

	assert.Nil(t, actual)
	assert.Error(t, err, "should return an error")
}

func TestPlaylist(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		length, historyJSON := mocks.LoadFixture("../fixtures/fip/history_response.json")
		resp.Header().Set("Content-Length", strconv.Itoa(length))
		_, _ = resp.Write(historyJSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	SetEndpointURL(server.URL)
	defer ResetEndpointURL()
	expected := []extractor.Track{
		{Title: "Scar tissue", Artist: "Red Hot Chili Peppers", Album: "Greatest hits"},
		{Title: "Off the wall", Artist: "Jil Is Lucky", Album: "Off the wall"},
		{Title: "Kalimba (Flute mix)", Artist: "Freakniks", Album: "Electro tunes"},
		{Title: "Tsukikaage no rendezvous", Artist: "Keiko Mari", Album: "Nippon girls: Japanese pop, beat & bossa nova 1966-1970"},
		{Title: "Un petit poisson, un petit oiseau", Artist: "Juliette Greco", Album: "Déshabillez-moi 1965-1969"},
		{Title: "I want to be happy", Artist: "Ray Brown", Album: "Brown Ray trio / Some of my best friends are guitarists"},
		{Title: "I'm so happy I can't stop crying", Artist: "Sting", Album: "Mercury falling"},
		{Title: "Sambarilove (feat. Roubinho Jacobina)", Artist: "Chiara Civello", Album: "Eclipse"},
		{Title: "Retiens l'été", Artist: "Double Francoise", Album: "Les bijoux"},
		{Title: "Serenade nº13 en Sol Maj K 525 \"\"une petite musique de nuit\"\" : I. Allegro", Artist: "I Musici", Album: "Mozart, pachelbel, albinoni"},
	}

	ts := time.Date(2019, time.July, 5, 0, 0, 0, 0, time.UTC).Unix()
	actual, err := client.Playlist(ts, 10)

	assert.Nil(t, err, "should not error")
	assert.Equal(t, expected, actual, "should return a playlist")
}

func TestEmptyResponse(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		emptyResp := []byte("{}")
		resp.Header().Set("Content-Length", strconv.Itoa(len(emptyResp)))
		_, _ = resp.Write(emptyResp)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	SetEndpointURL(server.URL)
	defer ResetEndpointURL()

	actual, err := client.Playlist(0, 10)

	assert.Nil(t, actual, "should not return a playlist")
	assert.Error(t, err)
}

func ExampleAPIClient_Playlist() {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		length, historyJSON := mocks.LoadFixture("../fixtures/fip/history_response.json")
		resp.Header().Set("Content-Length", strconv.Itoa(length))
		_, _ = resp.Write(historyJSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	SetEndpointURL(server.URL)
	defer ResetEndpointURL()

	var fipClient APIClient
	// Get the list of 10 tracks played on FIP since 2019-07-25 00:30:00 GMT (date
	// doesn't matter as the fixture will return the same data for any timestamp)
	tracks, err := fipClient.Playlist(1562284800, 10)
	if err != nil {
		log.Fatalf("Could not fetch FIP tracks: %v", err)
	}

	fmt.Println(tracks)
	// Output: [{Scar tissue Red Hot Chili Peppers Greatest hits} {Off the wall Jil Is Lucky Off the wall} {Kalimba (Flute mix) Freakniks Electro tunes} {Tsukikaage no rendezvous Keiko Mari Nippon girls: Japanese pop, beat & bossa nova 1966-1970} {Un petit poisson, un petit oiseau Juliette Greco Déshabillez-moi 1965-1969} {I want to be happy Ray Brown Brown Ray trio / Some of my best friends are guitarists} {I'm so happy I can't stop crying Sting Mercury falling} {Sambarilove (feat. Roubinho Jacobina) Chiara Civello Eclipse} {Retiens l'été Double Francoise Les bijoux} {Serenade nº13 en Sol Maj K 525 ""une petite musique de nuit"" : I. Allegro I Musici Mozart, pachelbel, albinoni}]
}

func TestEndCursorConvert(t *testing.T) {
	var mockJSON historyResponse
	mockJSON.Data.TimelineCursor.PageInfo.EndCursor = "MTU5Mjg5MDQxNw=="
	wanted := int64(1592890417)

	got, err := extractEndCursor(&mockJSON)

	assert.Nil(t, err, "should not error")
	assert.Equal(t, wanted, got, "should convert the base64 timestamp to an int64")
}

func TestPlaylist200(t *testing.T) {
	part1Sent := false
	handler := func(resp http.ResponseWriter, req *http.Request) {
		var fixture string
		if !part1Sent {
			fixture = "../fixtures/fip/history_100tracks_part1.json"
			part1Sent = true

		} else {
			fixture = "../fixtures/fip/history_100tracks_part2.json"
		}
		logger.Trace.Printf("using fixture %q", fixture)
		length, historyJSON := mocks.LoadFixture(fixture)
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.Header().Set("Content-Length", strconv.Itoa(length))
		_, _ = resp.Write(historyJSON)
	}
	server := mocks.Server(http.HandlerFunc(handler))
	defer server.Close()
	SetEndpointURL(server.URL)
	defer ResetEndpointURL()

	actual, err := client.Playlist(0, 200)

	assert.Nil(t, err, "should not error")
	assert.Equal(t, 200, len(actual), "should return 200 elements")
	assert.Equal(t, "Scar tissue", actual[0].Title, "should match the first track from the first response part")
	assert.Equal(t, "Belleville", actual[100].Title, "should match the first track from the second response part")
}
