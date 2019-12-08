package log

import (
	"context"
	"crypto/sha1"
	"net/http"
	"strings"

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
	index, err := getChannelHash(channelID)
	if err != nil {
		return errors.Wrap(err, "SendChat error.")
	}

	indexReq := esapi.IndexRequest{
		Index: index,                      // Index name
		Body:  strings.NewReader(jsonStr), // Document body
		// DocumentID: "1",                    // Document ID
		Refresh: "true", // Refresh
	}

	res, err := indexReq.Do(context.Background(), es)
	if err != nil {
		return errors.Wrap(err, "SendChat error.")

	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusBadRequest {
		return errors.Wrap(err, "SendChat error.")
	}

	return nil
}

func getChannelHash(channelID string) (string, error) {
	h := sha1.New()

	_, err := h.Write([]byte(channelID))
	if err != nil {
		return "", err
	}

	bs := h.Sum(nil)
	return string(bs), nil
}
