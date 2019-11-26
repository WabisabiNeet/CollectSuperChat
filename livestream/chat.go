package livestream

import (
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
		AccessibilityLabel string `json:"accessibilityLabel,omitempty"`
		AuthorChannelID    string `json:"authorChannelID,omitempty"`
		UserComment        string `json:"userComment,omitempty"`
		PublishedAt        string `json:"publishedAt,omitempty"`

		AmountDisplayString string  `json:"amountDisplayString,omitempty"`
		CurrencyRateToJPY   float64 `json:"currencyRateToJPY,string,omitempty"`
		AmountJPY           uint    `json:"amountJPY,omitempty"`
		Currency            string  `json:"currency,omitempty"`

		RawMessage *youtube.LiveChatMessage `json:"rawMessage,omitempty"`
	} `json:"message,omitempty"`
}
