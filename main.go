package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	type Youtube struct {
		Typename string `json:"__typename"`
		Link     string `json:"link"`
		Image    string `json:"image"`
	}
	type ExternalLinks struct {
		Typename string  `json:"__typename"`
		Youtube  Youtube `json:"youtube"`
	}
	type Node struct {
		Typename string `json:"__typename"`
		UUID     string `json:"uuid"`
		// the song's title is under the subtitle key
		Title         string        `json:"subtitle"`
		StartTime     int           `json:"start_time"`
		EndTime       int           `json:"end_time"`
		Cover         string        `json:"cover"`
		Label         string        `json:"label"`
		Album         string        `json:"album"`
		Interpreters  []string      `json:"interpreters"`
		MusicalKind   string        `json:"musical_kind"`
		Year          int           `json:"year"`
		ExternalLinks ExternalLinks `json:"external_links"`
		// the artist's name is under the title key
		Artist string `json:"title"`
	}
	type Edges struct {
		Typename string `json:"__typename"`
		Node     Node   `json:"node"`
		Cursor   string `json:"cursor"`
	}
	type PageInfo struct {
		Typename    string `json:"__typename"`
		EndCursor   string `json:"endCursor"`
		HasNextPage bool   `json:"hasNextPage"`
	}
	type TimelineCursor struct {
		Typename   string   `json:"__typename"`
		TotalCount int      `json:"totalCount"`
		Edges      []Edges  `json:"edges"`
		PageInfo   PageInfo `json:"pageInfo"`
	}
	type Data struct {
		TimelineCursor TimelineCursor `json:"timelineCursor"`
	}
	type FipHistoryResponse struct {
		Data Data `json:"data"`
	}

	uri := "https://www.fip.fr/latest/api/graphql?operationName=History&variables=%7B%22first%22%3A10%2C%22after%22%3A%22MTU5MjQ2NzMyNA%3D%3D%22%2C%22stationId%22%3A7%2C%22preset%22%3A%22192x192%22%7D&extensions=%7B%22persistedQuery%22%3A%7B%22version%22%3A1%2C%22sha256Hash%22%3A%22f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270%22%7D%7D"

	response, err := http.Get(uri)

	if err != nil {
		log.Fatal(err)
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
		fmt.Printf("Raw: %#v", v)
	}
}
