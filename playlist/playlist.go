package playlist

type Track struct {
	Title  string
	Artist string
	Album  string
}

type Extractor interface {
	Playlist(timestampFrom int64) ([]Track, error)
}
