package selenium_test

import (
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/selenium"
)

func Test1(tt *testing.T) {
	selenium.OpenLiveChatWindow("Fth5LCkN0kw")
	defer selenium.CloseLiveChatWindow("Fth5LCkN0kw")
}
