package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/satori/go.uuid"
)

var (
	// watches is a map of "watch UUID" -> "watch struct", so that we can stop
	// a watcher in a different WS handler.
	// We could make some optimizations in future, like re-using a watcher if the
	// same path is requested more than once.
	watches = make(map[string]*watch)
)

type watch struct {
	id      string
	stop    chan struct{}
	stopped chan struct{}
}

func Watch(name string, data []byte, reply func([]byte)) {
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

	// Add all directories in the tree to watcher (fsnotify does not recursively
	// watch subdirectories).
	subdirs, err := dirtree(j.Path)
	if err != nil {
		logrus.Errorf("failed to determine subdirectories of %q, err: %v", j.Path, err)
		ReplyJSON(H{"error": ""}, reply)
		return
	}
	for _, d := range subdirs {
		watcher.Add(d)
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					ReplyJSON(H{"evt": "created", "path": event.Name, "id": w.id}, reply)

					fi, err := os.Stat(event.Name)
					if err != nil {
						logrus.Errorf("failed to stat created file %q, err: %v", event.Name, err)
						// Disregard error - maybe we don't have permissions to stat it.
						continue
					}

					if !fi.IsDir() {
						continue
					}

					// If file created is a dir, we start watching it _and_ its
					// subdirectories for changes too.
					// This is necessary because e.g. `mkdir -p /path/to/subdir/deep/dir`.
					dir := event.Name
					logrus.Debugf("created dir, watching it too: %q", dir)

					subdirs, err := dirtree(dir)
					if err != nil {
						logrus.Errorf("failed to determine subdirectories of %q, err: %v", dir, err)
						// Disregard error.
						continue
					}

					for _, d := range subdirs {
						watcher.Add(d)
					}
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					ReplyJSON(H{"evt": "modified", "path": event.Name, "id": w.id}, reply)
				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					ReplyJSON(H{"evt": "removed", "path": event.Name, "id": w.id}, reply)
				}

				if event.Op&fsnotify.Rename == fsnotify.Rename {
					// Renaming a file also triggers a separate fsnotify.Create event with
					// the new filename (tested on Linux).
					ReplyJSON(H{"evt": "renamed", "path": event.Name, "id": w.id}, reply)
				}
			case err := <-watcher.Errors:
				logrus.Errorf("fs.Watch: fsnotify watcher error on %q, err: %v", j.Path, err)
				ReplyJSON(H{"error": "file_watcher_error"}, reply)
				ReplyJSON(H{"evt": "stopped", "id": w.id}, reply)
				return
			case <-w.stop:
				ReplyJSON(H{"evt": "stopped", "id": w.id}, reply)
				w.stopped <- struct{}{}
				return
			}
		}
	}()
}

func Unwatch(name string, data []byte, reply func([]byte)) {
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

	stopWatch(j.ID)

	ReplyJSON(H{"evt": "unwatched", "id": j.ID}, reply)
}

func newWatch() *watch {
	w := &watch{
		id:      uuid.NewV4().String(),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}

	watches[w.id] = w

	return w
}

func stopWatch(id string) {
	w, ok := watches[id]
	if !ok {
		return
	}

	var once sync.Once
	once.Do(func() {
		w.stop <- struct{}{}
		<-w.stopped
		delete(watches, id)
	})
}

func dirtree(dir string) ([]string, error) {
	// If given a symlink to a dir, resolve it to its real path because
	// filepath.Walk does not handle symlinks.
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}

	var subdirs []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Warnf("error walking %q, err: %v", path, err)
			return nil // Don't return an error so that we continue walking elsewhere.
		}

		mode := info.Mode()
		if !mode.IsDir() {
			return nil
		}

		subdirs = append(subdirs, path)

		return nil
	}

	if err := filepath.Walk(realDir, walkFn); err != nil {
		return nil, err
	}

	return subdirs, nil
}
