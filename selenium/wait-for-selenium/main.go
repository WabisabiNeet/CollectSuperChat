package main

import (
	"fmt"
	"time"

	"github.com/sclevine/agouti"
)

func main() {
	for {
		fmt.Println("connecting...")
		seleniumServer := "http://selenium:4444/wd/hub"
		options := []agouti.Option{
			agouti.Browser("chrome"),
		}
		// free proxy 43.245.216.189:8080
		page, err := agouti.NewPage(seleniumServer, options...)
		if err != nil {
			fmt.Println(fmt.Sprintf("connect error:%v", err))
			time.Sleep(time.Second * 3)
			continue
		}
		page.CloseWindow()
		page.Session().Delete()
		fmt.Println("connect success")
		return
	}
}
