package ws

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func resetHub() {
	hub.stop <- struct{}{}
	hub = newHub()
	go hub.run()
}

var _ = Describe("ws/hub", func() {
	BeforeEach(func() {
		resetHub()
	})

	Describe("registerConn channel", func() {
		It("adds a connection to connections list", func() {
			conn := newConn(nil)

			Expect(hub.connections).To(HaveLen(0))

			hub.registerConn <- conn
			runtime.Gosched()

			Expect(hub.connections).To(HaveLen(1))
			Expect(hub.connections[conn.id]).To(Equal(conn))
		})
	})

	Describe("unregisterConn channel", func() {
		It("removes a connection from connections list", func() {
			conn := newConn(nil)

			hub.registerConn <- conn
			runtime.Gosched()
			Expect(hub.connections).To(HaveLen(1))

			hub.unregisterConn <- conn
			runtime.Gosched()

			Expect(hub.connections).To(HaveLen(0))
			Expect(conn.send).To(BeClosed())
		})
	})

	Describe("Send()", func() {
		It("sends a message to a single connection with a given id", func() {
			conn1 := newConn(nil)
			conn2 := newConn(nil)

			hub.registerConn <- conn1
			hub.registerConn <- conn2
			runtime.Gosched()

			Expect(conn1.send).To(HaveLen(0))
			Expect(conn2.send).To(HaveLen(0))

			Send(conn1.id, "123", "foo", []byte("foobar777"))
			runtime.Gosched()

			Expect(conn1.send).To(HaveLen(1))
			Expect(conn2.send).To(HaveLen(0))
			Expect(<-conn1.send).To(Equal([]byte("m|foo|123|foobar777")))
		})
	})

	Describe("Broadcast()", func() {
		It("sends message to every connection", func() {
			conn1 := newConn(nil)
			conn2 := newConn(nil)

			hub.registerConn <- conn1
			hub.registerConn <- conn2
			runtime.Gosched()

			Expect(conn1.send).To(HaveLen(0))
			Expect(conn2.send).To(HaveLen(0))

			Broadcast("foo", []byte("foobar777"))
			runtime.Gosched()

			Expect(conn1.send).To(HaveLen(1))
			Expect(conn2.send).To(HaveLen(1))
			Expect(<-conn1.send).To(Equal([]byte("m|foo||foobar777")))
			Expect(<-conn2.send).To(Equal([]byte("m|foo||foobar777")))
		})
	})

	Describe("RegisterHandler()", func() {
		It("registers a message handler", func() {
			handlerCalled := false
			handler := func(name string, data []byte, reply func([]byte)) {
				handlerCalled = true
			}

			Expect(hub.handlers).To(HaveLen(0))

			RegisterHandler("foo", handler)
			runtime.Gosched()

			Expect(hub.handlers).To(HaveLen(1))
			Expect(handlerCalled).To(BeFalse())

			hub.handlers["foo"]("foo", nil, nil)
			Expect(handlerCalled).To(BeTrue())
		})
	})

	Describe("UnregisterHandler()", func() {
		It("unregisters a message handler", func() {
			handler := func(name string, data []byte, reply func([]byte)) {}

			Expect(hub.handlers).To(HaveLen(0))

			RegisterHandler("foo", handler)
			runtime.Gosched()
			Expect(hub.handlers).To(HaveLen(1))

			UnregisterHandler("foo")
			runtime.Gosched()

			Expect(hub.handlers).To(HaveLen(0))
		})
	})

	Describe("parseAndHandleMessage()", func() {
		type handlerCallArgs struct {
			name  string
			data  []byte
			reply func([]byte)
		}

		var fooHandlerCallArgs, barHandlerCallArgs *handlerCallArgs
		var conn1, conn2 *conn

		BeforeEach(func() {
			fooHandlerCallArgs = nil
			barHandlerCallArgs = nil

			fooHandler := func(name string, data []byte, reply func([]byte)) {
				fooHandlerCallArgs = &handlerCallArgs{name, data, reply}
			}

			barHandler := func(name string, data []byte, reply func([]byte)) {
				barHandlerCallArgs = &handlerCallArgs{name, data, reply}
			}

			conn1 = newConn(nil)
			conn2 = newConn(nil)

			hub.registerConn <- conn1
			hub.registerConn <- conn2
			RegisterHandler("foo", fooHandler)
			RegisterHandler("bar", barHandler)
			runtime.Gosched()

			parseAndHandleMessage(conn1, []byte("m|foo|123|foobar777"))
			runtime.Gosched()
		})

		It("calls a handler for the message name", func() {
			Expect(fooHandlerCallArgs).NotTo(BeNil())
			Expect(fooHandlerCallArgs.name).To(Equal("foo"))
			Expect(fooHandlerCallArgs.data).To(Equal([]byte("foobar777")))
			Expect(fooHandlerCallArgs.reply).NotTo(BeNil())
			Expect(barHandlerCallArgs).To(BeNil())
		})

		It("passes a reply func to handler, which when called, sends a reply back to conn", func() {
			Expect(fooHandlerCallArgs).NotTo(BeNil())
			Expect(fooHandlerCallArgs.reply).NotTo(BeNil())

			Expect(conn1.send).To(HaveLen(0))
			Expect(conn2.send).To(HaveLen(0))

			fooHandlerCallArgs.reply([]byte("bazqux2016"))
			runtime.Gosched()

			Expect(conn1.send).To(HaveLen(1))
			Expect(conn2.send).To(HaveLen(0))
			Expect(<-conn1.send).To(Equal([]byte("m|foo|123|bazqux2016")))
		})
	})
})
