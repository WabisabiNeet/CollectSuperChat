package log

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/antonholmquist/jason"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
)

var (
	r map[string]interface{}
)

func Test1(tt *testing.T) {
	infoRes, err := es.Info()
	if err != nil {
		tt.Fatalf("Error getting response: %s", err)
	}
	// Check response status
	if infoRes.IsError() {
		tt.Fatalf("Error: %s", infoRes.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(infoRes.Body).Decode(&r); err != nil {
		tt.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	fmt.Println(fmt.Sprintf("Client: %s", elasticsearch.Version))
	fmt.Println(fmt.Sprintf("Server: %v", r))

	req := esapi.IndexRequest{
		Index: "test5",                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          // Index name
		Body:  strings.NewReader(`{"videoInfo":{"cid":"UCvInZx9h3jC2JzsIzoOebWg","ctitle":"Flare Ch. 不知火フレア","vid":"lUHT0PM28_E","vtitle":"【マインクラフト】はあと先輩と海底探索！【ホロライブ/不知火フレア】","scheduledStartTime":"2019-12-04T09:00:00.000Z","actualStartTime":"2019-12-04T09:04:01.144Z"},"message":{"messageID":"ChwKGkNOajBycVRYbS1ZQ0ZRUkFmQW9kcnZnTTFR","type":"PaidMessage","authorName":"まつりのJS","authorChannelID":"UCP8vvKnKBM7Sb9lniOydVGQ","userComment":"❣ 🔥 ❣","publishedAt":"1575451810735008","amountDisplayString":"A$5.00","currencyRateToJPY":"108.515","currency":"USD"}}`), // Document body
		// DocumentID: "1",                                     // Document ID
		Refresh: "true", // Refresh
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	fmt.Println(res)
}

func Test2(tt *testing.T) {
	file, err := os.Open("testdata/chatdata.txt")
	if err != nil {
		tt.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()

		req := esapi.IndexRequest{
			Index: "flare_test",           // Index name
			Body:  strings.NewReader(str), // Document body
			// DocumentID: "1",                    // Document ID
			Refresh: "true", // Refresh
		}

		res, err := req.Do(context.Background(), es)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()
	}
}

func Test3(tt *testing.T) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"message.messageID": "CjoKGkNJWDFsODdSbS1ZQ0Zkclp3UW9kNndjTFpREhxDTUtLenNYUm0tWUNGVXBMS2dvZGxDNEg3QS0w",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		tt.Fatal(err)
	}

	searchReq := esapi.SearchRequest{
		Index: []string{"flare_test"},
		Body:  &buf,
	}

	res, err := searchReq.Do(context.Background(), es)
	if err != nil {
		tt.Fatal(err)
	}
	defer res.Body.Close()

	root, err := jason.NewObjectFromReader(res.Body)
	if err != nil {
		tt.Fatal(err)
	}

	hits, err := root.GetObjectArray("hits", "hits")
	if err != nil || len(hits) == 0 {
		tt.Fatal(err)
	}

	docid, err := hits[0].GetString("_id")
	if err != nil || len(hits) == 0 {
		tt.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("docid:%v", docid))
}
