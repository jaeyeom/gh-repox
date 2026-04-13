package github

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"not found lower", errors.New("not found"), true},
		{"HTTP 404", errors.New("HTTP 404: Not Found"), true},
		{"gh stderr", errors.New("gh: Not Found (HTTP 404): exit status 1"), true},
		{"unrelated", errors.New("permission denied"), false},
		{"403", errors.New("HTTP 403: Forbidden"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsNotFoundError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryOnNotFound_SuccessFirstTry(t *testing.T) {
	calls := 0
	err := RetryOnNotFound(context.Background(), 3, time.Millisecond, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryOnNotFound_SuccessAfterRetries(t *testing.T) {
	calls := 0
	err := RetryOnNotFound(context.Background(), 5, time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return errors.New("HTTP 404: Not Found")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryOnNotFound_Non404StopsImmediately(t *testing.T) {
	calls := 0
	err := RetryOnNotFound(context.Background(), 5, time.Millisecond, func() error {
		calls++
		return errors.New("permission denied")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected 1 call for non-404 error, got %d", calls)
	}
}

func TestRetryOnNotFound_AllAttemptsExhausted(t *testing.T) {
	calls := 0
	err := RetryOnNotFound(context.Background(), 3, time.Millisecond, func() error {
		calls++
		return errors.New("HTTP 404: Not Found")
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

func TestRetryOnNotFound_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := RetryOnNotFound(ctx, 5, time.Second, func() error {
		calls++
		cancel() // cancel after first call
		return errors.New("HTTP 404: Not Found")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected 1 call before context cancelled, got %d", calls)
	}
}
