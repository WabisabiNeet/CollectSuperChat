package livestream_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
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