package sh

import (
	"encoding/base64"
	"encoding/json"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/shell"
)

var (
	ShellStartAndAttach = shell.StartAndAttach
	ShellAttach         = shell.Attach
	ShellDetach         = shell.Detach
	ShellInput          = shell.Input
	ShellResize         = shell.Resize
	ShellKill           = shell.Kill
	ShellList           = shell.List
)

func Start(name string, data []byte, reply func([]byte)) {
	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		Cmd   string   `json:"cmd"`
		Args  []string `json:"args"`
		Env   []string `json:"env"`
		WD    string   `json:"wd"`
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

	if j.Args == nil {
		j.Args = []string{}
	}

	if j.Cols <= 6 {
		j.Cols = 6
	}

	if j.Rows <= 2 {
		j.Rows = 2
	}

	subscriptionID, shellSessionID, err := ShellStartAndAttach(j.Cmd, j.Args, j.Env, j.WD, j.RunAs, j.Cols, j.Rows, reply)
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

	subscriptionID := ShellAttach(j.ShellSessionID, j.Cols, j.Rows, reply)

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

	ShellDetach(j.ShellSessionID, j.SubscriptionID)

	ReplyJSON(H{
		"detached": true,
	}, reply)
}

func Input(name string, data []byte, reply func([]byte)) {
	// this handler does not reply
	if len(data) == 0 {
		return
	}

	// [shellSessionID, base64-encoded-data]

	var j []string
	if err := json.Unmarshal(data, &j); err != nil {
		return
	}

	if len(j) < 2 || j[0] == "" || j[1] == "" {
		return
	}

	shellSessionID := j[0]
	encodedData := j[1]

	data, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return
	}

	ShellInput(shellSessionID, data)
}

func Resize(name string, data []byte, reply func([]byte)) {
	// this handler does not reply
	if len(data) == 0 {
		return
	}

	// [shellSessionID, cols, rows]
	var j []interface{}

	if err := json.Unmarshal(data, &j); err != nil {
		return
	}

	if len(j) < 3 {
		return
	}

	shellSessionID, ok := j[0].(string)
	if !ok {
		return
	}

	cols, ok := j[1].(float64)
	if !ok {
		return
	}

	rows, ok := j[2].(float64)
	if !ok {
		return
	}

	if cols <= 6 {
		cols = 6
	}

	if rows <= 2 {
		rows = 2
	}

	ShellResize(shellSessionID, uint(cols), uint(rows))
}

func Kill(name string, data []byte, reply func([]byte)) {
	// this handler does not reply
	if len(data) == 0 {
		return
	}

	var j struct {
		ShellSessionID string `json:"shell_session_id"`
	}

	if err := json.Unmarshal(data, &j); err != nil {
		return
	}

	if j.ShellSessionID == "" {
		return
	}

	ShellKill(j.ShellSessionID)
}

func List(name string, data []byte, reply func([]byte)) {
	ShellList(reply)
}
