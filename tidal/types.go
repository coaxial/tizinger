package tidal

import (
	"strings"
	"time"
)

// tidalTimestamp is the timestamp format Tidal uses in API responses.
type tidalTimestamp struct {
	time.Time
}

// UnmarshalJSON allows for unmarshalling timestramp strings for Tidal's JSON
// responses into a time.Time object.
func (t *tidalTimestamp) UnmarshalJSON(buf []byte) error {
	ts, err := time.Parse("2006-01-02T15:04:05.000-0700", strings.Trim(string(buf), `"`))
	if err != nil {
		return err
	}
	t.Time = ts
	return err
}

func (t tidalTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Time.Format("2006-01-02T15:04:05-0700") + `"`), nil
}

// loginResponse is the JSON object returned from a successful login request.
type loginResponse struct {
	SessionID   string `json:"sessionId"`
	CountryCode string `json:"countryCode"`
	UserID      int    `json:"userId"`
}

// trackResponse is the JSON object returned for a search request
type trackResponse struct {
	Limit              int      `json:"limit"`
	Offset             int      `json:"offset"`
	TotalNumberOfItems int      `json:"totalNumberOfItems"`
	Tracks             []tracks `json:"items"`
}

type artist struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
type artists struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
type album struct {
	ID         int         `json:"id"`
	Title      string      `json:"title"`
	Cover      string      `json:"cover"`
	VideoCover interface{} `json:"videoCover"`
}
type tracks struct {
	ID                   int            `json:"id"`
	Title                string         `json:"title"`
	Duration             int            `json:"duration"`
	ReplayGain           float64        `json:"replayGain"`
	Peak                 float64        `json:"peak"`
	AllowStreaming       bool           `json:"allowStreaming"`
	StreamReady          bool           `json:"streamReady"`
	StreamStartDate      tidalTimestamp `json:"streamStartDate"`
	PremiumStreamingOnly bool           `json:"premiumStreamingOnly"`
	TrackNumber          int            `json:"trackNumber"`
	VolumeNumber         int            `json:"volumeNumber"`
	Version              interface{}    `json:"version"`
	Popularity           int            `json:"popularity"`
	Copyright            string         `json:"copyright"`
	URL                  string         `json:"url"`
	Isrc                 string         `json:"isrc"`
	Editable             bool           `json:"editable"`
	Explicit             bool           `json:"explicit"`
	AudioQuality         string         `json:"audioQuality"`
	AudioModes           []string       `json:"audioModes"`
	Artist               artist         `json:"artist"`
	Artists              []artists      `json:"artists"`
	Album                album          `json:"album"`
}

type playlist struct {
	UUID           string `json:"uuid"`
	Title          string `json:"title"`
	NumberOfTracks int    `json:"numberOfTracks"`
	NumberOfVideos int    `json:"numberOfVideos"`
	Creator        struct {
		ID int `json:"id"`
	} `json:"creator"`
	Description     interface{}    `json:"description"`
	Duration        int            `json:"duration"`
	LastUpdated     tidalTimestamp `json:"lastUpdated"`
	Created         tidalTimestamp `json:"created"`
	Type            string         `json:"type"`
	PublicPlaylist  bool           `json:"publicPlaylist"`
	URL             string         `json:"url"`
	Image           string         `json:"image"`
	Popularity      int            `json:"popularity"`
	SquareImage     string         `json:"squareImage"`
	PromotedArtists []interface{}  `json:"promotedArtists"`
	LastItemAddedAt interface{}    `json:"lastItemAddedAt"`
}
