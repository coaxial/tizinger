package main

import (
	"fmt"
	"time"

	"github.com/coaxial/tizinger/fip"
)

func main() {
	var fipClient fip.APIClient
	list, _ := fipClient.Playlist(time.Now().Unix())
	fmt.Printf("%#v", list)
}
