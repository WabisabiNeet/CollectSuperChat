package chromedp_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/chromedp"
	"github.com/WabisabiNeet/CollectSuperChat/livestream"
)

func TestChomedp1(t *testing.T) {
	if err := chromedp.InitChrome(); err != nil {
		t.Fatal(err)
	}
	defer chromedp.TerminateChrome()

	vid1 := ""
	vid2 := ""
	w1, err := chromedp.OpenLiveChatWindow(vid1)
	if err != nil {
		t.Fatal(err)
	}
	w2, err := chromedp.OpenLiveChatWindow(vid2)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	fn := func(ctx context.Context, vid string, w <-chan string) {
		file, err := os.Create(fmt.Sprintf("%v.txt", vid))
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case json := <-w:
				messages, finished, err := livestream.GetLiveChatMessagesFromProxy(json)
				if err != nil {
					t.Fatal(err)
				}
				if finished {
					return
				}

				for _, m := range messages {
					_, err := file.WriteString(m.Message.MessageID)

					if err != nil {
						t.Fatal(err)
						return
					}
				}
			}
		}
	}

	fn(ctx, vid1, w1)
	fn(ctx, vid1, w2)

	time.Sleep(time.Second * 20)
	cancel()
}
