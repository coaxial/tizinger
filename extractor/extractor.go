// Package extractor defines the interface for an Extractor
package extractor

import "github.com/coaxial/tizinger/playlist"

// An Extractor fetches historical playlist data from a source to return
// playlist data that can be further parsed by Tizinger.
type Extractor interface {
	Playlist(timestampFrom int64) (playlist.Tracklist, error)
}
