package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	type Node struct {
		// the song's title is under the subtitle key
		Title       string `json:"subtitle"`
		StartTime   int    `json:"start_time"`
		EndTime     int    `json:"end_time"`
		Album       string `json:"album"`
		MusicalKind string `json:"musical_kind"`
		Year        int    `json:"year"`
		// the artist's name is under the title key
		Artist string `json:"title"`
	}
	type Edges struct {
		Node   Node   `json:"node"`
		Cursor string `json:"cursor"`
	}
	type PageInfo struct {
		EndCursor   string `json:"endCursor"`
		HasNextPage bool   `json:"hasNextPage"`
	}
	type TimelineCursor struct {
		Edges    []Edges  `json:"edges"`
		PageInfo PageInfo `json:"pageInfo"`
	}
	type Data struct {
		TimelineCursor TimelineCursor `json:"timelineCursor"`
	}
	type FipHistoryResponse struct {
		Data Data `json:"data"`
	}

	// endpoint := "https://www.fip.fr/latest/api/graphql?operationName=History"
	first := 10
	timestamp := "MTU5MjQ2NzMyNA=="
	stationId := 7
	req, err := http.NewRequest("GET", "https://www.fip.fr/latest/api/graphql", nil)
	if err != nil {
		log.Fatal(err)
	}
	query := req.URL.Query()
	query.Add("operationName", "History")
	query.Add("variables", fmt.Sprintf("{\"first\":%d,\"after\":\"%s\",\"stationId\":%d}", first, timestamp, stationId))
	query.Add("extensions", "{\"persistedQuery\":{\"version\":1,\"sha256Hash\":\"f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270\"}}")
	req.URL.RawQuery = query.Encode()
	fmt.Printf("%#v", req.URL.String())
	// uri := fmt.Sprintf(
	//   "https://www.fip.fr/latest/api/graphql?operationName=History&variables={\"first\":%v,\"after\"%%22%%3A%%22MTU5MjQ2NzMyNA%%3D%%3D%%22%%2C%%22stationId%%22%%3A7%%2C%%22preset%%22%%3A%%22192x192%%22%%7D&extensions=%%7B%%22persistedQuery%%22%%3A%%7B%%22version%%22%%3A1%%2C%%22sha256Hash%%22%%3A%%22f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270%22%7D%7D"
	//   first
	// )

	// response, err := http.Get(req)
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		fmt.Printf("%#v", response)
	}

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var responseObject FipHistoryResponse
	json.Unmarshal(responseData, &responseObject)

	for _, v := range responseObject.Data.TimelineCursor.Edges {
		println("----")
		fmt.Printf("Title: %s\n", v.Node.Title)
		fmt.Printf("Artiste: %s\n", v.Node.Artist)
		fmt.Printf("Album: %s\n", v.Node.Album)
		fmt.Printf("Raw: %#v\n", v)
	}

}
