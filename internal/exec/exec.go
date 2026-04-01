package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Runner abstracts subprocess execution for testability.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (stdout string, stderr string, err error)
}

// RealRunner executes real subprocesses.
type RealRunner struct{}

// Compile-time check that RealRunner implements Runner.
var _ Runner = (*RealRunner)(nil)

func (r *RealRunner) Run(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// MockCall records a single mock call and its result.
type MockCall struct {
	Name   string
	Args   []string
	Stdout string
	Stderr string
	Err    error
}

// MockRunner records calls and returns preconfigured responses.
type MockRunner struct {
	Calls     []MockCall
	Responses []MockCall
	callIdx   int
}

func (m *MockRunner) Run(_ context.Context, name string, args ...string) (string, string, error) {
	call := MockCall{Name: name, Args: args}
	m.Calls = append(m.Calls, call)
	if m.callIdx < len(m.Responses) {
		resp := m.Responses[m.callIdx]
		m.callIdx++
		return resp.Stdout, resp.Stderr, resp.Err
	}
	m.callIdx++
	return "", "", fmt.Errorf("unexpected call: %s %v", name, args)
}
