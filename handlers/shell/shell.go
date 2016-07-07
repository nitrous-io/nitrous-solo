package shell

import (
	"encoding/base64"
	"encoding/json"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/shell"
)

func Start(name string, data []byte, reply func([]byte)) {
	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		Cmd   string   `json:"cmd"`
		Args  []string `json:"args"`
		RunAs string   `json:"run_as"`
		Cols  uint     `json:"cols"`
		Rows  uint     `json:"rows"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	if j.Cmd == "" || j.RunAs == "" {
		ReplyJSON(H{"error": "invalid_params"}, reply)
		return
	}

	if j.Cols <= 6 {
		j.Cols = 6
	}

	if j.Rows <= 2 {
		j.Rows = 2
	}

	subscriptionID, shellSessionID, err := shell.StartShellSessionAndAttach(j.Cmd, j.Args, j.RunAs, j.Cols, j.Rows, reply)
	if err != nil {
		ReplyJSON(H{
			"error":             "failed_to_start_shell_session",
			"error_description": err.Error(),
		}, reply)
		return
	}

	ReplyJSON(H{
		"subscription_id":  subscriptionID,
		"shell_session_id": shellSessionID,
	}, reply)
}

func Attach(name string, data []byte, reply func([]byte)) {
	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		ShellSessionID string `json:"shell_session_id"`
		Cols           uint   `json:"cols"`
		Rows           uint   `json:"rows"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	if j.ShellSessionID == "" {
		ReplyJSON(H{"error": "invalid_params"}, reply)
		return
	}

	if j.Cols <= 6 {
		j.Cols = 6
	}

	if j.Rows <= 2 {
		j.Rows = 2
	}

	subscriptionID := shell.Attach(j.ShellSessionID, j.Cols, j.Rows, reply)

	ReplyJSON(H{
		"subscription_id": subscriptionID,
	}, reply)
}

func Detach(name string, data []byte, reply func([]byte)) {
	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		ShellSessionID string `json:"shell_session_id"`
		SubscriptionID string `json:"subscription_id"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	if j.ShellSessionID == "" || j.SubscriptionID == "" {
		ReplyJSON(H{"error": "invalid_params"}, reply)
		return
	}

	shell.Detach(j.ShellSessionID, j.SubscriptionID)

	ReplyJSON(H{
		"detached": true,
	}, reply)
}

func Input(name string, data []byte, _ func([]byte)) {
	// this handler does not reply
	if len(data) == 0 {
		return
	}

	var j struct {
		ShellSessionID string `json:"s"`
		Data           string `json:"d"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		return
	}

	if j.ShellSessionID == "" || j.Data == "" {
		return
	}

	data, err := base64.StdEncoding.DecodeString(j.Data)
	if err != nil {
		return
	}

	shell.Input(j.ShellSessionID, data)
}
