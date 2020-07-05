// Package extractor defines the interface for an Extractor
package extractor

// A Client fetches historical playlist data from a source to return
// playlist data that can be further parsed by Tizinger.
type Client interface {
	Playlist(timestampFrom int64, tracksCount int) (Tracklist, error)
}
