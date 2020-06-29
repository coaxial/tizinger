package main

import (
	"fmt"
	"time"

	"github.com/coaxial/tizinger/fip"
)

func main() {
	var fipExtractor fip.Extractor
	list, _ := fipExtractor.Playlist(time.Now().Unix())
	fmt.Printf("%#v", list)
}
