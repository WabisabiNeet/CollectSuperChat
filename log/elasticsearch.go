package log

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
)

var (
	es *elasticsearch.Client

	// ElasticsearchCloudid is id
	ElasticsearchCloudid string

	// ElasticsearchUser is user
	ElasticsearchUser string

	// ElasticsearchPass is user
	ElasticsearchPass string
)

func init() {
	cloudid := os.Getenv("ELASTICSEARCH_CLOUDID")
	user := os.Getenv("ELASTICSEARCH_USER")
	pass := os.Getenv("ELASTICSEARCH_PASS")

	cfg := elasticsearch.Config{}
	if cloudid != "" {
		if user == "" || pass == "" {
			Fatal("use cloudid, but user or pass is nil. user[%v], pass[%v]", user, pass)
		}
		cfg.CloudID = cloudid
		cfg.Username = user
		cfg.Password = pass
	} else {
		cfg.Addresses = []string{
			"http://elasticsearch:9200",
			// "http://192.168.10.11:9200", // for debug
		}
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
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.New(fmt.Sprintf("status:[%v]", res.StatusCode))
		}
		return errors.New(fmt.Sprintf("status:[%v] body[%v]", res.StatusCode, string(b)))

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
	return fmt.Sprintf("%x", bs), nil
}
