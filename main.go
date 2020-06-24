package main

import (
	"fmt"

	Extractors "github.com/coaxial/tizinger/extractors"
)

func main() {
	var fipExtractor Extractors.FipExtractor
	list, _ := fipExtractor.Playlist(0)
	fmt.Printf("%#v", list)
}
