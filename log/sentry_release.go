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

	commit := os.Getenv("GIT_COMMIT")
	if commit == "" {
		commit = "debug"
	}

	servername := os.Getenv("HOSTNAME")

	sentry.Init(sentry.ClientOptions{
		Dsn:        dsn,
		Release:    commit,
		ServerName: servername,
	})
}
