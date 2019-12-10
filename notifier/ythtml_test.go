package notifier_test

import (
	"fmt"
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/notifier"
)

func Test1(tt *testing.T) {
	h := notifier.YoutubeHTML{
		CollectChat: func(vid string) {
			fmt.Println(vid)
		},
	}

	h.PollingStart()
}
