package main

import (
	"fmt"
	"time"

	"github.com/coaxial/tizinger/extractors"
)

func main() {
	var fipExtractor extractors.FipExtractor
	// TODO use two days  ago at midnight
	list, _ := fipExtractor.Playlist(time.Now().Unix())
	fmt.Printf("%#v", list)
}
