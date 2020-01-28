package livestream

import (
	"fmt"
	"strconv"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/mattn/go-jsonpointer"
)

// GetLiveChatMessagesFromProxy2 scrape live chat
func GetLiveChatMessagesFromProxy2(chatJSON string) ([]*ChatMessage, bool, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var data interface{}

	err := json.UnmarshalFromString(chatJSON, &data)
	if err != nil {
		return nil, true, err
	}

	finished := false
	finished = !jsonpointer.Has(data, "/response/continuationContents/liveChatContinuation/continuations")

	messages := []*ChatMessage{}

	iactions, err := jsonpointer.Get(data, "/response/continuationContents/liveChatContinuation/actions")
	if err != nil {
		// no chat.
		return messages, finished, nil
	}
	if _, ok := iactions.([]interface{}); !ok {
		return messages, finished, nil
	}

	for _, action := range iactions.([]interface{}) {
		item, err := jsonpointer.Get(action, "/addChatItemAction/item")
		if err != nil {
			continue
		}

		if jsonpointer.Has(item, "/liveChatTextMessageRenderer") {
			message, err := getLiveChatTextMessage2(&item)
			if err != nil {
				log.Info(fmt.Sprintf("liveChatTextMessageRenderer error : %v", err))
				continue
			}
			messages = append(messages, message)
		} else if jsonpointer.Has(item, "/liveChatPaidMessageRenderer") {
			message, err := getLiveChatPaidMessage2(&item)
			if err != nil {
				log.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
				continue
			}
			messages = append(messages, message)
		} else if jsonpointer.Has(item, "/liveChatPaidStickerRenderer") {
			message, err := getLiveChatPaidStickerMessage2(&item)
			if err != nil {
				log.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
				continue
			}
			messages = append(messages, message)
		} else if jsonpointer.Has(item, "/liveChatMembershipItemRenderer") {
			message, err := getLiveChatMembershipMessage2(&item)
			if err != nil {
				log.Info(fmt.Sprintf("liveChatMembershipItemRenderer error : %v", err))
				continue
			}
			messages = append(messages, message)
		}
	}

	return messages, finished, nil
}

// GetReplayChatMessagesFromProxy2 scrape live chat
// func GetReplayChatMessagesFromProxy2(chatJSON string) ([]*ChatMessage, bool, error) {
// 	root, err := jason.NewObjectFromReader(strings.NewReader(chatJSON))
// 	if err != nil {
// 		return nil, true, err
// 	}

// 	finished := false
// 	continuations, err := root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "continuations")
// 	if err != nil {
// 		// chat end.
// 		finished = true
// 	}
// 	existsLiveChatReplayContinuationData := false
// 	for _, continuation := range continuations {
// 		_, err := continuation.GetObject("liveChatReplayContinuationData")
// 		if err == nil {
// 			existsLiveChatReplayContinuationData = true
// 			break
// 		}
// 	}
// 	if !existsLiveChatReplayContinuationData {
// 		finished = true
// 	}

// 	messages := []*ChatMessage{}

// 	actions, err := root.GetObjectArray("response", "continuationContents", "liveChatContinuation", "actions")
// 	if err != nil {
// 		// no chat.
// 		return messages, finished, nil
// 	}

// 	for _, action := range actions {
// 		actions2, err := action.GetObjectArray("replayChatItemAction", "actions")
// 		if err != nil {
// 			continue
// 		}
// 		for _, action2 := range actions2 {
// 			item, err := action2.GetObject("addChatItemAction", "item")
// 			if err != nil {
// 				continue
// 			}

// 			m := item.Map()
// 			if _, ok := m["liveChatTextMessageRenderer"]; ok {
// 				message, err := getLiveChatTextMessage(item)
// 				if err != nil {
// 					log.Info(fmt.Sprintf("liveChatTextMessageRenderer error : %v", err))
// 					continue
// 				}
// 				messages = append(messages, message)
// 			} else if _, ok := m["liveChatPaidMessageRenderer"]; ok {
// 				message, err := getLiveChatPaidMessage(item)
// 				if err != nil {
// 					log.Info(fmt.Sprintf("liveChatPaidMessageRenderer error : %v", err))
// 					continue
// 				}
// 				messages = append(messages, message)
// 			} else if _, ok := m["liveChatPaidStickerRenderer"]; ok {
// 				message, err := getLiveChatPaidStickerMessage(item)
// 				if err != nil {
// 					log.Info(fmt.Sprintf("liveChatPaidStickerRenderer error : %v", err))
// 					continue
// 				}
// 				messages = append(messages, message)
// 			} else if _, ok := m["liveChatMembershipItemRenderer"]; ok {
// 				message, err := getLiveChatMembershipMessage(item)
// 				if err != nil {
// 					log.Info(fmt.Sprintf("liveChatMembershipItemRenderer error : %v", err))
// 					continue
// 				}
// 				messages = append(messages, message)
// 			}
// 		}
// 	}

// 	return messages, finished, nil
// }

func getLiveChatTextMessage2(item *interface{}) (*ChatMessage, error) {
	mr, err := jsonpointer.Get(*item, "/liveChatTextMessageRenderer")
	if err != nil {
		return nil, err
	}

	m := &ChatMessage{}
	m, err = getCommonMessageInfo2(&mr, m)
	if err != nil {
		return nil, err
	}

	isModerator := false
	isOwner := false
	accessibilityLabel := ""

	iauthorBadges, err := jsonpointer.Get(mr, "/authorBadges") // メンバー/モデレーター
	if _, ok := iauthorBadges.([]interface{}); err == nil && ok {
		for _, badge := range iauthorBadges.([]interface{}) {
			iconType, err := jsonpointer.Get(badge, "/liveChatAuthorBadgeRenderer/icon/iconType")
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

			label, err := jsonpointer.Get(badge, "/liveChatAuthorBadgeRenderer/accessibility/accessibilityData/label")
			if err == nil {
				if label != "" {
					accessibilityLabel = label.(string)
				}
			}
		}
	}

	m.Message.MessageType = "TextMessage"
	m.Message.IsModerator = isModerator
	m.Message.IsOwner = isOwner
	m.Message.AccessibilityLabel = accessibilityLabel

	return m, nil
}

func getLiveChatPaidMessage2(item *interface{}) (*ChatMessage, error) {
	mr, err := jsonpointer.Get(*item, "/liveChatPaidMessageRenderer")
	if err != nil {
		return nil, err
	}

	m := &ChatMessage{}
	m, err = getCommonMessageInfo2(&mr, m)
	if err != nil {
		return nil, err
	}

	purchase, err := jsonpointer.Get(mr, "/purchaseAmountText/simpleText") // 金額(通貨記号付き)
	if err != nil {
		return nil, err
	}
	m.Message.MessageType = "PaidMessage"
	m.Message.AmountDisplayString = purchase.(string)
	m.Message.AmountJPY = 0
	m.Message.Currency = ""

	return m, nil
}

func getLiveChatPaidStickerMessage2(item *interface{}) (*ChatMessage, error) {
	mr, err := jsonpointer.Get(*item, "/liveChatPaidStickerRenderer")
	if err != nil {
		return nil, err
	}

	m := &ChatMessage{}
	m, err = getCommonMessageInfo2(&mr, m)
	if err != nil {
		return nil, err
	}

	purchase, err := jsonpointer.Get(mr, "/purchaseAmountText/simpleText") // 金額(通貨記号付き)
	if err != nil {
		return nil, err
	}

	m.Message.MessageType = "PaidMessage-Sticker"
	m.Message.AmountDisplayString = purchase.(string)
	m.Message.AmountJPY = 0
	m.Message.Currency = ""

	return m, nil
}

func getLiveChatMembershipMessage2(item *interface{}) (*ChatMessage, error) {
	mr, err := jsonpointer.Get(*item, "/liveChatTextMessageRenderer")
	// mr, err := item.GetObject("liveChatTextMessageRenderer")
	if err != nil {
		return nil, err
	}

	m := &ChatMessage{}
	m, err = getCommonMessageInfo2(&mr, m)
	if err != nil {
		return nil, err
	}

	isModerator := false
	isOwner := false
	accessibilityLabel := ""

	iauthorBadges, err := jsonpointer.Get(mr, "/authorBadges") // メンバー/モデレーター
	if _, ok := iauthorBadges.([]interface{}); err == nil && ok {
		for _, badge := range iauthorBadges.([]interface{}) {
			iconType, err := jsonpointer.Get(badge, "/liveChatAuthorBadgeRenderer/icon/iconType")
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

			label, err := jsonpointer.Get(badge, "/liveChatAuthorBadgeRenderer/accessibility/accessibilityData/label")
			if err == nil {
				if label != "" {
					accessibilityLabel = label.(string)
				}
			}
		}
	}

	m.Message.MessageType = "MembershipMessage"
	m.Message.IsModerator = isModerator
	m.Message.IsOwner = isOwner
	m.Message.AccessibilityLabel = accessibilityLabel

	return m, nil
}

func getCommonMessageInfo2(renderer *interface{}, message *ChatMessage) (*ChatMessage, error) {
	id, err := jsonpointer.Get(*renderer, "/id")
	if err != nil {
		return nil, err
	}

	messageStr := ""
	iruns, err := jsonpointer.Get(*renderer, "/message/runs")
	if _, ok := iruns.([]interface{}); err == nil && ok {
		for _, r := range iruns.([]interface{}) {
			text, _ := jsonpointer.Get(r, "/text") //表示メッセージ
			messageStr += text.(string)
		}
	}

	author, err := jsonpointer.Get(*renderer, "/authorName/simpleText") //名前
	if err != nil {
		return nil, err
	}

	timestamp, err := jsonpointer.Get(*renderer, "/timestampUsec") //タイムスタンプ(UnixEpoch)
	if err != nil {
		return nil, err
	}
	publishedAtMicroSec, err := strconv.ParseInt(timestamp.(string), 10, 64)
	var publishedAtSec int64 = 0
	if err == nil {
		publishedAtSec = publishedAtMicroSec / 1000 / 1000
	}

	autherChannelID, err := jsonpointer.Get(*renderer, "/authorExternalChannelId") //投稿者チャンネルID
	if err != nil {
		return nil, err
	}

	message.Message.MessageID = id.(string)
	message.Message.AuthorName = author.(string)
	message.Message.AuthorChannelID = autherChannelID.(string)
	message.Message.UserComment = messageStr
	message.Message.PublishedAt = time.Unix(publishedAtSec, 0).In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339)

	return message, nil
}
