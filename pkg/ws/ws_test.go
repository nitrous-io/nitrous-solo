package ws

import (
	"testing"
	"time"

	. "github.com/nitrous-io/nitrous-solo/testhelper"
	ginkgo "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	BeforeEach = ginkgo.BeforeEach
	AfterEach  = ginkgo.AfterEach
	Describe   = ginkgo.Describe
	It         = ginkgo.It
)

func Test(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "pkg/ws")
}

func resetHub() {
	oldHub := hub
	hub = newHub()
	oldHub.stop <- struct{}{}
	go hub.run()
}

var _ = Describe("pkg/ws", func() {
	BeforeEach(func() {
		resetHub()
		SetDefaultEventuallyTimeout(10 * time.Second)
	})

	AfterEach(func() {
		SetDefaultEventuallyTimeout(1 * time.Second)
	})

	Describe("registerConn()", func() {
		It("adds a connection to connections list", func() {
			conn := newConn(nil)

			Expect(hub.connections).To(HaveLen(0))

			hub.registerConn(conn)

			Expect(hub.connections).To(HaveLen(1))
			Expect(hub.connections[conn.id]).To(Equal(conn))
		})
	})

	Describe("unregisterConn()", func() {
		It("removes a connection from connections list", func() {
			conn := newConn(nil)

			hub.registerConn(conn)
			Expect(hub.connections).To(HaveLen(1))

			hub.unregisterConn(conn)

			Expect(hub.connections).To(HaveLen(0))
			Expect(conn.send).To(BeClosed())
		})
	})

	Describe("Send()", func() {
		It("sends a message to a single connection with a given id", func() {
			conn1 := newConn(nil)
			conn2 := newConn(nil)

			hub.registerConn(conn1)
			hub.registerConn(conn2)

			Expect(conn1.send).To(HaveLen(0))
			Expect(conn2.send).To(HaveLen(0))

			Send(conn1.id, "123", "foo", []byte("foobar777"))

			Eventually(func() Any {
				return conn1.send
			}).Should(HaveLen(1))

			time.Sleep(5 * time.Millisecond) // give conn2 more time to receive something
			Expect(conn2.send).To(HaveLen(0))
			Expect(<-conn1.send).To(Equal([]byte("m|foo|123|foobar777")))
		})
	})

	Describe("Broadcast()", func() {
		It("sends message to every connection", func() {
			conn1 := newConn(nil)
			conn2 := newConn(nil)

			hub.registerConn(conn1)
			hub.registerConn(conn2)

			Expect(conn1.send).To(HaveLen(0))
			Expect(conn2.send).To(HaveLen(0))

			Broadcast("foo", []byte("foobar777"))

			Eventually(func() Any {
				return conn1.send
			}).Should(HaveLen(1))

			Eventually(func() Any {
				return conn2.send
			}).Should(HaveLen(1))

			Expect(<-conn1.send).To(Equal([]byte("m|foo||foobar777")))
			Expect(<-conn2.send).To(Equal([]byte("m|foo||foobar777")))
		})
	})

	Describe("RegisterHandler()", func() {
		It("registers a message handler", func() {
			handlerCalled := false
			handler := func(c *Context) {
				handlerCalled = true
			}

			Expect(hub.handlers).To(HaveLen(0))

			RegisterHandler("foo", handler)

			Expect(hub.handlers).To(HaveLen(1))
			Expect(handlerCalled).To(BeFalse())

			hub.handlers["foo"](nil)
			Expect(handlerCalled).To(BeTrue())
		})
	})

	Describe("UnregisterHandler()", func() {
		It("unregisters a message handler", func() {
			handler := func(*Context) {}

			Expect(hub.handlers).To(HaveLen(0))

			RegisterHandler("foo", handler)
			Expect(hub.handlers).To(HaveLen(1))

			UnregisterHandler("foo")
			Expect(hub.handlers).To(HaveLen(0))
		})
	})

	Describe("parseAndHandleMessage()", func() {
		var fooHandlerCallContext, barHandlerCallContext *Context
		var conn1, conn2 *conn

		BeforeEach(func() {
			fooHandlerCallContext = nil
			barHandlerCallContext = nil

			fooHandler := func(c *Context) {
				fooHandlerCallContext = c
			}

			barHandler := func(c *Context) {
				barHandlerCallContext = c
			}

			conn1 = newConn(nil)
			conn1.meta = M{
				"meta":  "data",
				"hello": 123,
			}
			conn2 = newConn(nil)

			hub.registerConn(conn1)
			hub.registerConn(conn2)
			RegisterHandler("foo", fooHandler)
			RegisterHandler("bar", barHandler)
		})

		It("calls a handler for the message name", func() {
			parseAndHandleMessage(conn1, []byte("m|foo|123|foobar777"))
			Eventually(func() Any {
				return fooHandlerCallContext
			}).ShouldNot(BeNil())
			Expect(fooHandlerCallContext.Name).To(Equal("foo"))
			Expect(fooHandlerCallContext.Data).To(Equal([]byte("foobar777")))
			Expect(fooHandlerCallContext.Reply).NotTo(BeNil())
			Expect(fooHandlerCallContext.Meta).To(Equal(M{
				"meta":  "data",
				"hello": 123,
			}))
			Expect(fooHandlerCallContext.OnClose).NotTo(BeNil())
			Expect(fooHandlerCallContext.Done).NotTo(BeNil())

			Expect(barHandlerCallContext).To(BeNil())

			parseAndHandleMessage(conn1, []byte("m|bar|245|lol"))
			Eventually(func() Any {
				return barHandlerCallContext
			}).ShouldNot(BeNil())
			Expect(barHandlerCallContext.Name).To(Equal("bar"))
			Expect(barHandlerCallContext.Data).To(Equal([]byte("lol")))
			Expect(barHandlerCallContext.Reply).NotTo(BeNil())
			Expect(barHandlerCallContext.Meta).To(Equal(M{
				"meta":  "data",
				"hello": 123,
			}))
			Expect(barHandlerCallContext.OnClose).NotTo(BeNil())
			Expect(barHandlerCallContext.Done).NotTo(BeNil())
		})

		It("passes a reply func to handler, which when called, sends a reply back to conn", func() {
			parseAndHandleMessage(conn1, []byte("m|foo|123|foobar777"))
			Eventually(func() Any {
				return fooHandlerCallContext
			}).ShouldNot(BeNil())
			Expect(fooHandlerCallContext.Reply).NotTo(BeNil())

			Expect(conn1.send).To(HaveLen(0))
			Expect(conn2.send).To(HaveLen(0))

			fooHandlerCallContext.Reply([]byte("bazqux2016"))

			Eventually(func() Any {
				return conn1.send
			}).Should(HaveLen(1))

			time.Sleep(5 * time.Millisecond) // give conn2 more time to receive something
			Expect(conn2.send).To(HaveLen(0))
			Expect(<-conn1.send).To(Equal([]byte("m|foo|123|bazqux2016")))
		})

		It("passes an read-only onClose channel that gets called when connection is closed", func() {
			parseAndHandleMessage(conn1, []byte("m|foo|123|foobar777"))
			parseAndHandleMessage(conn1, []byte("m|bar|123|foobar777"))

			Eventually(func() Any {
				return fooHandlerCallContext
			}).ShouldNot(BeNil())
			Expect(fooHandlerCallContext.OnClose).NotTo(BeNil())

			Eventually(func() Any {
				return barHandlerCallContext
			}).ShouldNot(BeNil())
			Expect(barHandlerCallContext.OnClose).NotTo(BeNil())

			Expect(conn1.onCloseChans).To(HaveLen(2))
			for ch, _ := range conn1.onCloseChans {
				ch <- struct{}{}
			}
			Eventually(fooHandlerCallContext.OnClose).Should(Receive())
			Eventually(barHandlerCallContext.OnClose).Should(Receive())
		})

		It("passes a Done callback that lets you close and remove the onClose channel", func() {
			parseAndHandleMessage(conn1, []byte("m|foo|123|foobar777"))
			parseAndHandleMessage(conn1, []byte("m|bar|123|foobar777"))

			Eventually(func() Any {
				return fooHandlerCallContext
			}).ShouldNot(BeNil())
			Expect(fooHandlerCallContext.OnClose).NotTo(BeNil())
			Expect(fooHandlerCallContext.Done).NotTo(BeNil())

			Eventually(func() Any {
				return barHandlerCallContext
			}).ShouldNot(BeNil())
			Expect(barHandlerCallContext.OnClose).NotTo(BeNil())
			Expect(barHandlerCallContext.Done).NotTo(BeNil())

			Expect(conn1.onCloseChans).To(HaveLen(2))

			// call Done
			barHandlerCallContext.Done()

			Expect(conn1.onCloseChans).To(HaveLen(1))

			for ch, _ := range conn1.onCloseChans {
				ch <- struct{}{}
			}
			Eventually(fooHandlerCallContext.OnClose).Should(Receive())
			Eventually(barHandlerCallContext.OnClose).Should(BeClosed())
		})
	})
})
