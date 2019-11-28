package test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sclevine/agouti"
)

func TestSeleniumServer(tt *testing.T) {
	url := "http://0000:4444/wd/hub"
	options := []agouti.Option{
		agouti.Browser("chrome"),
		agouti.ChromeOptions(
			"args", []string{
				"--proxy-server=0000:8081",
			}),
	}
	// free proxy 43.245.216.189:8080
	page, _ := agouti.NewPage(url, options...)

	page.Navigate("https://yahoo.co.jp")
	defer page.CloseWindow()
	<-time.Tick(time.Second * 5)

	page.Screenshot("Screenshot01.png")
}

func TestScrapeLiveStreamingChat(tt *testing.T) {
	// ---------------------------------------------------------------

	fmt.Fprintf(os.Stderr, "*** 開始 ***\n")

	targetURL := "https://www.youtube.com/watch?v=wzA2HBKmbnw"

	cap := agouti.NewCapabilities()
	cap["loggingPrefs"] = agouti.Capabilities{
		"performance": "ALL",
		// "enablePage":  false,
		// "enableNetwork": false,
	}
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions(
			"args", []string{
				// "--headless",
				"--autoplay-policy=no-user-gesture-required",
			}),
		agouti.Debug,
		// agouti.Desired(cap),
		// agouti.ChromeOptions(
		// 	"perfLoggingPrefs", agouti.Capabilities{
		// 		"enableNetwork": true,
		// 		"enablePage":    false,
		// 	}),
	)

	err := driver.Start()
	defer driver.Stop()
	if err != nil {
		log.Printf("Failed to start driver: %v", err)
	}

	page, err := driver.NewPage(
		agouti.Browser("chrome"),
		agouti.Desired(cap),
		agouti.ChromeOptions(
			"perfLoggingPrefs", agouti.Capabilities{
				"enableNetwork": true,
				"enablePage":    false,
			}),
	)
	if err != nil {
		log.Printf("Failed to open page: %v", err)
	}

	err = page.Navigate(targetURL)
	if err != nil {
		log.Printf("Failed to navigate: %v", err)
	}

	limit := 1 * time.Minute
	for begin := time.Now(); time.Since(begin) < limit; {
		// // page.Screenshot(fmt.Sprintf("%v.png", time.Now().Format("20060102150405")))
		// iframe := page.FindByID("chatframe")
		// iframe.SwitchToFrame()

		// // comments := iframe.Find("div#items.style-scope.yt-live-chat-item-list-renderer").AllByClass("style-scope.yt-live-chat-item-list-renderer")
		// // div#items.style-scope.yt-live-chat-item-list-renderer .style-scope.yt-live-chat-item-list-renderer
		// comments := page.All("div#items.style-scope.yt-live-chat-item-list-renderer .style-scope.yt-live-chat-item-list-renderer")
		// count, err := comments.Count()
		// if err != nil {
		// 	fmt.Println(fmt.Sprintf("comennt.Count() error:%v", err))
		// }
		// fmt.Println(fmt.Sprintf("Count:%v", count))

		// for i := 0; i < count; i++ {
		// 	comment := comments.At(i)
		// 	id, err := comment.Attribute("id")
		// 	if err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	authorType, err := comment.Attribute("author-type")
		// 	if err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	author, err := comment.Find("span#author-name").Text()
		// 	if err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	message, err := comment.Find("span#message").Text()
		// 	if err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	t, err := time.Parse("15:04 PM", "9:08 PM")
		// 	if err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	elements, err := comment.Elements()
		// 	fmt.Println(fmt.Sprintf("id:%v authorType:%v author:%v message:%v time:%v elements:%v", id, authorType, author, message, t, elements))
		// }

		// // html2, _ := page.HTML()
		// // fmt.Println(html2)
		// fmt.Println("----------------------------------------------------")
		// time.Sleep(10 * time.Second)
		// page.SwitchToParentFrame()

		iframe := page.FindByID("chatframe")
		iframe.SwitchToFrame()
		logs, err := page.ReadAllLogs("performance")
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, log := range logs {
			fmt.Println(fmt.Sprintf("[%v] %v", log.Time, log.Message))
		}
		time.Sleep(5 * time.Second)
	}
	time.Sleep(5 * time.Second)

	fmt.Fprintf(os.Stderr, "*** 終了 ***\n")
}
