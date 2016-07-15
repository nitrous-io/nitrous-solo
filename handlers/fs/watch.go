package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	"github.com/satori/go.uuid"
)

var (
	// watches is a map of "watch UUID" -> "watch struct", so that we can stop
	// a watcher in a different WS handler.
	watches     = make(map[string]*watch)
	watchesLock sync.Mutex // map operations aren't thread safe
)

type watch struct {
	id       string
	stop     chan struct{}
	stopOnce *sync.Once
}

func newWatch() *watch {
	w := &watch{
		id:       uuid.NewV4().String(),
		stop:     make(chan struct{}, 1),
		stopOnce: &sync.Once{},
	}

	watchesLock.Lock()
	defer watchesLock.Unlock()
	watches[w.id] = w
	return w
}

func Watch(c *ws.Context) {
	noDoneAfterReturn := false
	defer func() {
		if !noDoneAfterReturn {
			c.Done()
		}
	}()
	data := c.Data
	reply := c.Reply

	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	fi, err := os.Stat(j.Path)
	if err != nil {
		if os.IsNotExist(err) {
			ReplyJSON(H{"error": "invalid_path"}, reply)
			return
		}
	}

	// Only allow directories to be watched. Technically, we can watch regular
	// files too, but we don't have a use case for it right now.
	if !fi.IsDir() {
		ReplyJSON(H{"error": "path_not_directory"}, reply)
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Warnf("failed to start fsnotify watcher, err: %v", err)
		ReplyJSON(H{"error": "temporarily_unavailable"}, reply)
		return
	}

	w := newWatch()

	// Return a UUID so that clients can stop watching cleanly.
	ReplyJSON(H{"evt": "started", "id": w.id}, reply)

	watchPath, err := filepath.EvalSymlinks(j.Path)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	if err := watcher.Add(watchPath); err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	noDoneAfterReturn = true

	stop := func() {
		ReplyJSON(H{"evt": "stopped", "id": w.id}, reply)
		c.Done()
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create > 0 {
					ReplyJSON(H{"evt": "created", "path": event.Name, "id": w.id}, reply)
				}

				if event.Op&fsnotify.Write > 0 {
					ReplyJSON(H{"evt": "modified", "path": event.Name, "id": w.id}, reply)
				}

				if event.Op&fsnotify.Remove > 0 {
					ReplyJSON(H{"evt": "removed", "path": event.Name, "id": w.id}, reply)
				}

				if event.Op&fsnotify.Rename > 0 {
					// Renaming a file also triggers a separate fsnotify.Create event with
					// the new filename (tested on Linux).
					ReplyJSON(H{"evt": "renamed", "path": event.Name, "id": w.id}, reply)
				}

			case err := <-watcher.Errors:
				// watcher shouldn't stop when it publishes error
				logrus.Warnf("fs.Watch: fsnotify watcher error on %q, err: %v", j.Path, err)
				ReplyJSON(H{"error": "file_watcher_error"}, reply)

			case <-c.OnClose:
				stop()
				return

			case <-w.stop:
				stop()
				return
			}
		}
	}()
}

func Unwatch(c *ws.Context) {
	defer c.Done()
	data := c.Data
	reply := c.Reply

	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	w, ok := watches[j.ID]
	if !ok {
		return
	}

	watchesLock.Lock()
	defer watchesLock.Unlock()
	delete(watches, j.ID)

	w.stopOnce.Do(func() {
		w.stop <- struct{}{}
	})
}
