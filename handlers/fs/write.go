package fs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

func Write(c *ws.Context) {
	defer c.Done()
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

	if !filepath.IsAbs(j.Path) {
		ReplyJSON(H{"error": "path_is_not_absolute"}, reply)
		return
	}

	fi, err := os.Lstat(j.Path)
	if err != nil && !os.IsNotExist(err) {
		ReplyUnexpectedError(err, reply)
		return
	}

	if fi != nil {
		if fi.Mode()&os.ModeSymlink > 0 {
			// file is symlink, try to follow
			fi, err = os.Stat(j.Path)
			if err != nil && !os.IsNotExist(err) {
				if strings.Contains(strings.ToLower(err.Error()), "too many levels of symbolic links") {
					ReplyJSON(H{"error": "symlink_too_deep"}, reply)
					return
				}

				ReplyUnexpectedError(err, reply)
				return
			}

			// was possible to follow symlink, evaluate symlink and get target path
			p, err := filepath.EvalSymlinks(j.Path)
			if err != nil {
				ReplyUnexpectedError(err, reply)
				return
			}
			j.Path = p
		}

		if fi.IsDir() {
			ReplyJSON(H{"error": "path_is_directory"}, reply)
			return
		}
	}

	if fi != nil {
		// if file exists, make sure it's writable
		ff, err := os.OpenFile(j.Path, os.O_RDWR, fi.Mode().Perm())
		if err != nil {
			if os.IsPermission(err) {
				ReplyJSON(H{"error": "permission_denied"}, reply)
				return
			}
			ReplyUnexpectedError(err, reply)
			return
		}
		ff.Close()
	}

	content, err := base64.StdEncoding.DecodeString(j.Content)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	f, err := ioutil.TempFile("", "nitrous")
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}
	defer f.Close()

	if _, err = f.Write(content); err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	if fi != nil {
		// if file already exists at path
		// copy permission
		if err := f.Chmod(fi.Mode().Perm()); err != nil {
			ReplyUnexpectedError(err, reply)
			return
		}

		// copy owner
		s, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			ReplyUnexpectedError(err, reply)
			return
		}

		if err := f.Chown(int(s.Uid), int(s.Gid)); err != nil {
			ReplyUnexpectedError(err, reply)
			return
		}
	} else {
		// it is a new file
		u, err := user.Lookup(runAs)
		if err != nil {
			ReplyJSON(H{"error": "invalid_user"}, reply)
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

		if err := f.Chmod(0644); err != nil {
			ReplyUnexpectedError(err, reply)
			return
		}
	}

	f.Close()

	// move tmp file over to the intended file path
	if err := os.Rename(f.Name(), j.Path); err != nil {
		os.Remove(f.Name())
		ReplyUnexpectedError(err, reply)
		return
	}

	ReplyJSON(H{"written": true}, reply)
}
