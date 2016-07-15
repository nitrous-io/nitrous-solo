package sh

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestExecutor is a fake implementation of the Executor interface that allows
// mocking of exec.Command.
type TestExecutor struct {
	LastCmd *exec.Cmd
}

func (e *TestExecutor) Command(name string, arg ...string) *exec.Cmd {
	e.LastCmd = exec.Command(name, arg...)
	return e.LastCmd
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

				stdout, err := base64.StdEncoding.DecodeString(j.Stdout)
				Expect(err).To(BeNil())
				Expect(stdout).To(Equal([]uint8("Running!\n")))
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
							"cmd": "ls",
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
							"cmd": "/bin/bash",
							"args": ["-c", "echo $FOO $BAZ !"],
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
					Expect(j.Stderr).To(Equal(""))

					stdout, err := base64.StdEncoding.DecodeString(j.Stdout)
					Expect(err).To(BeNil())
					Expect(stdout).To(Equal([]uint8("BAR QUX !\n")))
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
				Expect(j["error"]).To(Equal("invalid_user"))
			})
		})

		Context("when command exits with a non-zero exit code", func() {
			It("returns the exit code", func() {
				Run(&ws.Context{
					Data: []byte(`{
						"cmd": "/bin/bash",
						"args": ["-c", "exit 42"],
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
						"cmd": "/bin/bash",
						"args": ["-c", "echo -n 'some error text' > /dev/stderr"],
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

				stderr, err := base64.StdEncoding.DecodeString(j.Stderr)
				Expect(err).To(BeNil())
				Expect(stderr).To(Equal([]uint8("some error text")))
			})
		})
	})
})
