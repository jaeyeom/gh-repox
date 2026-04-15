package exec

import (
	"context"
	"fmt"

	"github.com/jaeyeom/go-cmdexec"
)

// Runner abstracts subprocess execution for testability.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (stdout string, stderr string, err error)
}

// RealRunner executes real subprocesses via go-cmdexec.
type RealRunner struct {
	executor cmdexec.Executor
}

// Compile-time check that RealRunner implements Runner.
var _ Runner = (*RealRunner)(nil)

func (r *RealRunner) Run(ctx context.Context, name string, args ...string) (string, string, error) {
	if r.executor == nil {
		r.executor = cmdexec.NewBasicExecutor()
	}
	result, err := r.executor.Execute(ctx, cmdexec.ToolConfig{
		Command: name,
		Args:    args,
	})
	if err != nil {
		return "", "", fmt.Errorf("command execution failed: %w", err)
	}
	if result.ExitCode != 0 {
		return result.Output, result.Stderr, fmt.Errorf("exit status %d", result.ExitCode)
	}
	return result.Output, result.Stderr, nil
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
