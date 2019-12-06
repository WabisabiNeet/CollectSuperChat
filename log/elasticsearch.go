package log

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/antonholmquist/jason"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
)

var (
	es *elasticsearch.Client
)

func init() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elasticsearch:9200",
		},
	}
	var err error
	es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		Fatal(err.Error())
	}
}

// SendChat send chat message to Elasticsearch
func SendChat(channelID, messageID, jsonStr string) error {
	// searchReq := esapi.SearchRequest{
	// 	Index: []string{channelID},
	// 	Query: fmt.Sprintf(`"match":{"message.messageID":"%d"}}`, messageID),
	// }

	indexReq := esapi.IndexRequest{
		Index: "flare_test",               // Index name
		Body:  strings.NewReader(jsonStr), // Document body
		// DocumentID: "1",                    // Document ID
		Refresh: "true", // Refresh
	}

	res, err := indexReq.Do(context.Background(), es)
	if err != nil {
		return errors.Wrap(err, "SendChat error.")

	}
	defer res.Body.Close()

	return nil
}

func searchMessage(channelID, messageID string) (string, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"message.messageID": "CjoKGkNJWDFsODdSbS1ZQ0Zkclp3UW9kNndjTFpREhxDTUtLenNYUm0tWUNGVXBMS2dvZGxDNEg3QS0w",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return "", err
	}

	searchReq := esapi.SearchRequest{
		Index: []string{"flare_test"},
		Body:  &buf,
	}

	res, err := searchReq.Do(context.Background(), es)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	root, err := jason.NewObjectFromReader(res.Body)
	if err != nil {
		return "", err
	}

	hits, err := root.GetObjectArray("hits", "hits")
	if err != nil {
		return "", err
	}
	if len(hits) == 0 {
		return "", nil
	}

	// docid, err := hits[0].GetString("_id")
	// if err != nil {
	// 	return false, err
	// }
	return "", nil
}
