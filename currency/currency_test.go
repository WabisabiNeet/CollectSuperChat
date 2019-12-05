package currency_test

import (
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/currency"
)

func TestCurrencyFromRSS(tt *testing.T) {
	for _, c := range currency.Currencies {
		c.ScrapeRataToJPY()
		if c.RateToJPY == 0 {
			tt.Fatal("RateToJPY is 0.0")
		}
	}
}

func TestGetCurrency(tt *testing.T) {
	testdata := []struct {
		AmountStr    string
		ExpectedCode string
	}{
		{AmountStr: `¥5.00`, ExpectedCode: "JPY"},
		{AmountStr: `￥5.00`, ExpectedCode: "JPY"},
		{AmountStr: "$5.00", ExpectedCode: "USD"},
		{AmountStr: "C$5.00", ExpectedCode: "CAD"},
		{AmountStr: "€5.00", ExpectedCode: "EUR"},
		{AmountStr: "HK$5.00", ExpectedCode: "HKD"},
		{AmountStr: "₩5.00", ExpectedCode: "KRW"},
		{AmountStr: "NT$5.00", ExpectedCode: "TWD"},
		{AmountStr: "A$5.00", ExpectedCode: "AUD"},
		{AmountStr: "NZ$5.00", ExpectedCode: "NZD"},
		{AmountStr: "Mex$5.00", ExpectedCode: "MXN"},
		{AmountStr: "B$5.00", ExpectedCode: "BND"},
		{AmountStr: "FJ$5.00", ExpectedCode: "FJD"},
		{AmountStr: "Rp5.00", ExpectedCode: "IDR"},
		{AmountStr: "Rs.5.00", ExpectedCode: "INR"},
		{AmountStr: "S$5.00", ExpectedCode: "SGD"},
		{AmountStr: "฿5.00", ExpectedCode: "THB"},
		{AmountStr: "₫5.00", ExpectedCode: "VND"},
		{AmountStr: "CHF5.00", ExpectedCode: "CHF"},
		{AmountStr: "£5.00", ExpectedCode: "GBP"},
	}

	for _, d := range testdata {
		c, err := currency.GetCurrency(d.AmountStr)
		if err != nil {
			tt.Fatalf("%v AmountStr[%v]", err, d.AmountStr)
		}

		if c.Code != d.ExpectedCode {
			tt.Fatalf("unexpected code:[%v] expected:[%v]", c.Code, d.ExpectedCode)
		}
	}

}

func TestGetAmountValue(tt *testing.T) {
	testdata := []struct {
		AmountStr     string
		ExpectedValue float64
	}{
		{AmountStr: "¥1000", ExpectedValue: 1000},
		{AmountStr: "¥100.1", ExpectedValue: 100.1},
		{AmountStr: "¥1,000", ExpectedValue: 1000},
		{AmountStr: "¥1,000.1", ExpectedValue: 1000.1},
		{AmountStr: "¥1,000,000", ExpectedValue: 1000000},
		{AmountStr: "¥1,000,000.1", ExpectedValue: 1000000.1},
	}
	cur := currency.Currencies[0]
	for _, t := range testdata {
		val, err := cur.GetAmountValue(t.AmountStr)
		if err != nil {
			tt.Fatalf("%v AmountStr[%v]", err, t.AmountStr)
		}
		if val != t.ExpectedValue {
			tt.Fatalf("unexpected value:[%v] expected:[%v]", val, t.ExpectedValue)
		}
	}
}
