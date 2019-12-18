package log_test

import (
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/log"
)

func TestOutputSuperChat(tt *testing.T) {
	defer log.SyncSuerChat()
	for i := 0; i < 1000000; i++ {
		log.OutputSuperChat(`{"key":"value"}`)
	}
}
