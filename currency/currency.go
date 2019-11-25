package currency

import ()

// Currency is currency code & symbol set.
type Currency struct {
	Code   string
	Symbol string
}

// Currencies is currency  table
var Currencies = []*Currency{
	{Code: "JPY", Symbol: "¥"},    // 円
	{Code: "USD", Symbol: "$"},    // アメリカドル
	{Code: "CAD", Symbol: "C$"},   // カナダドル
	{Code: "EUR", Symbol: "€"},    // ユーロ
	{Code: "HKD", Symbol: "HK$"},  // 香港ドル
	{Code: "KRW", Symbol: "₩"},    // 韓国ウォン
	{Code: "TWD", Symbol: "NT$"},  // ニュー台湾ドル
	{Code: "AUD", Symbol: "A$"},   // オーストラリアドル
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
	{Code: "NZD", Symbol: "NZ$"},  // ニュージーランドドル
	// {Code: "CNY", Symbol: "¥"}, // 人民元
	// {Code: "PHP", Symbol: "₱"},// フィリピンペソ
}
