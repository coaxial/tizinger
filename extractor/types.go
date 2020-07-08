package extractor

// Track represents a music track's metadata
type Track struct {
	Title  string
	Artist string
	Album  string
}

// Tracklist is the list of tracks played
type Tracklist []Track
