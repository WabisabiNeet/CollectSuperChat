package currency

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/mmcdole/gofeed"
)

// Currency is currency code & symbol set.
type Currency struct {
	Code      string
	Symbol    string
	RateToJPY float64
}

// e.g https://www.fx-exchange.com/jpy/usd.html
const feedURLBase = `https://www.fx-exchange.com/%s/jpy.xml`

// Currencies is currency  table
var Currencies = []*Currency{
	{Code: "JPY", Symbol: "¥"},    // 円
	{Code: "JPY", Symbol: "￥"},    // 円
	{Code: "USD", Symbol: "$"},    // アメリカドル
	{Code: "CAD", Symbol: "C$"},   // カナダドル
	{Code: "EUR", Symbol: "€"},    // ユーロ
	{Code: "HKD", Symbol: "HK$"},  // 香港ドル
	{Code: "KRW", Symbol: "₩"},    // 韓国ウォン
	{Code: "TWD", Symbol: "NT$"},  // ニュー台湾ドル
	{Code: "AUD", Symbol: "A$"},   // オーストラリアドル
	{Code: "NZD", Symbol: "NZ$"},  // ニュージーランドドル
	{Code: "MXN", Symbol: "Mex$"}, // メキシコペソ
	{Code: "BND", Symbol: "B$"},   // ブルネイドル
	{Code: "FJD", Symbol: "FJ$"},  // フィジードル
	{Code: "IDR", Symbol: "Rp"},   // インドネシアルピア
	{Code: "INR", Symbol: "Rs."},  // インドルピー
	{Code: "SGD", Symbol: "S$"},   // シンガポールドル
	{Code: "THB", Symbol: "฿"},    // タイバーツ
	{Code: "VND", Symbol: "₫"},    // ベトナムドン
	{Code: "CHF", Symbol: "CHF"},  // スイスフラン
	{Code: "GBP", Symbol: "£"},    // 英ポンド
	// {Code: "CNY", Symbol: "¥"}, // 人民元
	// {Code: "PHP", Symbol: "₱"},// フィリピンペソ
}

// ScrapeRataToJPY get currency rate to JPY
func (c *Currency) ScrapeRataToJPY() {
	if c.Code == "JPY" {
		c.RateToJPY = 1
		return
	}

	u := fmt.Sprintf(feedURLBase, strings.ToLower(c.Code))
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(u)
	items := feed.Items

	desc := ""
	for _, item := range items {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.Description)

		desc = item.Description
	}

	baseRateStr := fmt.Sprintf("%s=", c.Code)
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

		if !strings.Contains(text, baseRateStr) {
			continue
		}
		usdRate = text
		break
	}

	fmt.Println("usdRate:[" + usdRate + "]")
	usdRate = strings.TrimRight(strings.ToUpper(usdRate), `JPY<BR/>`)
	strs := strings.SplitAfter(usdRate, baseRateStr)

	rate := strs[len(strs)-1]
	fmt.Println()
	fmt.Println(rate)

	if s, err := strconv.ParseFloat(rate, 64); err == nil {
		c.RateToJPY = s
	} else {
		// エラーログ
	}
}
