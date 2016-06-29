package ws_test

import (
	"runtime"

	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("hub", func() {
	BeforeEach(func() {
		ws.ResetHub_()
	})

	Describe("RegisterConn", func() {
		It("adds a connection to connections list", func() {
			conn := &ws.Conn{}

			Expect(ws.HubConnections_()).To(HaveLen(0))
			ws.RegisterConn_(conn)
			runtime.Gosched()

			Expect(ws.HubConnections_()).To(HaveLen(1))
			Expect(ws.HubConnections_()[conn]).NotTo(BeNil())
		})
	})

	Describe("UnregisterConn", func() {
		It("removes a connection from connections list", func() {
			conn := &ws.Conn{Send: make(chan []byte, 10)}

			ws.RegisterConn_(conn)
			runtime.Gosched()
			Expect(ws.HubConnections_()).To(HaveLen(1))

			ws.UnregisterConn_(conn)
			runtime.Gosched()

			Expect(ws.HubConnections_()).To(HaveLen(0))
			Expect(conn.Send).To(BeClosed())
		})
	})

	Describe("Broadcast", func() {
		It("sends message to every connection", func() {
			conn1 := &ws.Conn{Send: make(chan []byte, 10)}
			conn2 := &ws.Conn{Send: make(chan []byte, 10)}

			ws.RegisterConn_(conn1)
			ws.RegisterConn_(conn2)
			runtime.Gosched()

			Expect(conn1.Send).To(HaveLen(0))
			Expect(conn2.Send).To(HaveLen(0))

			ws.Broadcast([]byte("foo"))
			runtime.Gosched()

			Expect(conn1.Send).To(HaveLen(1))
			Expect(conn2.Send).To(HaveLen(1))
			Expect(<-conn1.Send).To(Equal([]byte("foo")))
			Expect(<-conn2.Send).To(Equal([]byte("foo")))
		})
	})

	Describe("RegisterHandler", func() {
		It("registers a message handler", func() {
			handlerCalled := false
			handler := func(data []byte, reply chan []byte) {
				handlerCalled = true
			}

			Expect(ws.HubHandlers_()).To(HaveLen(0))

			ws.RegisterHandler("foo", handler)
			runtime.Gosched()

			Expect(ws.HubHandlers_()).To(HaveLen(1))
			Expect(handlerCalled).To(BeFalse())
			ws.HubHandlers_()["foo"](nil, nil)
			Expect(handlerCalled).To(BeTrue())
		})
	})

	Describe("UnregisterHandler", func() {
		It("unregisters a message handler", func() {
			handler := func(data []byte, reply chan []byte) {}

			Expect(ws.HubHandlers_()).To(HaveLen(0))

			ws.RegisterHandler("foo", handler)
			runtime.Gosched()
			Expect(ws.HubHandlers_()).To(HaveLen(1))

			ws.UnregisterHandler("foo")
			runtime.Gosched()

			Expect(ws.HubHandlers_()).To(HaveLen(0))
		})
	})
})
