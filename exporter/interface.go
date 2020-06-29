package exporter

import "github.com/coaxial/tizinger/playlist"

// Client defines the interface for an exporter.
type Client interface {
	CreatePlaylist(name string, tracks []playlist.Tracklist) (ok bool, err error)
}
