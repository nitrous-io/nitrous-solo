package sh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"testing"

	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestExecutor is a fake implementation of the Executor interface that allows
// mocking of exec.Command.
type TestExecutor struct {
	LastCmd *exec.Cmd
}

// Command returns an *exec.Cmd that runs the test executable from "go test",
// with the single TestHelperProcess test. The actual command and args are
// passed to the test executable (and can be inspected in the TestHelperProcess
// test).
func (e *TestExecutor) Command(name string, arg ...string) *exec.Cmd {
	testBinArgs := []string{"-test.run=TestHelperProcess", "--"}
	testBinArgs = append(testBinArgs, name)
	testBinArgs = append(testBinArgs, arg...)
	cmd := exec.Command(os.Args[0], testBinArgs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	e.LastCmd = cmd

	return cmd
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command so that
// we don't actually run commands in tests.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	// We only want args after the 1st occurrence of "--".
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command given\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "echo":
		iargs := []interface{}{}
		for _, s := range args {
			if strings.HasPrefix(s, "$") {
				if e := os.Getenv(s[1:]); e != "" {
					iargs = append(iargs, e)
				}
			} else {
				iargs = append(iargs, s)
			}
		}
		fmt.Println(iargs...)
	case "exit":
		n, _ := strconv.Atoi(args[0])
		os.Exit(n)
	case "stderrfail":
		iargs := []interface{}{}
		for _, s := range args {
			iargs = append(iargs, s)
		}
		fmt.Fprintln(os.Stderr, iargs...)
		os.Exit(1)
	case "ls":
		fmt.Println("Not implemented yet.")
	}
}

var _ = Describe("sh", func() {
	var (
		res      []byte
		currUser *user.User

		origExecutor Executor
	)

	cb := func(r []byte) {
		res = r
	}

	BeforeEach(func() {
		res = nil

		// Replace executor with a fake.
		origExecutor = executor
		executor = &TestExecutor{}

		var err error
		currUser, err = user.Current()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		executor = origExecutor
	})

	Describe("Run", func() {
		Context("when command is valid and runs successfully", func() {
			It("returns the stdout, stderr and exit code", func() {
				Run(&ws.Context{
					Data: []byte(`{
						"cmd": "echo",
						"args": ["Running!"],
						"run_as": "` + currUser.Username + `"
					}`),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  currUser.Username,
					},
				})
				Expect(res).NotTo(BeNil())

				var j struct {
					Stdout   string `json:"stdout"`
					Stderr   string `json:"stderr"`
					ExitCode int    `json:"exit_code"`
					Error    string `json:"error"`
				}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j.Error).To(Equal(""))

				Expect(j.ExitCode).To(Equal(0))
				Expect(j.Stderr).To(Equal(""))
				Expect(j.Stdout).To(Equal("Running!\n"))
			})

			It("runs the command as the given user", func() {
				Run(&ws.Context{
					Data: []byte(`{
						"cmd": "echo",
						"args": ["Running!"],
						"run_as": "` + currUser.Username + `"
					}`),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  currUser.Username,
					},
				})
				Expect(res).NotTo(BeNil())

				te, ok := executor.(*TestExecutor)
				Expect(ok).To(BeTrue())
				Expect(te.LastCmd).NotTo(BeNil())

				cred := te.LastCmd.SysProcAttr.Credential
				Expect(cred).NotTo(BeNil())
				uid, err := strconv.Atoi(currUser.Uid)
				Expect(err).To(BeNil())
				Expect(cred.Uid).To(Equal(uint32(uid)))
				gid, err := strconv.Atoi(currUser.Gid)
				Expect(err).To(BeNil())
				Expect(cred.Gid).To(Equal(uint32(gid)))
			})

			Context("run_as is empty", func() {
				It("runs the command as current user", func() {
					Run(&ws.Context{
						Data: []byte(`{
							"cmd": "echo",
							"args": ["Running!"]
						}`),
						Reply: cb,
						Meta: map[string]interface{}{
							"login": "foo",
							"user":  currUser.Username,
						},
					})
					Expect(res).NotTo(BeNil())

					te, ok := executor.(*TestExecutor)
					Expect(ok).To(BeTrue())
					Expect(te.LastCmd).NotTo(BeNil())

					cred := te.LastCmd.SysProcAttr.Credential
					Expect(cred).NotTo(BeNil())
					uid, err := strconv.Atoi(currUser.Uid)
					Expect(err).To(BeNil())
					Expect(cred.Uid).To(Equal(uint32(uid)))
					gid, err := strconv.Atoi(currUser.Gid)
					Expect(err).To(BeNil())
					Expect(cred.Gid).To(Equal(uint32(gid)))
				})
			})

			Context("when a working directory is given", func() {
				var tmpDir string

				BeforeEach(func() {
					var err error
					tmpDir, err = ioutil.TempDir("", "test-run")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					os.RemoveAll(tmpDir)
				})

				It("runs the command from that directory", func() {
					Run(&ws.Context{
						Data: []byte(`{
							"cmd": "does not matter",
							"wd": "` + tmpDir + `",
							"run_as": "` + currUser.Username + `"
						}`),
						Reply: cb,
						Meta: map[string]interface{}{
							"login": "foo",
							"user":  currUser.Username,
						},
					})
					Expect(res).NotTo(BeNil())

					te, ok := executor.(*TestExecutor)
					Expect(ok).To(BeTrue())
					Expect(te.LastCmd).NotTo(BeNil())
					Expect(te.LastCmd.Dir).To(Equal(tmpDir))
				})

				It("returns an error when working directory does not exist", func() {
					Run(&ws.Context{
						Data: []byte(`{
							"cmd": "does not matter",
							"wd": "` + tmpDir + `-non-existent",
							"run_as": "` + currUser.Username + `"
						}`),
						Reply: cb,
						Meta: map[string]interface{}{
							"login": "foo",
							"user":  currUser.Username,
						},
					})
					Expect(res).NotTo(BeNil())

					var j map[string]interface{}
					err := json.Unmarshal(res, &j)
					Expect(err).To(BeNil())
					Expect(j["error"]).To(ContainSubstring("no such file or directory"))
				})
			})

			Context("when environment variables are given", func() {
				var tmpDir string

				BeforeEach(func() {
					var err error
					tmpDir, err = ioutil.TempDir("", "test-run")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					os.RemoveAll(tmpDir)
				})

				It("runs the command with given env", func() {
					Run(&ws.Context{
						Data: []byte(`{
							"cmd": "echo",
							"args": ["$FOO", "$BAZ", "!"],
							"env": ["FOO=BAR","BAZ=QUX"],
							"run_as": "` + currUser.Username + `"
						}`),
						Reply: cb,
						Meta: map[string]interface{}{
							"login": "foo",
							"user":  currUser.Username,
						},
					})
					Expect(res).NotTo(BeNil())

					var j struct {
						Stdout   string `json:"stdout"`
						Stderr   string `json:"stderr"`
						ExitCode int    `json:"exit_code"`
						Error    string `json:"error"`
					}
					err := json.Unmarshal(res, &j)
					Expect(err).To(BeNil())
					Expect(j.Error).To(Equal(""))

					Expect(j.ExitCode).To(Equal(0))
					Expect(j.Stdout).To(Equal("BAR QUX !\n"))
					Expect(j.Stderr).To(Equal(""))
				})
			})
		})

		Context("when invalid JSON is given", func() {
			It("returns an error", func() {
				Run(&ws.Context{
					Data:  []byte(`{invalid JSON`),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  currUser.Username,
					},
				})

				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("invalid_json"))
			})
		})

		Context("when cmd is empty", func() {
			It("returns an error", func() {
				Run(&ws.Context{
					Data: []byte(`{
						"cmd": "",
						"run_as": "` + currUser.Username + `"
					}`),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  currUser.Username,
					},
				})

				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("invalid_params"))
			})
		})

		Context("when an unknown user is given", func() {
			It("returns an error", func() {
				Run(&ws.Context{
					Data: []byte(`{
						"cmd": "does not matter",
						"run_as": "non-existent-user"
					}`),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  currUser.Username,
					},
				})

				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("unknown user"))
			})
		})

		Context("when command exits with a non-zero exit code", func() {
			It("returns the exit code", func() {
				Run(&ws.Context{
					Data: []byte(`{
						"cmd": "exit",
						"args": ["42"],
						"run_as": "` + currUser.Username + `"
					}`),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  currUser.Username,
					},
				})
				Expect(res).NotTo(BeNil())

				var j struct {
					ExitCode int `json:"exit_code"`
				}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j.ExitCode).To(Equal(42))
			})
		})

		Context("when command writes to stderr", func() {
			It("returns the stderr", func() {
				Run(&ws.Context{
					Data: []byte(`{
						"cmd": "stderrfail",
						"args": ["some", "error", "text"],
						"run_as": "` + currUser.Username + `"
					}`),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  currUser.Username,
					},
				})
				Expect(res).NotTo(BeNil())

				var j struct {
					Stderr string `json:"stderr"`
				}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())

				Expect(j.Stderr).To(Equal("some error text\n"))
			})
		})
	})
})
