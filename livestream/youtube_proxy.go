package livestream

import (
	"fmt"
	"strings"

	"github.com/antonholmquist/jason"
)

// GetLiveChatMessagesFromProxy scrape live chat
func GetLiveChatMessagesFromProxy(chatJSON string) ([]*ChatMessage, bool, error) {
	root, err := jason.NewObjectFromReader(strings.NewReader(chatJSON))
	if err != nil {
		return nil, true, err
	}

	finished := false
	_, err = root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "continuations")
	if err != nil {
		// chat end.
		finished = true
	}

	// for archive???
	// finished := false
	// for _, continuation := range continuations {
	// 	_, err := continuation.GetString("liveChatReplayContinuationData", "continuation")
	// 	if err != nil {
	// 		finished = true
	// 		break
	// 	}
	// }
	// if finished {
	// 	return
	// }

	messages := []*ChatMessage{}

	actions, err := root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "actions")
	if err != nil {
		// no chat.
		return messages, finished, nil
	}

	for _, action := range actions {
		item, err := action.GetObject("addChatItemAction", "item")
		if err != nil {
			continue
		}

		m := item.Map()
		if _, ok := m["liveChatTextMessageRenderer"]; ok {
			message, err := getLiveChatTextMessage(item)
			if err != nil {
				dbglog.Info(fmt.Sprintf("liveChatTextMessageRenderer error : %v", err))
				continue
			}
			messages = append(messages, message)
		} else if _, ok := m["liveChatPaidMessageRenderer"]; ok {
			message, err := getLiveChatPaidMessage(item)
			if err != nil {
				dbglog.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
				continue
			}
			messages = append(messages, message)
		} else if _, ok := m["liveChatPaidStickerRenderer"]; ok {
			message, err := getLiveChatPaidStickerMessage(item)
			if err != nil {
				dbglog.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
				continue
			}
			messages = append(messages, message)
		}
	}

	return messages, finished, nil
}

func getLiveChatTextMessage(item *jason.Object) (*ChatMessage, error) {
	mr, err := item.GetObject("liveChatTextMessageRenderer")
	if err != nil {
		return nil, err
	}

	m := &ChatMessage{}
	m, err = getCommonMessageInfo(mr, m)
	if err != nil {
		return nil, err
	}

	authorBadges, err := mr.GetObjectArray("authorBadges") // メンバー/モデレーター
	isModerator := false
	accessibilityLabel := ""
	for _, badge := range authorBadges {
		authorBadgeRenderer, err := badge.GetObject("liveChatAuthorBadgeRenderer")
		if err != nil {
			continue
		}

		iconType, err := authorBadgeRenderer.GetString("icon", "iconType")
		if err == nil {
			switch iconType {
			case "MODERATOR":
				isModerator = true
			default:
				dbglog.Info(fmt.Sprintf("unexpected iconType:%v", iconType))
			}

			continue
		}

		label, err := authorBadgeRenderer.GetString("accessibility", "accessibilityData", "label")
		if err == nil {
			if label != "" {
				accessibilityLabel = label
			}
		}
	}

	m.Message.MessageType = "TextMessage"
	m.Message.IsModerator = isModerator
	m.Message.AccessibilityLabel = accessibilityLabel

	return m, nil
}

func getLiveChatPaidMessage(item *jason.Object) (*ChatMessage, error) {
	mr, err := item.GetObject("liveChatPaidMessageRenderer")
	if err != nil {
		return nil, err
	}

	m := &ChatMessage{}
	m, err = getCommonMessageInfo(mr, m)
	if err != nil {
		return nil, err
	}

	purchase, err := mr.GetString("purchaseAmountText", "simpleText") // 金額(通貨記号付き)
	if err != nil {
		return nil, err
	}
	m.Message.MessageType = "PaidMessage"
	m.Message.AmountDisplayString = purchase
	m.Message.AmountJPY = 0
	m.Message.Currency = ""

	return m, nil
}

func getLiveChatPaidStickerMessage(item *jason.Object) (*ChatMessage, error) {
	mr, err := item.GetObject("liveChatPaidStickerRenderer")
	if err != nil {
		return nil, err
	}

	m := &ChatMessage{}
	m, err = getCommonMessageInfo(mr, m)
	if err != nil {
		return nil, err
	}

	purchase, err := mr.GetString("purchaseAmountText", "simpleText") // 金額(通貨記号付き)
	if err != nil {
		return nil, err
	}

	m.Message.MessageType = "PaidMessage-Sticker"
	m.Message.AmountDisplayString = purchase
	m.Message.AmountJPY = 0
	m.Message.Currency = ""

	return m, nil
}

func getCommonMessageInfo(renderer *jason.Object, message *ChatMessage) (*ChatMessage, error) {
	id, err := renderer.GetString("id")
	if err != nil {
		return nil, err
	}

	runs, err := renderer.GetObjectArray("message", "runs")
	messageStr := ""
	if err == nil {
		for _, r := range runs {
			text, _ := r.GetString("text") //表示メッセージ
			messageStr += text
		}
		messageStr, _ = runs[0].GetString("text") //表示メッセージ
	}
	author, err := renderer.GetString("authorName", "simpleText") //名前
	if err != nil {
		return nil, err
	}
	timestamp, err := renderer.GetString("timestampUsec") //タイムスタンプ(UnixEpoch)
	if err != nil {
		return nil, err
	}
	autherChannelID, err := renderer.GetString("authorExternalChannelId") //投稿者チャンネルID
	if err != nil {
		return nil, err
	}

	message.Message.MessageID = id
	message.Message.AuthorName = author
	message.Message.AuthorChannelID = autherChannelID
	message.Message.UserComment = messageStr
	message.Message.PublishedAt = timestamp

	return message, nil
}
