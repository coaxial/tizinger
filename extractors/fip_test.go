package extractors

import (
	"net/http"
	"testing"

	"github.com/coaxial/tizinger/playlist"
	"github.com/coaxial/tizinger/utils/mocks"
	"github.com/stretchr/testify/assert"
)

var extractor FipExtractor

func TestPlaylistErr(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Header().Set("Content-Type", "application/html")
		badReqResp := mocks.LoadFixture("../fixtures/fip/bad_req.json")
		resp.Write(badReqResp)
	}
	server := mocks.Server(handler)
	defer server.Close()
	extractor.SetEndpointUrl(server.URL)

	actual, err := extractor.Playlist(0)

	assert.Nil(t, actual)
	assert.Error(t, err)
}

func TestPlaylist(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		historyJson := mocks.LoadFixture("../fixtures/fip/history_response.json")
		resp.Write(historyJson)
	}
	server := mocks.Server(handler)
	defer server.Close()
	expected := []playlist.Track{
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
	extractor.SetEndpointUrl(server.URL)

	actual, err := extractor.Playlist(0)

	assert.Nil(t, err)
	assert.Equal(t, actual, expected, "should return a playlist")
}

func TestEmptyResponse(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		emptyResp := []byte("{}")
		resp.Write(emptyResp)
	}
	server := mocks.Server(handler)
	defer server.Close()
	extractor.SetEndpointUrl(server.URL)

	actual, err := extractor.Playlist(0)

	assert.Nil(t, actual, "should not return a playlist")
	assert.Error(t, err)
}
