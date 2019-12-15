package notifier

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/jhillyerd/enmime"
	"github.com/pkg/errors"
	"google.golang.org/api/gmail/v1"
)

func getVideoIDsFromList(srv *gmail.Service, socialLabelID string) (vids []string, historyID uint64, err error) {
	messages, _ := srv.Users.Messages.List("me").LabelIds(socialLabelID).Do()
	for _, m := range messages.Messages {
		vid, his, err := getVideoIDfromMail(srv, m)
		if historyID < his {
			historyID = his
		}
		if err != nil {
			switch err.Error() {
			case "invalid live stream start time":
				return vids, historyID, nil
			case "not live stream mail":
				continue
			default:
				log.Error(err.Error())
				continue
			}
		}

		vids = append(vids, vid)
	}

	return vids, historyID, nil
}

func getVideoIDfromHistroy(srv *gmail.Service, h *gmail.History) (vids []string, historyID uint64, err error) {
	for _, hma := range h.MessagesAdded {
		vid, his, err := getVideoIDfromMail(srv, hma.Message)
		if historyID < his {
			historyID = his
		}
		if err != nil {
			switch err.Error() {
			case "invalid live stream start time":
				return vids, historyID, nil
			case "not live stream mail":
				continue
			default:
				log.Error(err.Error())
				continue
			}
		}

		vids = append(vids, vid)
	}

	return vids, historyID, nil
}

func getVideoIDfromMail(srv *gmail.Service, m *gmail.Message) (vid string, history uint64, err error) {
	mm, err := srv.Users.Messages.Get("me", m.Id).Format("raw").Do()
	if err != nil {
		return "", 0, err
	}

	// アーカイブが最大12時間だから、開始時は余裕もって13時間前までのメールをチェックする
	if time.Now().Add(time.Hour * -13).After(time.Unix(mm.InternalDate/1000, 0)) {
		return "", mm.HistoryId, fmt.Errorf("invalid live stream start time")
	}

	html, err := getLiveStreamHTML(mm.Raw)
	if err != nil {
		return "", mm.HistoryId, err
	}

	stringReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		return "", mm.HistoryId, errors.Wrap(err, html)
	}

	liveURL := ""
	sss := doc.Find("a")
	sss.EachWithBreak(func(_ int, s *goquery.Selection) bool {
		url, exists := s.Attr("href")
		if !exists || !strings.Contains(url, "watch") {
			return true
		}

		liveURL = url
		return false
	})

	vid, err = parseVideoID(liveURL)
	if err != nil {
		return "", mm.HistoryId, errors.Wrap(err, html)
	}

	return vid, mm.HistoryId, nil
}

func getLiveStreamHTML(src string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	enve, err := enmime.ReadEnvelope(strings.NewReader(string(decoded)))
	if err != nil {
		return "", err
	}

	subject := enve.GetHeader("Subject")
	log.Info(subject)
	if !strings.Contains(subject, "ライブ配信中です") && !strings.Contains(subject, "プレミア公開を開始しました") {
		return "", fmt.Errorf("not live stream mail")
	}

	return enve.HTML, nil
}

func parseVideoID(liveURL string) (string, error) {
	u, err := url.Parse(liveURL)
	if err != nil {
		return "", err
	}

	u, err = url.Parse(u.Query().Get("u"))
	if err != nil {
		return "", err
	}

	return u.Query().Get("v"), nil
}

// Gmail struct.
type Gmail struct {
	CollectChat func(vid string)
}

// PollingStart polling gmail.
func (n *Gmail) PollingStart(wg *sync.WaitGroup) {
	defer wg.Done()

	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	b := GetCredentials()
	// If modifying these scopes, delete your previously saved token.json.
	config := GetConfig(b)
	client := GetClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatal("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatal("Unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		log.Error("No labels found.")
		return
	}

	socialLabelID := ""
	for _, l := range r.Labels {
		if l.Name != "CATEGORY_SOCIAL" {
			continue
		}
		socialLabelID = l.Id
	}
	if socialLabelID == "" {
		log.Error("CATEGORY_SOCIAL can not found.")
		return
	}

	t := time.NewTicker(2 * time.Minute)
	defer t.Stop()
	for {
		vids, historyID, err := getVideoIDsFromList(srv, socialLabelID)
		if err != nil {
			continue
		}

		for _, v := range vids {
			n.CollectChat(v)
		}

		for {
			log.Info("history timer tick.")
			histroyRes, err := srv.Users.History.List("me").
				StartHistoryId(historyID).
				HistoryTypes("messageAdded").
				LabelId(socialLabelID).
				Do()
			if err != nil {
				continue
			}

			for _, h := range histroyRes.History {
				vids, his, err := getVideoIDfromHistroy(srv, h)
				if err != nil {
					continue
				}
				if historyID < his {
					historyID = his
				}

				for _, v := range vids {
					n.CollectChat(v)
				}
			}

			select {
			case <-t.C:
			case <-quit:
				return
			}

		}
	}
}
