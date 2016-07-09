package fake

import "github.com/satori/go.uuid"

var (
	ShellStartAndAttachCalls Calls
	ShellStartAndAttachError error
	ShellAttachCalls         Calls
	ShellDetachCalls         Calls
	ShellInputCalls          Calls
	ShellResizeCalls         Calls
	ShellKillCalls           Calls
	ShellListCalls           Calls
)

func FakeShellReset() {
	ShellStartAndAttachCalls = Calls{}
	ShellStartAndAttachError = nil
	ShellAttachCalls = Calls{}
	ShellDetachCalls = Calls{}
	ShellInputCalls = Calls{}
	ShellResizeCalls = Calls{}
	ShellKillCalls = Calls{}
	ShellListCalls = Calls{}
}

func ShellStartAndAttach(cmd string, args, env []string, wd, runAs string, cols, rows uint, reply func([]byte)) (subscriptionID string, shellSessionID string, err error) {
	err = ShellStartAndAttachError
	if err == nil {
		subscriptionID = uuid.NewV4().String()
		shellSessionID = uuid.NewV4().String()
	}

	ShellStartAndAttachCalls.Add(
		List{cmd, args, env, wd, runAs, cols, rows, reply},
		List{subscriptionID, shellSessionID, err},
		nil,
	)
	return
}

func ShellAttach(shellSessionID string, cols, rows uint, reply func([]byte)) (subscriptionID string) {
	subscriptionID = uuid.NewV4().String()

	ShellAttachCalls.Add(
		List{shellSessionID, cols, rows, reply},
		List{subscriptionID},
		nil,
	)
	return
}

func ShellDetach(shellSessionID, subscriptionID string) {
	ShellDetachCalls.Add(
		List{shellSessionID, subscriptionID},
		List{},
		nil,
	)
}

func ShellInput(shellSessionID string, data []byte) {
	ShellInputCalls.Add(
		List{shellSessionID, data},
		List{},
		nil,
	)
}

func ShellResize(shellSessionID string, cols, rows uint) {
	ShellResizeCalls.Add(
		List{shellSessionID, cols, rows},
		List{},
		nil,
	)
}

func ShellKill(shellSessionID string) {
	ShellKillCalls.Add(
		List{shellSessionID},
		List{},
		nil,
	)
}

func ShellList(reply func([]byte)) {
	ShellListCalls.Add(
		List{reply},
		List{},
		nil,
	)
}
