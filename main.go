package main

import (
	"fmt"
	"time"

	"github.com/coaxial/tizinger/extractor"
)

func main() {
	var fipClient extractor.Client
	list, _ := fipClient.Playlist(time.Date(2019, 07, 05, 0, 0, 0, 0, time.UTC).Unix(), 10)
	fmt.Printf("%#v", list)
}
