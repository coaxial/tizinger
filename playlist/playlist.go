package playlist

// Track represents a music track's metadata
type Track struct {
	Title  string
	Artist string
	Album  string
}

// Extractor types are unmarshalling historical playlist data from their
// sources into Track structs
type Extractor interface {
	Playlist(timestampFrom int64) ([]Track, error)
}
