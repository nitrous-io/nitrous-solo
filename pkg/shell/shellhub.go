package shell

import (
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/nitrous-io/pty"
	"github.com/satori/go.uuid"
)

var hub *shellHubType

func init() {
	hub = newShellHub()
	go hub.run()
}

type subscription struct {
	id             string
	shellSessionID string
	reply          func([]byte)
}

func newSubscription(shellSessionID string, reply func([]byte)) subscription {
	return subscription{
		id:             uuid.NewV4().String(),
		shellSessionID: shellSessionID,
		reply:          reply,
	}
}

type resizeParams struct {
	shellSessionID string
	cols           uint
	rows           uint
}

type shellInput struct {
	shellSessionID string
	data           []byte
}

type shellOutput struct {
	shellSessionID string
	kind           string
	data           []byte
	then           func()
}

type replyFunc func([]byte)

type shellHubType struct {
	sessions      map[string]*shellSession           // <shellSessionID, *shellSession>
	subscriptions map[string]map[string]subscription // <shellSessionID, <subscriptionID, subscription>
	lock          sync.Mutex

	input   chan shellInput
	publish chan shellOutput
	resize  chan resizeParams
	kill    chan string // shellSessionID
	list    chan replyFunc

	stop chan struct{}
}

func newShellHub() *shellHubType {
	return &shellHubType{
		sessions:      map[string]*shellSession{},
		subscriptions: map[string]map[string]subscription{},

		input:   make(chan shellInput, 256),
		publish: make(chan shellOutput, 256),
		resize:  make(chan resizeParams, 256),
		kill:    make(chan string, 256),
		list:    make(chan replyFunc, 256),

		stop: make(chan struct{}),
	}
}

func (h *shellHubType) run() {
	for {
		select {
		case s := <-h.input:
			if sess, ok := h.sessions[s.shellSessionID]; ok {
				sess.pty.Write(s.data)
			}

		case p := <-h.resize:
			if sess, ok := h.sessions[p.shellSessionID]; ok {
				if sess.pty != nil {
					pty.Setsize(sess.pty, uint16(p.rows), uint16(p.cols))
				}
			}

		case s := <-h.publish:
			if subs, ok := h.subscriptions[s.shellSessionID]; ok {
				for _, sub := range subs {
					var d string
					if s.kind == "d" {
						d = base64.StdEncoding.EncodeToString(s.data)
					} else {
						d = string(s.data)
					}

					if j, err := json.Marshal([]string{
						s.kind,
						d,
					}); err == nil {
						sub.reply(j)
					}
				}
				if s.then != nil {
					s.then()
				}
			}

		case shellSessionID := <-h.kill:
			if sess, ok := h.sessions[shellSessionID]; ok {
				if sess.cmd != nil && sess.cmd.Process != nil {
					sess.cmd.Process.Kill()
				}
			}

		case reply := <-h.list:
			sessIDs := make([]string, len(h.sessions))
			i := 0
			for sessID := range h.sessions {
				sessIDs[i] = sessID
				i++
			}
			if b, err := json.Marshal(sessIDs); err == nil {
				reply(b)
			}

		case <-h.stop:
			close(h.input)
			close(h.publish)
			close(h.resize)
			close(h.kill)
			close(h.list)
			close(h.stop)

			for _, sess := range h.sessions {
				if sess.cmd != nil && sess.cmd.Process != nil {
					sess.cmd.Process.Kill()
				}
			}
			return // exit function
		}
	}
}

func (h *shellHubType) register(sess *shellSession) {
	h.lock.Lock()
	defer h.lock.Unlock()
	hub.sessions[sess.id] = sess
}

func (h *shellHubType) subscribe(s subscription, cols, rows uint) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.sessions[s.shellSessionID]; !ok {
		return
	}

	if subs, ok := h.subscriptions[s.shellSessionID]; ok {
		subs[s.id] = s
	} else {
		h.subscriptions[s.shellSessionID] = map[string]subscription{
			s.id: s,
		}
	}

	if sess, ok := h.sessions[s.shellSessionID]; ok {
		if sess.pty != nil {
			pty.Setsize(sess.pty, uint16(rows), uint16(cols))
		}
	}
}

func (h *shellHubType) unsubscribe(shellSessionID, subscriptionID string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if subs, ok := h.subscriptions[shellSessionID]; ok {
		delete(subs, subscriptionID)
	}
}

func (h *shellHubType) unregister(shellSessionID string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.sessions, shellSessionID)
	delete(h.subscriptions, shellSessionID)
}
