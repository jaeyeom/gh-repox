package exec

import (
	"context"
	"testing"
)

func TestRealRunner_Echo(t *testing.T) {
	r := &RealRunner{}
	stdout, stderr, err := r.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "hello\n" {
		t.Errorf("got stdout=%q, want %q", stdout, "hello\n")
	}
	if stderr != "" {
		t.Errorf("got stderr=%q, want empty", stderr)
	}
}

func TestMockRunner(t *testing.T) {
	m := &MockRunner{
		Responses: []MockCall{
			{Stdout: "ok\n"},
		},
	}
	stdout, _, err := m.Run(context.Background(), "gh", "repo", "create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "ok\n" {
		t.Errorf("got %q, want %q", stdout, "ok\n")
	}
	if len(m.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.Calls))
	}
	if m.Calls[0].Name != "gh" {
		t.Errorf("got name=%q, want gh", m.Calls[0].Name)
	}
}

func TestMockRunner_UnexpectedCall(t *testing.T) {
	m := &MockRunner{}
	_, _, err := m.Run(context.Background(), "gh", "api")
	if err == nil {
		t.Fatal("expected error for unexpected call")
	}
}
