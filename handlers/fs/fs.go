package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/nitrous-io/nitrous-solo/handlers"
)

func List(name string, data []byte, reply func([]byte)) {
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

		if strings.Contains(err.Error(), "too many levels of symbolic links") {
			ReplyJSON(H{"error": "symlink_too_deep"}, reply)
			return
		}

		ReplyUnexpectedError(err, reply)
		return
	}

	if !fi.IsDir() {
		ReplyJSON(H{"error": "path_not_directory"}, reply)
		return
	}

	list, err := ioutil.ReadDir(j.Path)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	files := make([]string, len(list))
	for i, fi := range list {
		name := fi.Name()
		if fi.Mode()&os.ModeSymlink > 0 {
			fii, err := os.Stat(filepath.Join(j.Path, name))
			// follow symlink if possible
			if err == nil {
				fi = fii
			}
		}

		if fi.IsDir() {
			name += "/"
		}
		files[i] = name
	}

	ReplyJSON(H{"files": files}, reply)
	return
}
