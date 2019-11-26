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
