package exporter

import "github.com/coaxial/tizinger/extractor"

// Client defines the interface for an exporter.
type Client interface {
	CreatePlaylist(name string, tracks []extractor.Tracklist) (ok bool, err error)
}
