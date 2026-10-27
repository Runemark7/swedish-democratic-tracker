package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestTransientUpstream(t *testing.T) {
	reset := errors.New("read tcp 10.0.0.1:443: connection reset by peer")
	if !transientUpstream(reset) {
		t.Fatal("connection reset should be retried")
	}
	timeout := errors.New(`Get "https://data.riksdagen.se": context deadline exceeded (Client.Timeout exceeded while awaiting headers)`)
	if !transientUpstream(timeout) {
		t.Fatal("client timeout should be retried")
	}
	var nerr net.Error = timeoutError{}
	if !transientUpstream(nerr) {
		t.Fatal("net.Error timeout should be retried")
	}
	if transientUpstream(errors.New("decode dokumentlista: invalid character")) {
		t.Fatal("a decode error is not a reset")
	}
	if transientUpstream(nil) {
		t.Fatal("nil is not transient")
	}
}

func TestHeldComplete(t *testing.T) {
	if !heldComplete(100, 2, 98) {
		t.Fatal("unreachable voteringar are not a gap")
	}
	if heldComplete(100, 0, 99) {
		t.Fatal("a short session is not complete")
	}
	if heldComplete(0, 0, 10) {
		t.Fatal("no denominator is not complete")
	}
	if !heldComplete(100, 0, 105) {
		t.Fatal("a surplus over the listing is still complete")
	}
}

func TestRetryUpstream_RetriesResetThenSucceeds(t *testing.T) {
	prev := upstreamBackoffs
	upstreamBackoffs = []time.Duration{0, 0}
	t.Cleanup(func() { upstreamBackoffs = prev })

	n := 0
	err := retryUpstream(context.Background(), "2025/26", func() error {
		n++
		if n < 3 {
			return errors.New("connection reset by peer")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("calls = %d, want 3", n)
	}
}

func TestRetryUpstream_DoesNotRetryADecodeError(t *testing.T) {
	prev := upstreamBackoffs
	upstreamBackoffs = []time.Duration{time.Hour}
	t.Cleanup(func() { upstreamBackoffs = prev })

	n := 0
	err := retryUpstream(context.Background(), "2025/26", func() error {
		n++
		return errors.New("decode dokumentlista: invalid character")
	})
	if err == nil || n != 1 {
		t.Fatalf("err = %v, calls = %d", err, n)
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }
