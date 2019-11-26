// +build release

package log

import (
	"os"

	"github.com/getsentry/sentry-go"
)

func initSentry() {
	dsn := os.Getenv("SUPER_CHAT_SENTRY_DSN")
	if dsn == "" {
		dbglog.Fatal("SENTRY_DSN is not exist.")
	}
	sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
}
