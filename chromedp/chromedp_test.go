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

	vids := []string{
		"ZMdAaib_Fqs",
		"fPkSVqBpFT8",
		"77mRtYwyBLQ",
		"PolG-apCNKs",
		"SKq8Pl_I9R0",
		"eAkw-k-OuVc",
		"opRmTs4WdEQ",
		"ss1a64BJZLI",
		"1VYQYJtQ0aA",
		"20f8nXpqp-g",
		"EmC19Lnv1Ag",
		"nmY7up0eJEU",
		"faQkGfooFWQ",
		"b3lr10YQ0D8",
		"b5jnwoRiYxg",
		"OyXDD1oXNTA",
		"V1E0jgAId1U",
		"EpgAZQYfm8s",
		"MQ7-28plUVQ",
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
				// messages, finished, err := livestream.GetLiveChatMessagesFromProxy(json)
				messages, finished, err := livestream.GetReplayChatMessagesFromProxy2(json)
				if err != nil {
					t.Fatal(err)
				}
				if finished {
					return
				}

				for _, m := range messages {
					_, err := file.WriteString(fmt.Sprintf("%v\n", m.Message.MessageID))

					if err != nil {
						t.Fatal(err)
						return
					}
				}
			}
		}
	}

	for _, vid := range vids {
		w2, err := chromedp.OpenLiveChatWindow(vid, true)
		if err != nil {
			t.Fatal(err)
		}
		go fn(ctx, vid, w2)
	}

	time.Sleep(time.Second * 60)
	cancel()
}
