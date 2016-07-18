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
	"time"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

var (
	executor Executor = &StdExecutor{}

	DefaultTimeout = 30 * time.Second
)

type Request struct {
	Cmd     string   `json:"cmd"`
	Env     []string `json:"env"`
	Args    []string `json:"args"`
	RunAs   string   `json:"run_as"`
	WD      string   `json:"wd"`
	Timeout uint     `json:timeout`
}

func Run(c *ws.Context) {
	defer c.Done()
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
		var ok bool
		req.RunAs, ok = c.Meta["user"].(string)
		if !ok {
			ReplyUnexpectedError(nil, reply)
			return
		}
	}

	if req.Cmd == "" || req.RunAs == "" {
		ReplyJSON(H{"error": "invalid_params"}, reply)
		return
	}

	cmd := executor.Command(req.Cmd, req.Args...)
	cmd.Dir = req.WD
	cmd.Env = req.Env

	usr, err := user.Lookup(req.RunAs)
	if err != nil {
		ReplyJSON(H{"error": "invalid_user"}, reply)
		return
	}

	uid, err := strconv.ParseUint(usr.Uid, 10, 32)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	gid, err := strconv.ParseUint(usr.Gid, 10, 32)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	// Run command as the given user. This will not work unless the user running
	// this process has setuid/setgid privileges.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)},
	}

	var (
		stderr  bytes.Buffer
		output  []byte
		timeout time.Duration
	)

	cmd.Stderr = &stderr
	done := make(chan struct{})
	if req.Timeout != 0 {
		timeout = time.Duration(req.Timeout) * time.Millisecond
	} else {
		timeout = DefaultTimeout
	}

	go func() {
		output, err = cmd.Output()
		done <- struct{}{}
	}()

	select {
	case <-done:
		break
	case <-time.After(timeout):
		cmd.Process.Kill()
		ReplyJSON(H{"error": "command_timeout"}, reply)
		return
	}

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
