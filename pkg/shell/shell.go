package shell

import "github.com/satori/go.uuid"

type subscription struct {
	id             string
	shellSessionID string
	reply          func([]byte)
}

type attachParams struct {
	subscription subscription
	cols         uint
	rows         uint
}

type detachParams struct {
	subscriptionID string
	shellSessionID string
}

type resizeParams struct {
	shellSessionID string
	cols           uint
	rows           uint
}

func newSubscription(shellSessionID string, reply func([]byte)) subscription {
	return subscription{
		id:             uuid.NewV4().String(),
		shellSessionID: shellSessionID,
		reply:          reply,
	}
}

type shellInput struct {
	shellSessionID string
	data           []byte
}

type shellOutput struct {
	shellSessionID string
	kind           string
	data           []byte
}

func StartShellSessionAndAttach(cmd string, args []string, runAs string, cols, rows uint, reply func([]byte)) (subscriptionID string, shellSessionID string, err error) {
	sess, err := startShellSession(cmd, args, runAs)
	if err != nil {
		return "", "", err
	}

	hub.new <- sess

	shellSessionID = sess.id
	subscriptionID = Attach(shellSessionID, cols, rows, reply)
	return subscriptionID, shellSessionID, nil
}

func Attach(shellSessionID string, cols, rows uint, reply func([]byte)) (subscriptionID string) {
	sub := newSubscription(shellSessionID, reply)
	hub.attach <- attachParams{
		subscription: sub,
		cols:         cols,
		rows:         rows,
	}
	return sub.id
}

func Detach(shellSessionID, subscriptionID string) {
	hub.detach <- detachParams{
		subscriptionID: subscriptionID,
		shellSessionID: shellSessionID,
	}
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
