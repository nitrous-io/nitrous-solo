package shell

import (
	"encoding/base64"
	"encoding/json"

	"github.com/nitrous-io/pty"
)

var hub *shellHubType

func init() {
	hub = newShellHub()
	go hub.run()
}

type shellHubType struct {
	sessions      map[string]*shellSession           // <shellSessionID, *shellSession>
	subscriptions map[string]map[string]subscription // <shellSessionID, <subscriptionID, subscription>

	new     chan *shellSession
	attach  chan attachParams
	detach  chan detachParams
	input   chan shellInput
	publish chan shellOutput
	resize  chan resizeParams
	finish  chan string // shellSessionID

	stop chan struct{}
}

func newShellHub() *shellHubType {
	return &shellHubType{
		sessions:      map[string]*shellSession{},
		subscriptions: map[string]map[string]subscription{},

		new:     make(chan *shellSession, 256),
		attach:  make(chan attachParams, 256),
		detach:  make(chan detachParams, 256),
		input:   make(chan shellInput, 256),
		publish: make(chan shellOutput, 256),
		resize:  make(chan resizeParams, 256),
		finish:  make(chan string, 256),

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

		case shellSessionID := <-h.finish:
			delete(h.sessions, shellSessionID)
			delete(h.subscriptions, shellSessionID)

		case <-h.stop:
			close(h.new)
			close(h.attach)
			close(h.detach)
			close(h.input)
			close(h.publish)
			close(h.resize)
			close(h.finish)
			close(h.stop)

			for _, sess := range h.sessions {
				sess.cmd.Process.Kill()
			}
			return // exit function
		}
	}
}
