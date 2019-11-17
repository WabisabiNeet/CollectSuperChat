package livestream

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antonholmquist/jason"
)

const youtubeVideoURLBase = "https://www.youtube.com/watch"
const youtubeLiveChatURLBase = "https://www.youtube.com/live_chat"
const youtubeLiveChatReplayURLBase = "https://www.youtube.com/live_chat"

// ScrapeLiveStreamingChat scrape live chat
func ScrapeLiveStreamingChat(vid string) {
	url, _ := url.Parse(youtubeVideoURLBase)
	q := url.Query()
	q.Add("v", vid)
	url.RawQuery = q.Encode()

	vhtml, err := getHTMLString(url.String())
	if err != nil {
		dbglog.Error(err.Error())
		return
	}

	jsonStr, err := getytInitialData(vhtml)
	if err != nil {
		return
	}

	jsonStr = strings.TrimRight(jsonStr, ";")
	li := LiveChatInfo{}
	err = json.Unmarshal([]byte(jsonStr), &li)
	if err != nil {
		return
	}

	continuation := li.Contents.TwoColumnWatchNextResults.ConversationBar.LiveChatRenderer.Continuations[0].ReloadContinuationData.Continuation
	for continuation != "" {
		url, _ = url.Parse(youtubeLiveChatURLBase)
		q := url.Query()
		q.Add("continuation", continuation)
		url.RawQuery = q.Encode()

		chtml, err := getHTMLString(url.String())
		if err != nil {
			dbglog.Error(err.Error())
			return
		}

		jsonStr, err = getytInitialData(chtml)
		if err != nil {
			return
		}
		jsonStr = strings.TrimRight(jsonStr, ";")

		root, err := jason.NewObjectFromReader(strings.NewReader(jsonStr))
		if err != nil {
			return
		}

		continuations, err := root.GetObjectArray("continuationContents", "liveChatContinuation", "continuations")
		if err != nil {
			return
		}
		continuation, err = continuations[0].GetString("invalidationContinuationData", "continuation")
		if err != nil {
			break
		}

		actions, err := root.GetObjectArray("continuationContents", "liveChatContinuation", "actions")
		if err != nil {
			return
		}

		for _, action := range actions {
			// runs, err := action.GetObjectArray("addChatItemAction", "item", "liveChatTextMessageRenderer", "message", "runs")
			item, err := action.GetObject("addChatItemAction", "item")
			if err != nil {
				continue
			}

			m := item.Map()
			if _, ok := m["liveChatTextMessageRenderer"]; ok {
				mr, err := item.GetObject("liveChatTextMessageRenderer")
				runs, err := mr.GetObjectArray("message", "runs")
				message := ""
				if err == nil {
					message, _ = runs[0].GetString("text") //表示メッセージ
				}
				author, err := mr.GetString("authorName", "simpleText") //名前
				if err != nil {
					continue
				}
				timestamp, err := mr.GetString("timestampUsec") //タイムスタンプ(UnixEpoch)
				if err != nil {
					continue
				}
				autherChannelID, err := mr.GetString("authorExternalChannelId") //投稿者チャンネルID
				if err != nil {
					continue
				}

				fmt.Println(fmt.Sprintf("author:%v, timestamp:%v, channelID:%v, message:%v", author, timestamp, autherChannelID, message))
			} else if _, ok := m["liveChatPaidMessageRenderer"]; ok {
				mr, err := item.GetObject("liveChatPaidMessageRenderer")
				if err != nil {
					continue
				}
				runs, err := mr.GetObjectArray("message", "runs")
				message := ""
				if err == nil {
					message, _ = runs[0].GetString("text") //表示メッセージ
				}
				author, err := mr.GetString("authorName", "simpleText") //名前
				if err != nil {
					continue
				}
				timestamp, err := mr.GetString("timestampUsec") //タイムスタンプ(UnixEpoch)
				if err != nil {
					continue
				}
				autherChannelID, err := mr.GetString("authorExternalChannelId") //投稿者チャンネルID
				if err != nil {
					continue
				}
				purchase, err := mr.GetString("purchaseAmountText", "simpleText") // 金額(通貨記号付き)
				if err != nil {
					continue
				}
				fmt.Println(fmt.Sprintf("author:%v, timestamp:%v, channelID:%v, message:%v, purchase:%v", author, timestamp, autherChannelID, message, purchase))

			} else if _, ok := m["liveChatPaidStickerRenderer"]; ok {
				mr, err := item.GetObject("liveChatPaidStickerRenderer")
				if err != nil {
					continue
				}
				runs, err := mr.GetObjectArray("message", "runs")
				message := ""
				if err == nil {
					message, _ = runs[0].GetString("text") //表示メッセージ
				}
				author, err := mr.GetString("authorName", "simpleText") //名前
				if err != nil {
					continue
				}
				timestamp, err := mr.GetString("timestampUsec") //タイムスタンプ(UnixEpoch)
				if err != nil {
					continue
				}
				autherChannelID, err := mr.GetString("authorExternalChannelId") //投稿者チャンネルID
				if err != nil {
					continue
				}
				purchase, err := mr.GetString("purchaseAmountText", "simpleText") // 金額(通貨記号付き)
				if err != nil {
					continue
				}
				fmt.Println(fmt.Sprintf("author:%v, timestamp:%v, channelID:%v, message:%v, purchase:%v", author, timestamp, autherChannelID, message, purchase))
			}

			// c := ChatMessage{}
		}
		//liveChatPaidMessageRenderer
		//liveChatTextMessageRenderer
		//liveChatPaidStickerRenderer
	}
}

func getHTMLString(url string) (string, error) {
	c := http.DefaultClient

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	return string(byteArray), nil
}

func getytInitialData(html string) (string, error) {
	stringReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		return "", err
	}

	scripts := doc.Find("script")
	jsonStr := ""
	scripts.EachWithBreak(func(_ int, s *goquery.Selection) bool {
		script := strings.TrimSpace(s.Text())
		if !strings.Contains(script, `window["ytInitialData"]`) {
			return true
		}

		scanner := bufio.NewScanner(strings.NewReader(script))
		for bufSize := bufio.MaxScanTokenSize; !scanner.Scan() && scanner.Err() == bufio.ErrTooLong; bufSize *= 2 {
			if bufSize > 10*1024*1024 {
				return false
			}
			scanner = bufio.NewScanner(strings.NewReader(script))
			scanner.Buffer(make([]byte, bufSize), bufSize)
		}

		ytInitialData := scanner.Text()
		for !strings.Contains(ytInitialData, `window["ytInitialData"]`) && scanner.Scan() {
			ytInitialData = scanner.Text()
		}

		splits := strings.SplitAfter(ytInitialData, ` = `)
		if len(splits) < 2 {
			return true
		}

		jsonStr = splits[1]
		return false
	})

	if jsonStr == "" {
		return "", errors.New("json parse error?")
	}

	return jsonStr, nil
}
