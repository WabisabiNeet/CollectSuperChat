package main

import (
	"github.com/WabisabiNeet/CollectSuperChat/notifier"
)

func main() {
	b := notifier.GetCredentials()
	config := notifier.GetConfig(b)
	notifier.GetClient(config)
}
