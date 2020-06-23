package main

import (
	"fmt"

	Extractors "github.com/coaxial/tizinger/extractors"
)

func main() {
	var fipExtractor Extractors.FipExtractor
	list, _ := fipExtractor.Playlist(0)
	fmt.Printf("%#v", list)
	// type Node struct {
	// 	// the song's title is under the subtitle key
	// 	Title       string `json:"subtitle"`
	// 	StartTime   int    `json:"start_time"`
	// 	EndTime     int    `json:"end_time"`
	// 	Album       string `json:"album"`
	// 	MusicalKind string `json:"musical_kind"`
	// 	Year        int    `json:"year"`
	// 	// the artist's name is under the title key
	// 	Artist string `json:"title"`
	// }
	// type Edges struct {
	// 	Node   Node   `json:"node"`
	// 	Cursor string `json:"cursor"`
	// }
	// type PageInfo struct {
	// 	EndCursor   string `json:"endCursor"`
	// 	HasNextPage bool   `json:"hasNextPage"`
	// }
	// type TimelineCursor struct {
	// 	Edges    []Edges  `json:"edges"`
	// 	PageInfo PageInfo `json:"pageInfo"`
	// }
	// type Data struct {
	// 	TimelineCursor TimelineCursor `json:"timelineCursor"`
	// }
	// type FipHistoryResponse struct {
	// 	Data Data `json:"data"`
	// }

	// // endpoint := "https://www.fip.fr/latest/api/graphql?operationName=History"
	// first := 10
	// from := time.Now().Unix()
	// timestamp := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(from, 10)))

	// // timestamp := "MTU5MjQ2NzMyNA=="
	// stationId := 7
	// req, err := http.NewRequest("GET", "https://www.fip.fr/latest/api/graphql", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// query := req.URL.Query()
	// query.Add("operationName", "History")
	// query.Add("variables", fmt.Sprintf("{\"first\":%d,\"after\":\"%s\",\"stationId\":%d}", first, timestamp, stationId))
	// query.Add("extensions", "{\"persistedQuery\":{\"version\":1,\"sha256Hash\":\"f8f404573583a6a9410cd24637f214a0b93038696c1d20f19202111b51fd8270\"}}")
	// req.URL.RawQuery = query.Encode()

	// client := &http.Client{}
	// response, err := client.Do(req)

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if response.StatusCode != http.StatusOK {
	// 	fmt.Printf("%#v", response)
	// }

	// defer response.Body.Close()

	// responseData, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// var responseObject FipHistoryResponse
	// json.Unmarshal(responseData, &responseObject)

	// for _, v := range responseObject.Data.TimelineCursor.Edges {
	// 	println("----")
	// 	fmt.Printf("At: %s\n", time.Unix(int64(v.Node.StartTime), 0))
	// 	fmt.Printf("Title: %s\n", v.Node.Title)
	// 	fmt.Printf("Artiste: %s\n", v.Node.Artist)
	// 	fmt.Printf("Album: %s\n", v.Node.Album)
	// 	fmt.Printf("Raw: %#v\n", v)
	// }

}
