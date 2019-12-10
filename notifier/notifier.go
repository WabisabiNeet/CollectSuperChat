package notifier

import "sync"

// Notifier is interface
type Notifier interface {
	PollingStart(*sync.WaitGroup)
}
