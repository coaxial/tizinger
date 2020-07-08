package main

import (
	"fmt"
	"os"
	"time"

	"github.com/coaxial/tizinger/fip"
	"github.com/coaxial/tizinger/tidal"
	"github.com/coaxial/tizinger/utils/logger"
)

func main() {
	var fipClient fip.APIClient
	var tidalClient tidal.APIClient
	var exitCode int
	errorWords := "without errors"

	ts := time.Now().AddDate(0, 0, -1) // 24h ago
	count := 300
	plName := fmt.Sprintf("FIP %d-%d-%d, %d tracks", ts.Year(), ts.Month(), ts.Day(), count)
	logger.Info.Printf("getting %d tracks as aired on FIP up until %s to Tidal", count, ts.Format("2006-01-02 15:04:05"))

	list, err := fipClient.Playlist(ts.Unix(), count)
	if err != nil {
		logger.Error.Printf("error getting tracks from fip: %v", err)
		errorWords = "with errors"
		exitCode = 1
	}

	err = tidalClient.CreatePlaylist(plName, list)
	if err != nil {
		logger.Error.Printf("error creating playlist %q on Tidal: %v", plName, err)
		errorWords = "with errors"
		exitCode = 1
	}
	logger.Info.Printf("done processing, %s", errorWords)
	os.Exit(exitCode)
}
