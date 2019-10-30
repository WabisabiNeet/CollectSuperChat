package watch

import (
	"sync"
)

// Watcher struct.
type Watcher struct {
	wg *sync.WaitGroup
}

// NewWatcher Create Watcher struct.
func NewWatcher(wg *sync.WaitGroup) Watcher {
	return Watcher{
		wg: wg,
	}
}
