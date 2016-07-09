package sh_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/nitrous-io/nitrous-solo/handlers/sh"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	"github.com/nitrous-io/nitrous-solo/testhelper/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "sh")
}

var (
	origShellStartAndAttach = sh.ShellStartAndAttach
	origShellAttach         = sh.ShellAttach
	origShellDetach         = sh.ShellDetach
	origShellInput          = sh.ShellInput
	origShellResize         = sh.ShellResize
	origShellKill           = sh.ShellKill
	origShellList           = sh.ShellList
)

var _ = Describe("sh", func() {
	var res []byte

	cb := func(r []byte) {
		res = r
	}

	parseResponseJSON := func() (j map[string]interface{}) {
		err := json.Unmarshal(res, &j)
		Expect(err).To(BeNil())
		return j
	}

	BeforeEach(func() {
		res = nil
		origShellStartAndAttach = sh.ShellStartAndAttach
		origShellAttach = sh.ShellAttach
		origShellDetach = sh.ShellDetach
		origShellInput = sh.ShellInput
		origShellResize = sh.ShellResize
		origShellKill = sh.ShellKill
		origShellList = sh.ShellList
		sh.ShellStartAndAttach = fake.ShellStartAndAttach
		sh.ShellAttach = fake.ShellAttach
		sh.ShellDetach = fake.ShellDetach
		sh.ShellInput = fake.ShellInput
		sh.ShellResize = fake.ShellResize
		sh.ShellKill = fake.ShellKill
		sh.ShellList = fake.ShellList
		fake.FakeShellReset()
	})

	AfterEach(func() {
		sh.ShellStartAndAttach = origShellStartAndAttach
		sh.ShellAttach = origShellAttach
		sh.ShellDetach = origShellDetach
		sh.ShellInput = origShellInput
		sh.ShellResize = origShellResize
		sh.ShellKill = origShellKill
		sh.ShellList = origShellList
	})

	Describe("Start", func() {
		var params H

		BeforeEach(func() {
			params = H{
				"cmd":    "/bin/ls",
				"args":   []string{"-l"},
				"env":    []string{"APP_ENV=production", "TERM=xterm-256color"},
				"run_as": "nitrous",
				"wd":     "/home/nitrous",
				"cols":   80,
				"rows":   25,
			}
		})

		Context("when 'cmd' param is missing", func() {
			It("replies with an error", func() {
				delete(params, "cmd")
				sh.Start("sh.start", EncodeJSON(params), cb)

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("invalid_params"))
			})
		})

		Context("when 'run_as' param is missing", func() {
			It("replies with an error", func() {
				delete(params, "run_as")
				sh.Start("sh.start", EncodeJSON(params), cb)

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("invalid_params"))
			})
		})

		Context("when all params are valid", func() {
			Context("when shell session fails to start", func() {
				It("replies with an error", func() {
					fake.ShellStartAndAttachError = errors.New("command not found")
					sh.Start("sh.start", EncodeJSON(params), cb)

					j := parseResponseJSON()
					Expect(j["error"]).To(Equal("failed_to_start_shell_session"))
					Expect(j["error_description"]).To(Equal("command not found"))
				})
			})

			Context("when cols and rows are really low", func() {
				It("starts a shell session with minimum cols and rows and replies with subscription and shell session ids", func() {
					params["cols"] = 1
					params["rows"] = 1
					sh.Start("sh.start", EncodeJSON(params), cb)

					Expect(fake.ShellStartAndAttachCalls.Count()).To(Equal(1))
					call := fake.ShellStartAndAttachCalls.NthCall(1)

					Expect(call).NotTo(BeNil())
					Expect(call.Arguments[0]).To(Equal(params["cmd"]))
					Expect(call.Arguments[1]).To(Equal(params["args"]))
					Expect(call.Arguments[2]).To(Equal(params["env"]))
					Expect(call.Arguments[3]).To(Equal(params["wd"]))
					Expect(call.Arguments[4]).To(Equal(params["run_as"]))
					Expect(call.Arguments[5]).To(Equal(uint(6)))
					Expect(call.Arguments[6]).To(Equal(uint(2)))
					ExpectSameFunc(call.Arguments[7], cb)

					j := parseResponseJSON()
					Expect(j["subscription_id"]).To(Equal(call.ReturnValues[0]))
					Expect(j["shell_session_id"]).To(Equal(call.ReturnValues[1]))
				})
			})

			Context("when there is no other issue", func() {
				It("starts a shell session and replies with subscription and shell session ids", func() {
					sh.Start("sh.start", EncodeJSON(params), cb)

					Expect(fake.ShellStartAndAttachCalls.Count()).To(Equal(1))
					call := fake.ShellStartAndAttachCalls.NthCall(1)

					Expect(call).NotTo(BeNil())
					Expect(call.Arguments[0]).To(Equal(params["cmd"]))
					Expect(call.Arguments[1]).To(Equal(params["args"]))
					Expect(call.Arguments[2]).To(Equal(params["env"]))
					Expect(call.Arguments[3]).To(Equal(params["wd"]))
					Expect(call.Arguments[4]).To(Equal(params["run_as"]))
					Expect(call.Arguments[5]).To(Equal(uint(80)))
					Expect(call.Arguments[6]).To(Equal(uint(25)))
					ExpectSameFunc(call.Arguments[7], cb)

					j := parseResponseJSON()
					Expect(j["subscription_id"]).To(Equal(call.ReturnValues[0]))
					Expect(j["shell_session_id"]).To(Equal(call.ReturnValues[1]))
				})
			})
		})
	})

	Describe("Attach", func() {
		var params H

		BeforeEach(func() {
			params = H{
				"shell_session_id": "1234",
				"cols":             80,
				"rows":             25,
			}
		})

		Context("when 'shell_session_id' param is missing", func() {
			It("replies with an error", func() {
				delete(params, "shell_session_id")
				sh.Attach("sh.attach", EncodeJSON(params), cb)

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("invalid_params"))
			})
		})

		Context("when cols and rows are really low", func() {
			It("attaches with minimum cols and rows and replies with a subscription id", func() {
				params["cols"] = 1
				params["rows"] = 1
				sh.Attach("sh.attach", EncodeJSON(params), cb)

				Expect(fake.ShellAttachCalls.Count()).To(Equal(1))
				call := fake.ShellAttachCalls.NthCall(1)

				Expect(call).NotTo(BeNil())
				Expect(call.Arguments[0]).To(Equal("1234"))
				Expect(call.Arguments[1]).To(Equal(uint(6)))
				Expect(call.Arguments[2]).To(Equal(uint(2)))
				ExpectSameFunc(call.Arguments[3], cb)

				j := parseResponseJSON()
				Expect(j["subscription_id"]).To(Equal(call.ReturnValues[0]))
			})
		})

		Context("when there is no other issue", func() {
			It("attaches to a shell session and replies with a subscription id", func() {
				sh.Attach("sh.attach", EncodeJSON(params), cb)

				Expect(fake.ShellAttachCalls.Count()).To(Equal(1))
				call := fake.ShellAttachCalls.NthCall(1)

				Expect(call).NotTo(BeNil())
				Expect(call.Arguments[0]).To(Equal("1234"))
				Expect(call.Arguments[1]).To(Equal(uint(80)))
				Expect(call.Arguments[2]).To(Equal(uint(25)))
				ExpectSameFunc(call.Arguments[3], cb)

				j := parseResponseJSON()
				Expect(j["subscription_id"]).To(Equal(call.ReturnValues[0]))
			})
		})
	})

	Describe("Detach", func() {
		var params H

		BeforeEach(func() {
			params = H{
				"shell_session_id": "1234",
				"subscription_id":  "5678",
			}
		})

		Context("when 'shell_session_id' param is missing", func() {
			It("replies with an error", func() {
				delete(params, "shell_session_id")
				sh.Detach("sh.detach", EncodeJSON(params), cb)

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("invalid_params"))
			})
		})

		Context("when 'subscription_id' param is missing", func() {
			It("replies with an error", func() {
				delete(params, "subscription_id")
				sh.Detach("sh.detach", EncodeJSON(params), cb)

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("invalid_params"))
			})
		})

		Context("when there is no other issue", func() {
			It("detaches from the shell session with minimum cols and rows and replies with subscription id", func() {
				sh.Detach("sh.detach", EncodeJSON(params), cb)

				Expect(fake.ShellDetachCalls.Count()).To(Equal(1))
				call := fake.ShellDetachCalls.NthCall(1)

				Expect(call).NotTo(BeNil())
				Expect(call.Arguments[0]).To(Equal("1234"))
				Expect(call.Arguments[1]).To(Equal("5678"))

				j := parseResponseJSON()
				Expect(j["detached"]).To(BeTrue())
			})
		})
	})

	Describe("Input", func() {
		var params A

		BeforeEach(func() {
			params = A{
				"1234",
				base64.StdEncoding.EncodeToString([]byte("vim\n")),
			}
		})

		Context("when the params array is empty", func() {
			It("does nothing", func() {
				params = A{}
				sh.Input("sh.i", EncodeJSON(params), cb)

				Expect(fake.ShellInputCalls.Count()).To(BeZero())
			})
		})

		Context("when the params array has only one element", func() {
			It("does nothing", func() {
				params = A{"1234"}
				sh.Input("sh.i", EncodeJSON(params), cb)

				Expect(fake.ShellInputCalls.Count()).To(BeZero())
			})
		})

		Context("when params are valid", func() {
			It("decodes the base64-encoded data and sends an input to a shell session", func() {
				sh.Input("sh.i", EncodeJSON(params), cb)

				Expect(fake.ShellInputCalls.Count()).To(Equal(1))
				call := fake.ShellInputCalls.NthCall(1)

				Expect(call).NotTo(BeNil())
				Expect(call.Arguments[0]).To(Equal("1234"))
				Expect(call.Arguments[1]).To(Equal([]byte("vim\n")))
			})
		})
	})

	Describe("Resize", func() {
		var params A

		BeforeEach(func() {
			params = A{
				"1234",
				80,
				25,
			}
		})

		Context("when the params array is empty", func() {
			It("does nothing", func() {
				params = A{}
				sh.Resize("sh.r", EncodeJSON(params), cb)

				Expect(fake.ShellResizeCalls.Count()).To(BeZero())
			})

		})

		Context("when the params array has less than three elements", func() {
			It("does nothing", func() {
				params = A{"1234", 80}
				sh.Resize("sh.r", EncodeJSON(params), cb)

				Expect(fake.ShellResizeCalls.Count()).To(BeZero())
			})
		})

		Context("when params are valid", func() {
			It("sends a resize message to a shell session", func() {
				sh.Resize("sh.r", EncodeJSON(params), cb)

				Expect(fake.ShellResizeCalls.Count()).To(Equal(1))
				call := fake.ShellResizeCalls.NthCall(1)

				Expect(call).NotTo(BeNil())
				Expect(call.Arguments[0]).To(Equal("1234"))
				Expect(call.Arguments[1]).To(Equal(uint(80)))
				Expect(call.Arguments[2]).To(Equal(uint(25)))
			})
		})
	})

	Describe("Kill", func() {
		var params H

		BeforeEach(func() {
			params = H{
				"shell_session_id": "1234",
			}
		})

		Context("when 'shell_session_id' param is missing", func() {
			It("does nothing", func() {
				delete(params, "shell_session_id")
				sh.Kill("sh.kill", EncodeJSON(params), cb)

				Expect(fake.ShellKillCalls.Count()).To(BeZero())
			})
		})

		Context("when params are valid", func() {
			It("sends a kill message to a shell session", func() {
				sh.Kill("sh.kill", EncodeJSON(params), cb)

				Expect(fake.ShellKillCalls.Count()).To(Equal(1))
				call := fake.ShellKillCalls.NthCall(1)

				Expect(call).NotTo(BeNil())
				Expect(call.Arguments[0]).To(Equal("1234"))
			})
		})
	})

	Describe("List", func() {
		It("asks for a list of shell sessions with a reply func", func() {
			sh.List("sh.list", nil, cb)

			Expect(fake.ShellListCalls.Count()).To(Equal(1))
			call := fake.ShellListCalls.NthCall(1)

			Expect(call).NotTo(BeNil())
			ExpectSameFunc(call.Arguments[0], cb)
		})
	})
})
