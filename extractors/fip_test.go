package extractors

import (
	"net/http"
	"testing"

	"github.com/coaxial/tizinger/playlist"
	"github.com/coaxial/tizinger/utils/mocks"
	"github.com/stretchr/testify/assert"
)

var extractor FipExtractor

func TestPlaylist(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json")
		historyJson := mocks.LoadFixture("../fixtures/fip/history_response.json")
		resp.Write(historyJson)
	}
	server := mocks.Server(handler)
	defer server.Close()
	expected := []playlist.Track{
		playlist.Track{Title: "Scar tissue", Artist: "Red Hot Chili Peppers", Album: "Greatest hits"},
		playlist.Track{Title: "Off the wall", Artist: "Jil Is Lucky", Album: "Off the wall"},
		playlist.Track{Title: "Kalimba (Flute mix)", Artist: "Freakniks", Album: "Electro tunes"},
		playlist.Track{Title: "Tsukikaage no rendezvous", Artist: "Keiko Mari", Album: "Nippon girls: Japanese pop, beat & bossa nova 1966-1970"},
		playlist.Track{Title: "Un petit poisson, un petit oiseau", Artist: "Juliette Greco", Album: "Déshabillez-moi 1965-1969"},
		playlist.Track{Title: "I want to be happy", Artist: "Ray Brown", Album: "Brown Ray trio / Some of my best friends are guitarists"},
		playlist.Track{Title: "I'm so happy I can't stop crying", Artist: "Sting", Album: "Mercury falling"},
		playlist.Track{Title: "Sambarilove (feat. Roubinho Jacobina)", Artist: "Chiara Civello", Album: "Eclipse"},
		playlist.Track{Title: "Retiens l'été", Artist: "Double Francoise", Album: "Les bijoux"},
		playlist.Track{Title: "Serenade nº13 en Sol Maj K 525 \"\"une petite musique de nuit\"\" : I. Allegro", Artist: "I Musici", Album: "Mozart, pachelbel, albinoni"},
	}
	extractor.SetEndpointUrl(server.URL)

	actual, _ := extractor.Playlist(0)

	assert.Equal(t, actual, expected, "should return a playlist")
}
