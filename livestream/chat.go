package livestream

import (
	"fmt"

	"github.com/WabisabiNeet/CollectSuperChat/currency"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"
)

// ChatMessage : output json.
type ChatMessage struct {
	VideoInfo struct {
		ChannelID          string `json:"cid,omitempty"`
		ChannelTitle       string `json:"ctitle,omitempty"`
		VideoID            string `json:"vid,omitempty"`
		VideoTitle         string `json:"vtitle,omitempty"`
		ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
		ActualStartTime    string `json:"actualStartTime,omitempty"`
	} `json:"videoInfo,omitempty"`
	Message struct {
		MessageID          string `json:"messageID,omitempty"`
		MessageType        string `json:"type,omitempty"`
		AuthorName         string `json:"authorName,omitempty"`
		IsModerator        bool   `json:"isModerator,omitempty"`
		IsMember           bool   `json:"isMember,omitempty"`
		IsOwner            bool   `json:"isOwner,omitempty"`
		AccessibilityLabel string `json:"accessibilityLabel,omitempty"`
		AuthorChannelID    string `json:"authorChannelID,omitempty"`
		UserComment        string `json:"userComment,omitempty"`
		PublishedAt        string `json:"publishedAt,omitempty"`

		AmountDisplayString string  `json:"amountDisplayString,omitempty"`
		CurrencyRateToJPY   float64 `json:"currencyRateToJPY,omitempty"`
		AmountJPY           uint    `json:"amountJPY,omitempty"`
		Currency            string  `json:"currency,omitempty"`

		RawMessage *youtube.LiveChatMessage `json:"rawMessage,omitempty"`
	} `json:"message,omitempty"`
}

// ConvertToJPY convert to JPY
func (m *ChatMessage) ConvertToJPY() error {
	cur, err := currency.GetCurrency(m.Message.AmountDisplayString)
	if err != nil {
		return errors.Wrap(err, m.Message.AmountDisplayString)
	}
	if cur.RateToJPY == 0 {
		return fmt.Errorf("RateToJPY == 0 [%+v]", cur)
	}
	m.Message.CurrencyRateToJPY = cur.RateToJPY
	m.Message.Currency = cur.Code

	val, err := cur.GetAmountValue(m.Message.AmountDisplayString)
	if err != nil {
		return errors.Wrap(err, m.Message.AmountDisplayString)
	}

	m.Message.AmountJPY = uint(val * cur.RateToJPY)
	return nil
}
