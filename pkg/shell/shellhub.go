package shell

import (
	"encoding/base64"
	"encoding/json"

	"github.com/nitrous-io/pty"
)

var (
	hub         *shellHubType
	chanBufSize = 256
)

func init() {
	hub = newShellHub()
	go hub.run()
}

type replyFunc func([]byte)

type shellHubType struct {
	sessions      map[string]*shellSession           // <shellSessionID, *shellSession>
	subscriptions map[string]map[string]subscription // <shellSessionID, <subscriptionID, subscription>

	new     chan *shellSession
	attach  chan attachParams
	detach  chan detachParams
	input   chan shellInput
	publish chan shellOutput
	resize  chan resizeParams
	kill    chan string // shellSessionID
	finish  chan string // shellSessionID
	list    chan replyFunc

	stop chan struct{}
}

func newShellHub() *shellHubType {
	return &shellHubType{
		sessions:      map[string]*shellSession{},
		subscriptions: map[string]map[string]subscription{},

		new:     make(chan *shellSession, chanBufSize),
		attach:  make(chan attachParams, chanBufSize),
		detach:  make(chan detachParams, chanBufSize),
		input:   make(chan shellInput, chanBufSize),
		publish: make(chan shellOutput, chanBufSize),
		resize:  make(chan resizeParams, chanBufSize),
		kill:    make(chan string, chanBufSize),
		finish:  make(chan string, chanBufSize),
		list:    make(chan replyFunc, chanBufSize),

		stop: make(chan struct{}),
	}
}

func (h *shellHubType) run() {
	for {
		select {
		case sess := <-h.new:
			hub.sessions[sess.id] = sess

		case p := <-h.attach:
			s := p.subscription

			if _, ok := h.sessions[s.shellSessionID]; !ok {
				break
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
					pty.Setsize(sess.pty, uint16(p.rows), uint16(p.cols))
				}
			}

		case p := <-h.detach:
			if subs, ok := h.subscriptions[p.shellSessionID]; ok {
				delete(subs, p.subscriptionID)
			}

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
			}

		case shellSessionID := <-h.kill:
			if sess, ok := h.sessions[shellSessionID]; ok {
				if sess.cmd != nil && sess.cmd.Process != nil {
					sess.cmd.Process.Kill()
				}
			}

		case shellSessionID := <-h.finish:
			delete(h.sessions, shellSessionID)
			delete(h.subscriptions, shellSessionID)

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
			close(h.new)
			close(h.attach)
			close(h.detach)
			close(h.input)
			close(h.publish)
			close(h.resize)
			close(h.kill)
			close(h.finish)
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
