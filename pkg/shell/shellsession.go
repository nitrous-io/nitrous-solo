package shell

import (
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/nitrous-io/pty"
	"github.com/satori/go.uuid"
)

type shellSession struct {
	id  string
	cmd *exec.Cmd
	pty *os.File
}

type shellWriter struct {
	shellSessionID string
	kind           string
}

// publish output to subscribers
func (w *shellWriter) Write(p []byte) (n int, err error) {
	// copy data before publishing because shellhub processes data asynchronously
	// and io.Copy reuses the buffer (overwrites data at the same address)
	data := make([]byte, len(p))
	copy(data, p)
	hub.publish <- shellOutput{
		shellSessionID: w.shellSessionID,
		kind:           w.kind,
		data:           data,
	}
	return len(data), nil
}

func startShellSession(cmd string, args, env []string, wd, runAs string) (*shellSession, error) {
	u, err := user.Lookup(runAs)
	if err != nil {
		return nil, err
	}

	c := exec.Command(cmd, args...)
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}

	if env != nil && len(env) > 0 {
		c.Env = env
	}

	if wd != "" {
		c.Dir = wd
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, err
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, err
	}

	c.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}

	id := uuid.NewV4().String()

	s := &shellSession{
		id:  id,
		cmd: c,
	}

	outputWriter := &shellWriter{
		shellSessionID: id,
		kind:           "d",
	}

	p, err := pty.Start(c)
	if err != nil {
		return nil, err
	}
	s.pty = p

	go io.Copy(outputWriter, p)
	go func() {
		c.Wait()
		p.Close()
		hub.publish <- shellOutput{
			shellSessionID: id,
			kind:           "exit",
			data:           []byte(c.ProcessState.String()),
		}
		hub.finish <- id
	}()

	return s, nil
}
