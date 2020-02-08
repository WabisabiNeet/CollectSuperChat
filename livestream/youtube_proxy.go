package livestream

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/log"
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
	return nil, finished, nil

	// messages := []*ChatMessage{}

	// actions, err := root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "actions")
	// if err != nil {
	// 	// no chat.
	// 	return messages, finished, nil
	// }

	// for _, action := range actions {
	// 	item, err := action.GetObject("addChatItemAction", "item")
	// 	if err != nil {
	// 		continue
	// 	}

	// 	m := item.Map()
	// 	if _, ok := m["liveChatTextMessageRenderer"]; ok {
	// 		message, err := getLiveChatTextMessage(item)
	// 		if err != nil {
	// 			log.Info(fmt.Sprintf("liveChatTextMessageRenderer error : %v", err))
	// 			continue
	// 		}
	// 		messages = append(messages, message)
	// 	} else if _, ok := m["liveChatPaidMessageRenderer"]; ok {
	// 		message, err := getLiveChatPaidMessage(item)
	// 		if err != nil {
	// 			log.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
	// 			continue
	// 		}
	// 		messages = append(messages, message)
	// 	} else if _, ok := m["liveChatPaidStickerRenderer"]; ok {
	// 		message, err := getLiveChatPaidStickerMessage(item)
	// 		if err != nil {
	// 			log.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
	// 			continue
	// 		}
	// 		messages = append(messages, message)
	// 	} else if _, ok := m["liveChatMembershipItemRenderer"]; ok {
	// 		message, err := getLiveChatMembershipMessage(item)
	// 		if err != nil {
	// 			log.Info(fmt.Sprintf("liveChatMembershipItemRenderer error : %v", err))
	// 			continue
	// 		}
	// 		messages = append(messages, message)
	// 	}
	// }

	// return messages, finished, nil
}

// GetReplayChatMessagesFromProxy scrape live chat
func GetReplayChatMessagesFromProxy(chatJSON string) ([]*ChatMessage, bool, error) {
	root, err := jason.NewObjectFromReader(strings.NewReader(chatJSON))
	if err != nil {
		return nil, true, err
	}

	finished := false
	continuations, err := root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "continuations")
	if err != nil {
		// chat end.
		finished = true
	}
	existsLiveChatReplayContinuationData := false
	for _, continuation := range continuations {
		_, err := continuation.GetObject("liveChatReplayContinuationData")
		if err == nil {
			existsLiveChatReplayContinuationData = true
			break
		}
	}
	if !existsLiveChatReplayContinuationData {
		finished = true
	}

	messages := []*ChatMessage{}

	actions, err := root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "actions")
	if err != nil {
		// no chat.
		return messages, finished, nil
	}

	for _, action := range actions {
		actions2, err := action.GetObjectArray("replayChatItemAction", "actions")
		if err != nil {
			continue
		}
		for _, action2 := range actions2 {
			item, err := action2.GetObject("addChatItemAction", "item")
			if err != nil {
				continue
			}

			m := item.Map()
			if _, ok := m["liveChatTextMessageRenderer"]; ok {
				message, err := getLiveChatTextMessage(item)
				if err != nil {
					log.Info(fmt.Sprintf("liveChatTextMessageRenderer error : %v", err))
					continue
				}
				messages = append(messages, message)
			} else if _, ok := m["liveChatPaidMessageRenderer"]; ok {
				message, err := getLiveChatPaidMessage(item)
				if err != nil {
					log.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
					continue
				}
				messages = append(messages, message)
			} else if _, ok := m["liveChatPaidStickerRenderer"]; ok {
				message, err := getLiveChatPaidStickerMessage(item)
				if err != nil {
					log.Info(fmt.Sprintf("liveChatPaidStickerRenderer error : %v", err))
					continue
				}
				messages = append(messages, message)
			} else if _, ok := m["liveChatMembershipItemRenderer"]; ok {
				message, err := getLiveChatMembershipMessage(item)
				if err != nil {
					log.Info(fmt.Sprintf("liveChatMembershipItemRenderer error : %v", err))
					continue
				}
				messages = append(messages, message)
			}
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
	isOwner := false
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
			case "OWNER":
				isOwner = true
			case "VERIFIED":
			default:
				log.Warn(fmt.Sprintf("unexpected iconType:%v src[%v]", iconType, mr))
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
	m.Message.IsOwner = isOwner
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

func getLiveChatMembershipMessage(item *jason.Object) (*ChatMessage, error) {
	mr, err := item.GetObject("liveChatMembershipItemRenderer")
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
	isOwner := false
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
			case "OWNER":
				isOwner = true
			case "VERIFIED":
			default:
				log.Warn(fmt.Sprintf("unexpected iconType:%v src[%v]", iconType, mr))
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

	m.Message.MessageType = "MembershipMessage"
	m.Message.IsModerator = isModerator
	m.Message.IsOwner = isOwner
	m.Message.AccessibilityLabel = accessibilityLabel

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
	publishedAtMicroSec, err := strconv.ParseInt(timestamp, 10, 64)
	var publishedAtSec int64 = 0
	if err == nil {
		publishedAtSec = publishedAtMicroSec / 1000 / 1000
	}

	autherChannelID, err := renderer.GetString("authorExternalChannelId") //投稿者チャンネルID
	if err != nil {
		return nil, err
	}

	message.Message.MessageID = id
	message.Message.AuthorName = author
	message.Message.AuthorChannelID = autherChannelID
	message.Message.UserComment = messageStr
	message.Message.PublishedAt = time.Unix(publishedAtSec, 0).In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339)

	return message, nil
}
