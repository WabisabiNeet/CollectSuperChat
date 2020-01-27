package currency

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
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
	{Code: "CAD", Symbol: "CA$"},  // カナダドル
	{Code: "EUR", Symbol: "€"},    // ユーロ
	{Code: "HKD", Symbol: "HK$"},  // 香港ドル
	{Code: "KRW", Symbol: "₩"},    // 韓国ウォン
	{Code: "TWD", Symbol: "NT$"},  // ニュー台湾ドル
	{Code: "AUD", Symbol: "A$"},   // オーストラリアドル
	{Code: "NZD", Symbol: "NZ$"},  // ニュージーランドドル
	{Code: "MXN", Symbol: "Mex$"}, // メキシコペソ
	{Code: "MXN", Symbol: "MX$"},  // メキシコペソ
	{Code: "BND", Symbol: "B$"},   // ブルネイドル
	{Code: "FJD", Symbol: "FJ$"},  // フィジードル
	{Code: "IDR", Symbol: "Rp"},   // インドネシアルピア
	{Code: "INR", Symbol: "₹"},    // インドルピー
	{Code: "INR", Symbol: "Rs."},  // インドルピー
	{Code: "SGD", Symbol: "S$"},   // シンガポールドル
	{Code: "SGD", Symbol: "SGD"},  // シンガポールドル
	{Code: "THB", Symbol: "฿"},    // タイバーツ
	{Code: "VND", Symbol: "₫"},    // ベトナムドン
	{Code: "CHF", Symbol: "CHF"},  // スイスフラン
	{Code: "GBP", Symbol: "£"},    // 英ポンド
	{Code: "BRL", Symbol: "R$"},   // ブラジルレアル
	{Code: "PEN", Symbol: "PEN"},  // ペルーソル
	{Code: "RUB", Symbol: "RUB"},  // ロシアルーブル
	{Code: "PHP", Symbol: "PHP"},  // フィリピンペソ
	{Code: "CLP", Symbol: "CLP"},  // チリペソ
	{Code: "ARS", Symbol: "ARS"},  // アルゼンチンペソ
	{Code: "PLN", Symbol: "PLN"},  // ポーランドズウォティ
	{Code: "SEK", Symbol: "SEK"},  // スウェーデンクローナ
	{Code: "HUF", Symbol: "HUF"},  // ハンガリーフォリント
	{Code: "NOK", Symbol: "NOK"},  // ノルウェークローネ
	{Code: "CZK", Symbol: "CZK"},  // チェココルナ
	{Code: "CRC", Symbol: "CRC"},  // コスタリカコロン
	// {Code: "CNY", Symbol: "¥"}, // 人民元
}

// ErrUnknownCurrency error
var (
	ErrUnknownCurrency = errors.New("unknown currency")
	ErrParseCurrency   = errors.New("currency parse fail")
)

// GetCurrency return Currency
func GetCurrency(amountStr string) (*Currency, error) {
	for _, c := range Currencies {
		if strings.HasPrefix(amountStr, c.Symbol) {
			return c, nil
		}
	}
	return nil, ErrUnknownCurrency
}

// ScrapeRataToJPY get currency rate to JPY
func (c *Currency) ScrapeRataToJPY() error {
	if c.Code == "JPY" {
		c.RateToJPY = 1
		return nil
	}

	u := fmt.Sprintf(feedURLBase, strings.ToLower(c.Code))
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(u)
	if err != nil {
		return err
	}
	items := feed.Items

	desc := ""
	for _, item := range items {
		desc = item.Description
	}

	baseRateStr := fmt.Sprintf("%s=", c.Code)
	rateWords := ""
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
		rateWords = text
		break
	}

	rateWords = strings.TrimRight(strings.ToUpper(rateWords), `JPY<BR/>`)
	strs := strings.SplitAfter(rateWords, baseRateStr)
	rateStr := strs[len(strs)-1]

	if s, err := strconv.ParseFloat(rateStr, 64); err == nil {
		c.RateToJPY = s
	} else {
		return errors.Wrap(ErrParseCurrency, fmt.Sprintf("rateStr[%v]", rateStr))
	}

	return nil
}

// GetAmountValue return amount value.
func (c *Currency) GetAmountValue(amountStr string) (float64, error) {
	valueStr := strings.TrimPrefix(amountStr, c.Symbol)
	valueStr = strings.TrimSpace(valueStr)
	valueStr = strings.ReplaceAll(valueStr, ",", "")
	s, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, err
	}
	return s, nil
}
