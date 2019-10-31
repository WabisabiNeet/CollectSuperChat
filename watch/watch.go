package watch

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher struct.
type Watcher struct {
	wg     *sync.WaitGroup
	target string
}

// NewWatcher Create Watcher struct.
// w is sync Object.
// target is watching target.
func NewWatcher(w *sync.WaitGroup, targetFolder string) Watcher {
	return Watcher{
		wg:     w,
		target: targetFolder,
	}
}

// Start watch
func (w *Watcher) Start() error {
	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	os.MkdirAll(w.target, os.ModeDir|0755)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = watcher.Add(w.target)
	if err != nil {
		return err
	}

	go func() {
		w.wg.Add(1)
		defer w.wg.Done()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("create file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-quit:
				return
			}
		}
	}()

	return nil
}
