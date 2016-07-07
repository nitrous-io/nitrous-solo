package shell

import (
	"encoding/base64"
	"encoding/json"
	"os/user"
	"testing"

	. "github.com/nitrous-io/nitrous-solo/testhelper"
	"github.com/nitrous-io/pty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/shell")
}

func resetHub() {
	hub.stop <- struct{}{}
	hub = newShellHub()
	go hub.run()
}

var currentUsername string

func init() {
	u, _ := user.Current() // this should not fail
	currentUsername = u.Username
}

var _ = Describe("pkg/shell", func() {
	var res []byte

	cb := func(r []byte) {
		res = r
	}

	BeforeEach(func() {
		resetHub()
		res = []byte{}
	})

	Describe("StartShellSessionAndAttach", func() {
		Context("when a valid command is given", func() {
			It("starts a shell session and attaches a reply func", func() {
				Expect(hub.sessions).To(HaveLen(0))
				Expect(hub.subscriptions).To(HaveLen(0))

				subID, sessID, err := StartShellSessionAndAttach("/bin/bash", []string{"-c", `echo -n "hello world"; sleep 10`}, currentUsername, 80, 25, cb)
				Expect(err).To(BeNil())

				Tick()

				Expect(hub.sessions).To(HaveKey(sessID))
				Expect(hub.sessions[sessID]).NotTo(BeNil())
				rows, cols, err := pty.Getsize(hub.sessions[sessID].pty)
				Expect(err).To(BeNil())
				Expect(rows).To(Equal(25))
				Expect(cols).To(Equal(80))

				Expect(hub.subscriptions).To(HaveKey(sessID))
				Expect(hub.subscriptions[sessID]).To(HaveLen(1))
				Expect(hub.subscriptions[sessID]).To(HaveKey(subID))
				Expect(hub.subscriptions[sessID][subID]).NotTo(BeNil())

				Expect(res).NotTo(HaveLen(0))
				var j []string
				Expect(json.Unmarshal(res, &j)).To(BeNil())
				Expect(j[0]).To(Equal("d"))
				b, err := base64.StdEncoding.DecodeString(j[1])
				Expect(err).To(BeNil())
				Expect(string(b)).To(Equal("hello world"))
			})
		})

		Context("when an invalid command is given", func() {
			It("returns error and does not attach", func() {
				subID, sessID, err := StartShellSessionAndAttach("dcjknvd", []string{"foo", "123"}, currentUsername, 80, 25, cb)
				Expect(subID).To(BeEmpty())
				Expect(sessID).To(BeEmpty())
				Expect(err).NotTo(BeNil())
				Tick()

				Expect(hub.sessions).To(HaveLen(0))
				Expect(hub.subscriptions).To(HaveLen(0))
				Expect(res).To(HaveLen(0))
			})
		})
	})

	Describe("Attach", func() {
		var sessID, subID string

		BeforeEach(func() {
			var err error
			subID, sessID, err = StartShellSessionAndAttach("/bin/bash", []string{"-c", `echo -n "hello world"; sleep 10`}, currentUsername, 80, 25, cb)
			Expect(err).To(BeNil())
			Expect(subID).NotTo(BeEmpty())
			Expect(sessID).NotTo(BeEmpty())
		})

		It("attaches an additional reply func", func() {
			var res2 []byte
			cb2 := func(r []byte) {
				res2 = r
			}

			sub2ID := Attach(sessID, 100, 50, cb2)
			Expect(sub2ID).NotTo(BeEmpty())
			Expect(subID).NotTo(Equal(sub2ID))

			Tick()
			Expect(hub.sessions).To(HaveLen(1))
			rows, cols, err := pty.Getsize(hub.sessions[sessID].pty)
			Expect(err).To(BeNil())
			Expect(rows).To(Equal(50))
			Expect(cols).To(Equal(100))

			Expect(hub.subscriptions).To(HaveLen(1))
			Expect(hub.subscriptions[sessID]).To(HaveLen(2))
			Expect(hub.subscriptions[sessID]).To(HaveKey(subID))
			Expect(hub.subscriptions[sessID]).To(HaveKey(sub2ID))
			Expect(hub.subscriptions[sessID][sub2ID]).NotTo(BeNil())

			Expect(res2).NotTo(HaveLen(0))
			var j []string
			Expect(json.Unmarshal(res2, &j)).To(BeNil())
			Expect(j[0]).To(Equal("d"))
			b, err := base64.StdEncoding.DecodeString(j[1])
			Expect(err).To(BeNil())
			Expect(string(b)).To(Equal("hello world"))
		})
	})

	Describe("Detach", func() {
		var sessID, subID string

		BeforeEach(func() {
			var err error
			subID, sessID, err = StartShellSessionAndAttach("/bin/bash", []string{"-c", `echo -n "hello world"; sleep 10`}, currentUsername, 80, 25, cb)
			Expect(err).To(BeNil())
			Expect(subID).NotTo(BeEmpty())
			Expect(sessID).NotTo(BeEmpty())
		})

		It("detaches an attached reply func", func() {
			Tick()

			Expect(hub.subscriptions).To(HaveLen(1))
			Expect(hub.subscriptions[sessID]).To(HaveLen(1))
			Expect(hub.subscriptions[sessID]).To(HaveKey(subID))

			Detach(sessID, subID)
			Tick()

			Expect(hub.subscriptions[sessID]).To(HaveLen(0))
		})
	})

	Describe("Input", func() {
		var sessID string

		cb := func(r []byte) {
			if len(r) > 0 {
				var j []string
				Expect(json.Unmarshal(r, &j)).To(BeNil())
				Expect(j[0]).To(Equal("d"))
				b, err := base64.StdEncoding.DecodeString(j[1])
				Expect(err).To(BeNil())
				res = append(res, b...)
			}
		}

		BeforeEach(func() {
			var err error
			_, sessID, err = StartShellSessionAndAttach("/bin/bash", []string{"-c", `cat /dev/stdin > /dev/stdout`}, currentUsername, 80, 25, cb)
			Expect(err).To(BeNil())
			Expect(sessID).NotTo(BeEmpty())
		})

		It("writes to the shell session's stdin", func() {
			Tick()

			Input(sessID, []byte("hello\n"))

			Tick()

			Expect(res).NotTo(HaveLen(0))
			Expect(string(res)).To(ContainSubstring("hello"))
		})
	})

	Describe("Resize", func() {
		var sessID string

		BeforeEach(func() {
			var err error
			_, sessID, err = StartShellSessionAndAttach("/bin/bash", []string{"-c", `echo -n "hello world"; sleep 10`}, currentUsername, 80, 25, cb)
			Expect(err).To(BeNil())
			Expect(sessID).NotTo(BeEmpty())
		})

		It("resizes the shell session's pseudo terminal", func() {
			Tick()

			Resize(sessID, 200, 80)
			Tick()

			Expect(hub.sessions).To(HaveKey(sessID))
			rows, cols, err := pty.Getsize(hub.sessions[sessID].pty)
			Expect(err).To(BeNil())
			Expect(cols).To(Equal(200))
			Expect(rows).To(Equal(80))
		})
	})
})
