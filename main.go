package main

import (
	"fmt"
	"time"

	"github.com/coaxial/tizinger/fip"
	"github.com/coaxial/tizinger/utils/logger"
)

func main() {
	var fipClient fip.APIClient
	ts := time.Date(2019, time.July, 5, 0, 0, 0, 0, time.UTC).Unix()
	count := 300
	list, err := fipClient.Playlist(ts, count)
	if err != nil {
		logger.Error.Fatal(err)
	}

	for _, v := range list {
		fmt.Printf("Artist: %q, Album: %q, Title: %q\n", v.Artist, v.Album, v.Title)
	}
}
