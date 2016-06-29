package ws_test

import (
	"runtime"

	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("hub", func() {
	BeforeEach(func() {
		ws.ResetHub()
	})

	Describe("RegisterConn", func() {
		It("adds a connection to connections list", func() {
			conn := &ws.Conn{}
			hub := ws.GetHub()

			Expect(hub.Connections).To(HaveLen(0))
			hub.RegisterConn <- conn
			runtime.Gosched()

			Expect(hub.Connections).To(HaveLen(1))
			Expect(hub.Connections[conn]).NotTo(BeNil())
		})
	})

	Describe("UnregisterConn", func() {
		It("removes a connection from connections list", func() {
			conn := &ws.Conn{Send: make(chan []byte, 10)}
			hub := ws.GetHub()

			hub.RegisterConn <- conn
			Expect(hub.Connections).To(HaveLen(1))
			hub.UnregisterConn <- conn
			runtime.Gosched()

			Expect(hub.Connections).To(HaveLen(0))
			Expect(hub.Connections).NotTo(HaveKey(conn))
			Expect(conn.Send).To(BeClosed())
		})
	})

	Describe("Broadcast", func() {
		It("sends message to every connection", func() {
			conn1 := &ws.Conn{Send: make(chan []byte, 10)}
			conn2 := &ws.Conn{Send: make(chan []byte, 10)}
			hub := ws.GetHub()

			hub.RegisterConn <- conn1
			hub.RegisterConn <- conn2
			runtime.Gosched()

			Expect(conn1.Send).To(HaveLen(0))
			Expect(conn2.Send).To(HaveLen(0))

			hub.Broadcast <- []byte("foo")
			runtime.Gosched()

			Expect(conn1.Send).To(HaveLen(1))
			Expect(conn2.Send).To(HaveLen(1))
			Expect(<-conn1.Send).To(Equal([]byte("foo")))
			Expect(<-conn2.Send).To(Equal([]byte("foo")))
		})
	})
})
