package sh

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

var executor Executor = &StdExecutor{}

type Request struct {
	Cmd   string   `json:"cmd"`
	Env   []string `json:"env"`
	Args  []string `json:"args"`
	RunAs string   `json:"run_as"`
	WD    string   `json:"wd"`
}

func Run(c *ws.Context) {
	data := c.Data
	reply := c.Reply

	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	if req.RunAs == "" {
		req.RunAs, _ = c.Meta["user"].(string)
	}

	if req.Cmd == "" || req.RunAs == "" {
		ReplyJSON(H{"error": "invalid_params"}, reply)
		return
	}

	cmd := executor.Command(req.Cmd, req.Args...)
	cmd.Dir = req.WD
	cmd.Env = append(cmd.Env, req.Env...)

	usr, err := user.Lookup(req.RunAs)
	if err != nil {
		ReplyJSON(H{"error": "unknown user"}, reply)
		return
	}

	uid, err := strconv.ParseInt(usr.Uid, 10, 32)
	if err != nil {
		ReplyJSON(H{"error": "unknown user"}, reply)
		return
	}

	gid, err := strconv.ParseInt(usr.Gid, 10, 32)
	if err != nil {
		ReplyJSON(H{"error": "unknown user"}, reply)
		return
	}

	// Run command as the given user. This will not work unless the user running
	// this process has setuid/setgid privileges.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)},
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()

	res := H{
		"stdout":    base64.StdEncoding.EncodeToString(output),
		"stderr":    base64.StdEncoding.EncodeToString(stderr.Bytes()),
		"exit_code": 0,
	}
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				res["exit_code"] = status.ExitStatus()

				ReplyJSON(res, reply)
				return
			}
		}

		ReplyJSON(H{"error": fmt.Sprintf("error running command: %v", err)}, reply)
		return
	}

	ReplyJSON(res, reply)
}
