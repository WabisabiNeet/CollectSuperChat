package main

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
