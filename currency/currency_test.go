package currency

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"unicode"

	"github.com/WabisabiNeet/CollectSuperChat/currency"
	"github.com/mmcdole/gofeed"
)

func TestCurrencyFromLocalFile(tt *testing.T) {
	f, err := os.Open("./testdata/usd-jpy.xml")
	if err != nil {
		tt.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	fp := gofeed.NewParser()

	feed, _ := fp.Parse(bytes.NewReader(b))
	items := feed.Items

	desc := ""
	for _, item := range items {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.Description)

		desc = item.Description
	}

	usdRate := ""
	scanner := bufio.NewScanner(strings.NewReader(desc))
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		text = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, text)
		if !strings.Contains(text, "USD=") {
			continue
		}
		usdRate = text
		break
	}

	fmt.Println("usdRate:[" + usdRate + "]")
	usdRate = strings.TrimRight(strings.ToUpper(usdRate), `JPY<BR/>`)
	strs := strings.SplitAfter(usdRate, "USD=")

	fmt.Println()
	fmt.Println(strs[len(strs)-1])
}

func TestCurrencyFromRSS(tt *testing.T) {
	for _, c := range currency.Currencies {
		c.ScrapeRataToJPY()
		if c.RateToJPY == 0 {
			tt.Fatal("RateToJPY is 0.0")
		}
	}
}
