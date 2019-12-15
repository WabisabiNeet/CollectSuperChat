package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func main() {
	cfg := elasticsearch.Config{}
	cfg.CloudID = ""
	cfg.Username = ""
	cfg.Password = ""

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	filepath.Walk("superchat", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fmt.Println(path)
		sendfile(es, path)

		return nil
	})
}

func sendfile(es *elasticsearch.Client, path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(path)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		send(es, str)
	}
}

func send(es *elasticsearch.Client, str string) {
	req := esapi.IndexRequest{
		Index: "chatdata",             // Index name
		Body:  strings.NewReader(str), // Document body
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatal(str)
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
}
