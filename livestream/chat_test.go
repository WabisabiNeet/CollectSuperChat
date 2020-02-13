package livestream_test

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	jsoniter "github.com/json-iterator/go"
)

func Test41(tt *testing.T) {
	// 重複排除

	json := jsoniter.ConfigCompatibleWithStandardLibrary

	var bmessages []*livestream.ChatMessage
	filepath.Walk("testdata/base", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Fatal(path)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			str := scanner.Text()
			var data livestream.ChatMessage
			err := json.UnmarshalFromString(str, &data)
			if err != nil {
				tt.Fatal(err)
			}
			bmessages = append(bmessages, &data)
		}

		return nil
	})

	var tmessages []*livestream.ChatMessage
	filepath.Walk("testdata/target", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Fatal(path)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			str := scanner.Text()
			var data livestream.ChatMessage
			err := json.UnmarshalFromString(str, &data)
			if err != nil {
				tt.Fatal(err)
			}
			tmessages = append(tmessages, &data)
		}

		return nil
	})

	for _, tmessage := range tmessages {
		exists := false
		for _, bmessage := range bmessages {
			if bmessage.Message.MessageID == tmessage.Message.MessageID {
				exists = true
				break
			}
		}

		if !exists {
			outputJSON, err := json.Marshal(tmessage)
			if err != nil {
				tt.Error(err)
			}
			log.OutputSuperChat(string(outputJSON))
		}
	}
}
