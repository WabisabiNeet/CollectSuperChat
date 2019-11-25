package livestream

import (
	"google.golang.org/api/youtube/v3"
)

// ChatMessage : output json.
type ChatMessage struct {
	ChannelID          string `json:"cid,omitempty"`
	ChannelTitle       string `json:"ctitle,omitempty"`
	VideoID            string `json:"vid,omitempty"`
	VideoTitle         string `json:"vtitle,omitempty"`
	ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	ActualStartTime    string `json:"actualStartTime,omitempty"`

	ID                  string `json:"id,omitempty"`
	MessageType         string `json:"type,omitempty"`
	AuthorName          string `json:"authorName,omitempty"`
	IsModerator         bool   `json:"isModerator,omitempty"`
	IsMember            bool   `json:"isMember,omitempty"`
	AccessibilityLabel  string `json:"accessibilityLabel,omitempty"`
	AuthorChannelID     string `json:"authorChannelID,omitempty"`
	AmountDisplayString string `json:"amountDisplayString,omitempty"`
	AmountJPY           uint   `json:"amountJPY,omitempty"`
	Currency            string `json:"currency,omitempty"`
	UserComment         string `json:"userComment,omitempty"`
	PublishedAt         string `json:"publishedAt,omitempty"`

	Message *youtube.LiveChatMessage `json:"chat,omitempty"`
}
