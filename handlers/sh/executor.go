package sh

import "os/exec"

type Executor interface {
	Command(name string, arg ...string) *exec.Cmd
}

type StdExecutor struct{}

func (e *StdExecutor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
