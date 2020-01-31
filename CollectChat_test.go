package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/youtube/v3"
)

func Test(tt *testing.T) {
	ch := make(chan string, 1)
	ch <- "aaa"
	_, ok := <-ch
	if !ok {
		fmt.Println("error1")
		return
	}

	close(ch)
	_, ok = <-ch
	if ok {
		fmt.Println("error2")
		return
	}
}

func Test2(tt *testing.T) {
	u, _ := url.Parse("https://www.youtube.com/live_chat?is_popout=1")
	q := u.Query()
	q.Add("v", "aaa")
	u.RawQuery = q.Encode()

	fmt.Println(u.String())
}

func Test3(tt *testing.T) {
	b := getCredentials()
	// If modifying these scopes, delete your previously saved token.json.
	config := getConfig(b)
	client := getClient(config)

	ys, err := youtube.New(client)
	if err != nil {
		tt.Fatalf("Unable to retrieve Youtube client: %v", err)
	}

	sub := ys.Subscriptions.List("snippet")
	sub.Mine(true)
	sub.MaxResults(50)

	res, err := sub.Do()
	if err != nil {
		tt.Fatal(err)
	}

	fmt.Println(res.NextPageToken)
	for _, item := range res.Items {
		fmt.Println(fmt.Sprintf("%v %v", item.Snippet.ResourceId.ChannelId, item.Snippet.Title))
	}
}

func getCredentials() []byte {
	b, err := ioutil.ReadFile("credentials.json") // Download own credentials.json from google developer console.
	if err != nil {
		log.Fatal("Unable to read client secret file: %v", err)
	}

	return b
}

func getConfig(b []byte) *oauth2.Config {
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatal("Unable to parse client secret file to config: %v", err)
	}
	return config
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatal("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatal("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	log.Info("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// func TestGetLiveStreamID(tt *testing.T) {
// 	ctx := context.Background()
// 	ys, err := youtube.NewService(ctx, option.WithAPIKey(""))
// 	if err != nil {
// 		dbglog.Fatal(err.Error())
// 	}
// 	c := Collector{
// 		ID:              "1",
// 		YoutubeService:  ys,
// 		ProcessingCount: 0,
// 	}

// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)
// 	c.StartWatch(wg, "MzgvKiKYGHM")
// }

// StartWatch collect super chat.
// func (c *Collector) StartWatch(wg *sync.WaitGroup, vid string) {
// 	defer c.decrementCount()
// 	defer wg.Done()
// 	if vid == "" {
// 		dbglog.Fatal("vid is nil")
// 	}

// 	quit := make(chan os.Signal)
// 	defer close(quit)
// 	signal.Notify(quit, os.Interrupt)

// 	// e.g. https://www.youtube.com/watch?v=WziZomD9KC8
// 	videoInfo, err := livestream.GetLiveInfo(c.YoutubeService, vid)
// 	if err != nil {
// 		dbglog.Warn(fmt.Sprintf("[vid:%v] %v", vid, err))
// 		return
// 	} else if videoInfo.LiveStreamingDetails.ActiveLiveChatId == "" {
// 		dbglog.Info(fmt.Sprintf("[vid:%v] Live chat not active.", vid))
// 		return
// 	}

// 	chatlog := initSuperChatLogger(videoInfo.Snippet.ChannelId)
// 	defer chatlog.Sync()

// 	nextToken := ""
// 	var intervalMillis int64 = 60 * 1000
// 	for {
// 		messages, next, requireIntervalMillis, err := livestream.GetSuperChatRawMessages(c.YoutubeService, videoInfo.LiveStreamingDetails.ActiveLiveChatId, nextToken)
// 		if err != nil {
// 			switch t := err.(type) {
// 			case *googleapi.Error:
// 				if t.Code == 403 && t.Message == "The live chat is no longer live." {
// 					dbglog.Info(t.Message)
// 					return
// 				}
// 			default:
// 				dbglog.Fatal(t.Error())
// 			}
// 		}

// 		for _, message := range messages {
// 			c := livestream.ChatMessage{
// 				ChannelID:          videoInfo.Snippet.ChannelId,
// 				ChannelTitle:       videoInfo.Snippet.ChannelTitle,
// 				VideoID:            videoInfo.Id,
// 				VideoTitle:         videoInfo.Snippet.Title,
// 				ScheduledStartTime: videoInfo.LiveStreamingDetails.ScheduledStartTime,
// 				ActualStartTime:    videoInfo.LiveStreamingDetails.ActualStartTime,
// 			}

// 			// 暫定 日本円の場合だけ金額を入れる
// 			// 他国通貨の場合は要件等…
// 			var amount uint64
// 			var currency string
// 			switch message.Snippet.Type {
// 			case "superChatEvent":
// 				c.AmountDisplayString = message.Snippet.SuperChatDetails.AmountDisplayString
// 				c.AmountMicros = message.Snippet.SuperChatDetails.AmountMicros
// 				c.Currency = message.Snippet.SuperChatDetails.Currency
// 				c.Tier = message.Snippet.SuperChatDetails.Tier
// 				c.UserComment = message.Snippet.SuperChatDetails.UserComment

// 				amount = message.Snippet.SuperChatDetails.AmountMicros / 1000 / 1000
// 				currency = message.Snippet.SuperChatDetails.Currency
// 			case "superStickerEvent":
// 				c.AmountDisplayString = message.Snippet.SuperStickerDetails.AmountDisplayString
// 				c.AmountMicros = message.Snippet.SuperStickerDetails.AmountMicros
// 				c.Currency = message.Snippet.SuperStickerDetails.Currency
// 				c.Tier = message.Snippet.SuperStickerDetails.Tier

// 				amount = message.Snippet.SuperStickerDetails.AmountMicros / 1000 / 1000
// 				currency = message.Snippet.SuperStickerDetails.Currency
// 			}
// 			if currency == "JPY" {
// 				c.AmountJPY = amount
// 			}

// 			message.AuthorDetails.ChannelUrl = ""
// 			message.AuthorDetails.ProfileImageUrl = ""
// 			message.Etag = ""
// 			message.Id = ""
// 			message.Kind = ""
// 			message.Snippet.LiveChatId = ""
// 			message.Snippet.DisplayMessage = ""
// 			message.Snippet.AuthorChannelId = ""
// 			message.Snippet.SuperChatDetails = nil
// 			message.Snippet.SuperStickerDetails = nil
// 			c.Message = message

// 			outputJSON, err := json.Marshal(c)
// 			if err == nil {
// 				chatlog.Info(string(outputJSON))
// 			}
// 		}

// 		nextToken = next
// 		if nextToken == "" {
// 			dbglog.Info("nextToken is empty.")
// 			break
// 		}

// 		// switch countPerMinute := int64(len(messages)) * 60000 / intervalMillis; { // コメント分速からインターバル時間を決定
// 		// case countPerMinute > 1800:
// 		// 	// API取得上限を超えそうな場合は分速から必要とされる時間の2/3
// 		// 	intervalMillis = 60 * livestream.MaxMessageCount / countPerMinute * 1000 * 2 / 3
// 		// case countPerMinute > 1200:
// 		// 	intervalMillis = 60 * 1000
// 		// case countPerMinute > 800:
// 		// 	intervalMillis = 120 * 1000
// 		// case countPerMinute > 500:
// 		// 	intervalMillis = 180 * 1000
// 		// default:
// 		// 	intervalMillis = 240 * 1000
// 		// }
// 		// // Youtubeから指示されたInterval以下にはしない
// 		// if intervalMillis < requireIntervalMillis {
// 		// 	intervalMillis = requireIntervalMillis
// 		// }
// 		// intervalMillis = 10000
// 		intervalMillis = requireIntervalMillis
// 		// TODO: live開始10分前までは10分とかでいいかも
// 		// TODO: live開始直後はコメント集中しやすいからデフォルトを短縮してもいいかも

// 		dbglog.Info(fmt.Sprintf("[%v] messageCount:%v interval: %vms.", videoInfo.Snippet.ChannelId, len(messages), intervalMillis))
// 		select {
// 		case <-time.Tick(time.Duration(intervalMillis) * time.Millisecond):
// 		case <-quit:
// 			dbglog.Info(fmt.Sprintf("[%v] signaled.", videoInfo.Snippet.ChannelId))
// 			return
// 		}
// 	}
// }

// func getLiveStreamID(ys *youtube.Service, channel string, sig chan os.Signal) (string, error) {
// 	t := time.NewTicker(time.Minute)
// 	for {
// 		vid, err := livestream.GetLiveStreamID(ys, channel)
// 		if err != nil {
// 			if err.Error() != "live stream not found" {
// 				return "", err
// 			}
// 			select {
// 			case <-t.C:
// 				dbglog.Info(fmt.Sprintf("[%v] repert getLiveStreamID.", channel))
// 				continue
// 			case <-sig:
// 				return "", fmt.Errorf("signaled")
// 			}
// 		}

// 		return vid, nil
// 	}
// }

func TestChromedp2(tt *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	chromedp.ListenTarget(
		ctx,
		func(ev interface{}) {
			if ev, ok := ev.(*network.EventResponseReceived); ok {
				// fmt.Println(fmt.Sprintf("event received:%v", ev.Response.URL))
				// fmt.Println(ev.Type)

				if ev.Type != "XHR" {
					return
				}
				if !strings.Contains(ev.Response.URL, "get_live_chat") {
					return
				}

				go func() {
					// print response body
					c := chromedp.FromContext(ctx)
					rbp := network.GetResponseBody(ev.RequestID)
					body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
					if err != nil {
						fmt.Println(err)
					}
					fmt.Printf("%s\n", body)
				}()

			}
		},
	)

	//

	// navigate to a page, wait for an element, click
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate("https://www.youtube.com/live_chat?v=50NRqM8alh8&is_popout=1"),
		chromedp.Sleep(time.Second*15),
	)
	if err != nil {
		tt.Fatal(err)
	}
}
