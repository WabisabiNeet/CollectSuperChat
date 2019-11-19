package notifier

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/jhillyerd/enmime"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var dbglog *zap.Logger

func init() {
	dbglog = log.GetLogger()
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
	dbglog.Info(fmt.Sprintf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL))

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		dbglog.Fatal(fmt.Sprintf("Unable to read authorization code: %v", err))
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		dbglog.Fatal(fmt.Sprintf("Unable to retrieve token from web: %v", err))
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
	dbglog.Info(fmt.Sprintf("Saving credential file to: %s\n", path))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		dbglog.Fatal(fmt.Sprintf("Unable to cache oauth token: %v", err))
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

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
				dbglog.Warn(err.Error())
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
				dbglog.Warn(err.Error())
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
	dbglog.Info(subject)
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
func (n *Gmail) PollingStart() {
	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	b, err := ioutil.ReadFile("credentials.json") // Download own credentials.json from google developer console.
	if err != nil {
		dbglog.Fatal(fmt.Sprintf("Unable to read client secret file: %v", err))
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		dbglog.Fatal(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		dbglog.Fatal(fmt.Sprintf("Unable to retrieve Gmail client: %v", err))
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		dbglog.Fatal(fmt.Sprintf("Unable to retrieve labels: %v", err))
	}
	if len(r.Labels) == 0 {
		dbglog.Error("No labels found.")
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
		dbglog.Info("CATEGORY_SOCIAL can not found.")
		return
	}

	for {
		vids, historyID, err := getVideoIDsFromList(srv, socialLabelID)
		if err != nil {
			continue
		}

		for _, v := range vids {
			n.CollectChat(v)
		}

		for {
			dbglog.Info("history timer tick.")
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
			case <-time.Tick(30 * time.Second):
			case <-quit:
				return
			}

		}
	}
}