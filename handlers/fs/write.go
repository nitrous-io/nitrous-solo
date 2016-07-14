package fs

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"strconv"
	"strings"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

func Write(c *ws.Context) {
	data := c.Data
	reply := c.Reply

	runAs, ok := c.Meta["user"].(string)
	if !ok {
		ReplyUnexpectedError(errors.New("meta data is not intialized"), reply)
		return
	}

	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	fi, err := os.Stat(j.Path)
	if err != nil {
		if strings.Contains(err.Error(), "too many levels of symbolic links") {
			ReplyJSON(H{"error": "symlink_too_deep"}, reply)
			return
		}

		if !os.IsNotExist(err) {
			ReplyUnexpectedError(err, reply)
			return
		}
	} else {
		if fi.IsDir() {
			ReplyJSON(H{"error": "file_unwritable"}, reply)
			return
		}
	}

	f, err := os.OpenFile(j.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		if os.IsPermission(err) {
			ReplyJSON(H{"error": "permission_denied"}, reply)
			return
		}

		ReplyUnexpectedError(err, reply)
		return
	}
	defer f.Close()

	if _, err = f.Write([]byte(j.Content)); err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	u, err := user.Lookup(runAs)
	if err != nil {
		ReplyJSON(H{"error": "user_not_found"}, reply)
		return
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	if err := f.Chown(uid, gid); err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	ReplyJSON(H{"status": "ok"}, reply)
}
