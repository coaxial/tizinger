package main

import (
	"fmt"
	"time"

	"github.com/coaxial/tizinger/fip"
)

func main() {
	var fipExtractor fip.Extractor
	// TODO use two days  ago at midnight
	list, _ := fipExtractor.Playlist(time.Now().Unix())
	fmt.Printf("%#v", list)
}
