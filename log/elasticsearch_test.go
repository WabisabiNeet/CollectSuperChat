package log

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	jsoniter "github.com/json-iterator/go"
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

func Test4(tt *testing.T) {
	channel := "aaa"

	err := SendChat(channel, "messageID", `{"videoInfo":{"cid":"UC1CfXB_kRs3C-zaeTG3oGyg","ctitle":"Haato Channel 赤井はあと","vid":"8t8oUT8crfM","vtitle":"Let's play Dark Souls REMASTERED!!!","scheduledStartTime":"2019-12-08T10:00:00+09:00","actualStartTime":"2019-12-08T10:01:32+09:00"},"message":{"messageID":"CjsKGkNOdVQzcV8zcE9ZQ0ZZaS1nZ29kYVkwRElBEh1DS1hKbm92c3BPWUNGUzRDdHdBZDBUOEhzdzEyOA%3D%3D","type":"TextMessage","authorName":"Badger","authorChannelID":"UCW26VvYINAEy6wnean5wcRw","userComment":"WOW PRO GAMER","publishedAt":"2019-12-08T10:47:25+09:00"}}`)
	if err != nil {
		tt.Fatal(err)
	}
}

func Test5(tt *testing.T) {
	cfg := elasticsearch.Config{}
	cfg.Addresses = append(cfg.Addresses, "http://192.168.10.11:9200")

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	builder := strings.Builder{}
	const data = "testdata/superchat0/superchat-2019-12-23T10-23-25.673.txt"
	file, err := os.Open(data)
	if err != nil {
		tt.Fatal(data)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		builder.WriteString(fmt.Sprintln(`{ "index" : {} }`))
		builder.WriteString(fmt.Sprintln(str))
	}

	buld := esapi.BulkRequest{
		Index: "test1",
		Body:  strings.NewReader(builder.String()),
	}
	res, err := buld.Do(context.Background(), es)
	if err != nil {
		tt.Fatal(err)
	}
	if res.StatusCode >= http.StatusBadRequest {
		tt.Fatal(res.StatusCode)
	}

}

func Test6(tt *testing.T) {
	cfg := elasticsearch.Config{}
	cfg.Addresses = append(cfg.Addresses, "http://192.168.10.11:9200")

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	filepath.Walk("testdata/superchat0", func(path string, info os.FileInfo, err error) error {
		builder := strings.Builder{}

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fmt.Println(path)
		file, err := os.Open(path)
		if err != nil {
			log.Fatal(path)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			str := scanner.Text()
			builder.WriteString(fmt.Sprintln(`{ "index" : {} }`))
			builder.WriteString(fmt.Sprintln(str))
		}

		buld := esapi.BulkRequest{
			Index: "chatdata",
			Body:  strings.NewReader(builder.String()),
		}
		res, err := buld.Do(context.Background(), es)
		if err != nil {
			tt.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode >= http.StatusBadRequest {
			tt.Fatal(res.StatusCode)
		}

		// time.Sleep(30 * time.Second)
		return nil
	})
}

func Test7(tt *testing.T) {
	cfg := elasticsearch.Config{}
	cfg.Addresses = append(cfg.Addresses, "http://192.168.10.11:9200")

	es2, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	es = es2

	outputs := []string{
		`{"key1":"val1"}`,
		`{"key2":"val2"}`,
	}
	SendChats(outputs)
}

func Test8(tt *testing.T) {
	cfg := elasticsearch.Config{}
	cfg.Addresses = append(cfg.Addresses, "http://192.168.10.11:9200")

	es2, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	es = es2

	updateScript := fmt.Sprintf("ctx._source.videoInfo.vtitle = '%v'; ctx._source.videoInfo.actualStartTime = '%v';", "テストテスト4", "2020-02-19T21:00:02+09:00")
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"videoInfo.vid.keyword": "b6Y3ERpaKtg",
			},
		},
		"script": map[string]interface{}{
			"source": updateScript,
		},
	}
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		tt.Fatal(err)
	}

	res, err := es.UpdateByQuery(
		[]string{"chat*"},
		es.UpdateByQuery.WithBody(&buf),
	)
	if err != nil {
		tt.Fatal(err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		fmt.Println(res)
		tt.Fatal(res.StatusCode)
	}
}

func Test9(tt *testing.T) {
	cfg := elasticsearch.Config{}
	cfg.Addresses = append(cfg.Addresses, "http://192.168.10.11:9200")

	es2, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	es = es2

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"videoInfo.vid.keyword": "b6Y3ERpaKtg",
			},
		},
	}
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		tt.Fatal(err)
	}

	start := time.Now()
	for i := 0; i < 10000; i++ {
		search(tt, i, &buf)
	}
	duration := time.Now().Sub(start)
	fmt.Println(fmt.Sprintf("%v", duration))
}

func search(tt *testing.T, i int, buf *bytes.Buffer) {
	res, err := es.Search(
		es.Search.WithBody(buf),
	)
	if err != nil {
		tt.Fatal(err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		fmt.Println(res)
		tt.Fatalf("%v:%v", i, res.StatusCode)
	}
}

func Test10(tt *testing.T) {
	cfg := elasticsearch.Config{}
	cfg.Addresses = append(cfg.Addresses, "http://192.168.10.11:9200")

	es2, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	es = es2

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"terms": map[string]interface{}{
				"message.messageID.keyword": []string{
					"CjoKGkNPM1ZvT1RFeS1ZQ0ZjeGNtQW9kZ3hzRGV3EhxDSUNtMTZYRXktWUNGZlBHVEFJZHFSd013dy0y",
					"CjoKGkNPbWJqT1RFeS1ZQ0ZjY1BaQW9kaGxBQkpBEhxDTnpjMmNPX3ktWUNGUVVpS2dvZHhsWUc1UTE3",
				},
			},
		},
	}
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		tt.Fatal(err)
	}

	res, err := es.DeleteByQuery(
		[]string{"chat*"},
		&buf,
	)
	if err != nil {
		tt.Fatal(err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		fmt.Println(res)
		tt.Fatal(res.StatusCode)
	}
}

func Test11(tt *testing.T) {
	cfg := elasticsearch.Config{}
	cfg.Addresses = append(cfg.Addresses, "http://192.168.10.11:9200")

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	filepath.Walk("testdata/superchat0", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fmt.Println(path)
		file, err := os.Open(path)
		if err != nil {
			log.Fatal(path)
		}
		defer file.Close()

		ids := []string{}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			str := scanner.Text()
			var rr ChatMessage
			if err := json.NewDecoder(strings.NewReader(str)).Decode(&rr); err != nil {
				tt.Fatalf("Error parsing the response body: %s", err)
			}
			ids = append(ids, rr.Message.MessageID)
		}

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"terms": map[string]interface{}{
					"message.messageID.keyword": ids,
				},
			},
		}
		var buf bytes.Buffer
		if err = json.NewEncoder(&buf).Encode(query); err != nil {
			tt.Fatal(err)
		}

		res, err := es.DeleteByQuery(
			[]string{"chat*"},
			&buf,
		)
		if err != nil {
			tt.Fatal(err)
		}

		if res.StatusCode >= http.StatusBadRequest {
			fmt.Println(res)
			tt.Fatal(res.StatusCode)
		}

		res1, err1 := es.Indices.Refresh()
		if err1 != nil {
			tt.Fatal(err)
		}

		if res1.StatusCode >= http.StatusBadRequest {
			fmt.Println(res1)
			tt.Fatal(res1.StatusCode)
		}

		// time.Sleep(30 * time.Second)
		return nil
	})
}

// ChatMessage : output json.
type ChatMessage struct {
	VideoInfo struct {
		ChannelID          string `json:"cid,omitempty"`
		ChannelTitle       string `json:"ctitle,omitempty"`
		VideoID            string `json:"vid,omitempty"`
		VideoTitle         string `json:"vtitle,omitempty"`
		ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
		ActualStartTime    string `json:"actualStartTime,omitempty"`
	} `json:"videoInfo,omitempty"`
	Message struct {
		MessageID          string `json:"messageID,omitempty"`
		MessageType        string `json:"type,omitempty"`
		AuthorName         string `json:"authorName,omitempty"`
		IsModerator        bool   `json:"isModerator,omitempty"`
		IsMember           bool   `json:"isMember,omitempty"`
		IsOwner            bool   `json:"isOwner,omitempty"`
		AccessibilityLabel string `json:"accessibilityLabel,omitempty"`
		AuthorChannelID    string `json:"authorChannelID,omitempty"`
		UserComment        string `json:"userComment,omitempty"`
		PublishedAt        string `json:"publishedAt,omitempty"`

		AmountDisplayString string  `json:"amountDisplayString,omitempty"`
		CurrencyRateToJPY   float64 `json:"currencyRateToJPY,omitempty"`
		AmountJPY           uint    `json:"amountJPY,omitempty"`
		Currency            string  `json:"currency,omitempty"`
	} `json:"message,omitempty"`
}

var benchJSON = `{"videoInfo":{"cid":"UCdn5BQ06XqgXoAxIhbqw5Rg","ctitle":"フブキCh。白上フブキ","vid":"mRjCGiIsTNo","vtitle":"クリスマス？なにそれ？お焚き上げ？【#白上お焚き上げ】","scheduledStartTime":"2019-12-23T21:00:00+09:00"},"message":{"messageID":"CkUKGkNKZS1fc0tteS1ZQ0ZRRTlaQW9kN1ZZSkdBEidDTkNOZ0wybXktWUNGWFZRaFFvZFVWVURDQTE1NzcwODc5NzI1MjQ%3D","type":"TextMessage","authorName":"雅白雪","authorChannelID":"UC7uDFnYcl9jATBLdRVy3qdg","userComment":"待機","publishedAt":"2019-12-23T16:59:32+09:00"}}`

func BenchmarkJson1(bb *testing.B) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var data ChatMessage
	err := json.UnmarshalFromString(benchJSON, &data)
	if err != nil {
		bb.Fatal(err)
	}

	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		j, _ := json.Marshal(data)
		_ = string(j)
	}
}

func BenchmarkJson2(bb *testing.B) {
	var jsonitr = jsoniter.ConfigCompatibleWithStandardLibrary
	var data ChatMessage
	err := jsonitr.UnmarshalFromString(benchJSON, &data)
	if err != nil {
		bb.Fatal(err)
	}

	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		j, _ := json.Marshal(data)
		_ = string(j)
	}
}
