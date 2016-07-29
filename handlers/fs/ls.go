package fs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nitrous-io/nitrous-solo/data"
	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

func List(c *ws.Context) {
	defer c.Done()
	d := c.Data
	reply := c.Reply

	username, ok := c.Meta["user"].(string)
	if !ok {
		ReplyUnexpectedError(errors.New("meta data is not intialized"), reply)
		return
	}

	ud := data.D.UserData[username]
	if ud == nil {
		ReplyUnexpectedError(errors.New("invalid_found"), reply)
		return
	}

	if len(d) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(d, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	homePath, err := data.HomePath(ud)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	relPath, err := filepath.Rel(homePath, j.Path)
	if err != nil || strings.HasPrefix(relPath, "..") {
		ReplyJSON(H{"error": "invalid_path"}, reply)
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
}
