package livestream_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/antonholmquist/jason"
	jsoniter "github.com/json-iterator/go"
	"github.com/mattn/go-jsonpointer"
)

func TestGetLiveChatMessagesFromProxy(tt *testing.T) {
	f, err := os.Open("./testdata/live_01.txt")
	if err != nil {
		tt.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	messages, finished, err := livestream.GetLiveChatMessagesFromProxy(string(b))
	if err != nil {
		tt.Fatal(err)
	}
	if finished {
		tt.Fatal("this is not live finished data.")
	}
	if len(messages) == 0 {
		tt.Fatal("len(messages) == 0")
	}
}

func TestGetLiveChatMessagesFromProxy_LiveFinish(tt *testing.T) {
	f, err := os.Open("./testdata/live_02_finish.txt")
	if err != nil {
		tt.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	_, finished, err := livestream.GetLiveChatMessagesFromProxy(string(b))
	if err != nil {
		tt.Fatal(err)
	}
	if !finished {
		tt.Fatal("this is live finished data.")
	}
}

func TestGetReplayChatMessagesFromProxy(tt *testing.T) {
	f, err := os.Open("./testdata/arvhive_01.txt")
	if err != nil {
		tt.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	messages, finished, err := livestream.GetReplayChatMessagesFromProxy(string(b))
	if err != nil {
		tt.Fatal(err)
	}
	if len(messages) == 0 {
		tt.Fatal("len(messages) == 0")
	}
	if finished {
		tt.Fatal("this is not finished data.")
	}
}

func TestGetReplayChatMessagesFromProxy_finished(tt *testing.T) {
	f, err := os.Open("./testdata/arvhive_02_end.txt")
	if err != nil {
		tt.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	messages, finished, err := livestream.GetReplayChatMessagesFromProxy(string(b))
	if err != nil {
		tt.Fatal(err)
	}
	if len(messages) != 0 {
		tt.Fatal("len(messages) != 0")
	}
	if !finished {
		tt.Fatal("this is finished data.")
	}
}

func BenchmarkJason(bb *testing.B) {
	benchJSON := getTestJSON(bb)

	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		root, err := jason.NewObjectFromReader(strings.NewReader(benchJSON))
		if err != nil {
			bb.Fatal(err)
		}
		_, err = root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "continuations")
		actions, err := root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "actions")
		for _, action := range actions {
			_, err := action.GetObject("addChatItemAction", "item")
			if err != nil {
				continue
			}
			// item.Map()
		}
	}
}

func BenchmarkJsonIter(bb *testing.B) {
	benchJSON := getTestJSON(bb)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var data interface{}

	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		err := json.UnmarshalFromString(benchJSON, &data)
		if err != nil {
			bb.Fatal(err)
		}
		jsonpointer.Get(data, "/response/continuationContents/liveChatContinuation/continuations")
		actions, err := jsonpointer.Get(data, "/response/continuationContents/liveChatContinuation/actions")
		if err != nil {
			bb.Fatal(err)
		}
		if actions2, ok := actions.([]interface{}); ok {

			for _, action := range actions2 {
				_, err := jsonpointer.Get(action, "/addChatItemAction/item")
				if err != nil {
					continue
				}
			}
		} else {
			bb.Fatal("err")
		}

	}
}

func BenchmarkFromProxy1(bb *testing.B) {
	benchJSON := getTestJSON(bb)

	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		messages, finished, err := livestream.GetLiveChatMessagesFromProxy(benchJSON)
		if err != nil {
			bb.Fatal(err)
		}
		if finished {
			bb.Fatal("this is not live finished data.")
		}
		if len(messages) == 0 {
			bb.Fatal("len(messages) == 0")
		}
	}
}

func BenchmarkFromProxy2(bb *testing.B) {
	benchJSON := getTestJSON(bb)

	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		messages, finished, err := livestream.GetLiveChatMessagesFromProxy2(benchJSON)
		if err != nil {
			bb.Fatal(err)
		}
		if finished {
			bb.Fatal("this is not live finished data.")
		}
		if len(messages) == 0 {
			bb.Fatal("len(messages) == 0")
		}
	}
}

func getTestJSON(bb *testing.B) string {
	f, err := os.Open("./testdata/live_01.txt")
	if err != nil {
		bb.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		bb.Fatal(err)
	}
	return string(b)
}
