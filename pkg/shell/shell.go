package shell

import "github.com/nitrous-io/nitrous-solo/pkg/escapecsi"

func StartAndAttach(cmd string, args, env []string, wd, runAs string, cols, rows uint, reply func([]byte)) (subscriptionID string, shellSessionID string, err error) {
	sess, err := startShellSession(cmd, args, env, wd, runAs)
	if err != nil {
		return "", "", err
	}

	hub.register(sess)

	shellSessionID = sess.id
	subscriptionID, _ = Attach(shellSessionID, cols, rows, reply)
	return subscriptionID, shellSessionID, nil
}

func Attach(shellSessionID string, cols, rows uint, reply func([]byte)) (subscriptionID string, scrollback []byte) {
	sub := newSubscription(shellSessionID, reply)
	scrollback = hub.subscribe(sub, cols, rows)

	return sub.id, escapecsi.EscapeCSI(scrollback)
}

func Detach(shellSessionID, subscriptionID string) {
	hub.unsubscribe(shellSessionID, subscriptionID)
}

func Input(shellSessionID string, data []byte) {
	hub.input <- shellInput{
		shellSessionID: shellSessionID,
		data:           data,
	}
}

func Resize(shellSessionID string, cols, rows uint) {
	hub.resize <- resizeParams{
		shellSessionID: shellSessionID,
		cols:           cols,
		rows:           rows,
	}
}

func Kill(shellSessionID string) {
	hub.kill <- shellSessionID
}

func List(reply func([]byte)) {
	hub.list <- reply
}
