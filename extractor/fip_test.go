package extractor_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/coaxial/tizinger/extractor"
	"github.com/coaxial/tizinger/playlist"
	"github.com/coaxial/tizinger/utils/mocks"
	"github.com/stretchr/testify/assert"
)

var subject extractor.FipExtractor

func TestPlaylistErr(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Header().Set("Content-Type", "application/html")
		badReqResp := mocks.LoadFixture("../fixtures/fip/bad_req.json")
		resp.Write(badReqResp)
	}
	server := mocks.Server(handler)
	defer server.Close()
	subject.SetEndpointURL(server.URL)
	defer subject.SetEndpointURL("https://www.fip.fr/latest/api/graphql")

	actual, err := subject.Playlist(0)

	assert.Nil(t, actual)
	assert.Error(t, err, "should return an error")
}

func TestPlaylist(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		historyJSON := mocks.LoadFixture("../fixtures/fip/history_response.json")
		resp.Write(historyJSON)
	}
	server := mocks.Server(handler)
	subject.SetEndpointURL(server.URL)
	defer subject.SetEndpointURL("https://www.fip.fr/latest/api/graphql")
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

	actual, err := subject.Playlist(0)

	assert.Nil(t, err, "should not error")
	assert.Equal(t, expected, actual, "should return a playlist")
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
	subject.SetEndpointURL(server.URL)
	defer subject.SetEndpointURL("https://www.fip.fr/latest/api/graphql")

	actual, err := subject.Playlist(0)

	assert.Nil(t, actual, "should not return a playlist")
	assert.Error(t, err)
}

func ExamplePlaylist() {
	var fipExtractor extractor.FipExtractor
	// Get the list of tracks played on FIP since 2020-07-25 00:30:00 GMT
	tracks, err := fipExtractor.Playlist(1564014600)
	if err != nil {
		fmt.Sprintf("Could not fetch FIP tracks: %v", err)
	}

	fmt.Println(tracks)
	// Output: [{Riding the sun Howls Howls} {In the wake of adversity Dead Can Dance Within the realm of a dying sun} {Madame rêve Alain Bashung Osez Josephine} {The Planets op 32 : 3. Mercury, the Winged Messenger Orchestre Symphonique De Chicago Gustav Holst : Les Planètes} {Annie : The hard-knock life Alicia Morton BOF TV / Annie} {Bruce Lee Catastrophe Bruce Lee} {New comer 1 Walt Rockman Dusty fingers} {Cars Gary Numan The pleasure principle / Warriors} {Radio #1 Air 10000 hz legend} {Previsão do tempo Marcos Valle Previsao do tempo}]

}
