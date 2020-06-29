package fip

// node contains the playlist individual tracks information from the API
type node struct {
	// the song's title is under the subtitle key
	Title       string `json:"subtitle"`
	StartTime   int    `json:"start_time"`
	EndTime     int    `json:"end_time"`
	Album       string `json:"album"`
	MusicalKind string `json:"musical_kind"`
	Year        int    `json:"year"`
	// the artist's name is under the title key
	Artist string `json:"title"`
}

// edges is a wrapper key from the API
type edges struct {
	Node   node   `json:"node"`
	Cursor string `json:"cursor"`
}

// pageInfo contains pagination info from the API
type pageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

// timelineCursor is a wrapper for pagination info and track info from the API
type timelineCursor struct {
	Edges    []edges  `json:"edges"`
	PageInfo pageInfo `json:"pageInfo"`
}

// data is a wrapper for timelineCursor fom the API
type data struct {
	TimelineCursor timelineCursor `json:"timelineCursor"`
}

// historyResponse contains the whole response from the API
type historyResponse struct {
	Data data `json:"data"`
}
