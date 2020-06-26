package main

import (
	"fmt"
	"time"

	"github.com/coaxial/tizinger/extractor"
)

func main() {
	var fipExtractor extractor.FipExtractor
	// TODO use two days  ago at midnight
	list, _ := fipExtractor.Playlist(time.Now().Unix())
	fmt.Printf("%#v", list)
}
