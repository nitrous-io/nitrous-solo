package fs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"path/filepath"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
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
		Path      string `json:"path"`
		Content   string `json:"content"`
		Overwrite bool   `json:"overwrite"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	if !filepath.IsAbs(j.Path) {
		ReplyJSON(H{"error": "path_is_not_absolute"}, reply)
		return
	}

	content, err := base64.StdEncoding.DecodeString(j.Content)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	buf := bytes.NewBuffer(content)
	if err := atomicwrite.AtomicWrite(buf, j.Path, runAs, 0644, j.Overwrite); err != nil {
		switch err {
		case atomicwrite.ErrSymlinkTooDeep, atomicwrite.ErrPathIsDirectory,
			atomicwrite.ErrPermissionDenied, atomicwrite.ErrFileExist, atomicwrite.ErrInvalidUser:
			ReplyJSON(H{"error": err.Error()}, reply)
			return
		}

		ReplyUnexpectedError(err, reply)
		return
	}

	ReplyJSON(H{"written": true}, reply)
}
