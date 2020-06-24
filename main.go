package main

import (
	"fmt"

	"github.com/coaxial/tizinger/extractors"
)

func main() {
	var fipExtractor extractors.FipExtractor
	list, _ := fipExtractor.Playlist(0)
	fmt.Printf("%#v", list)
}
