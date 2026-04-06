package main

import (
	"testing"
	"time"
)

func TestOrReturnsNilForNoChannels(t *testing.T) {
	t.Parallel()

	for _, isRecursive := range []bool{false, true} {
		if got := Or(isRecursive); got != nil {
			t.Fatalf("Or(%v) = %v, want nil", isRecursive, got)
		}
	}
}

func TestOrReturnsSameChannelForSingleInput(t *testing.T) {
	t.Parallel()

	ch := make(chan any)
	for _, isRecursive := range []bool{false, true} {
		if got := Or(isRecursive, ch); got != ch {
			t.Fatalf("Or(%v) returned a different channel", isRecursive)
		}
	}
}

func TestOrClosesWhenAnyChannelCloses(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		isRecursive bool
	}{
		{name: "non_recursive", isRecursive: false},
		{name: "recursive", isRecursive: true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ch1 := make(chan any)
			ch2 := make(chan any)
			ch3 := make(chan any)

			orCh := Or(tc.isRecursive, ch1, ch2, ch3)

			close(ch2)

			waitClosed(t, orCh)
		})
	}
}

func TestOrClosesWhenAnyChannelReceivesValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		isRecursive bool
	}{
		{name: "non_recursive", isRecursive: false},
		{name: "recursive", isRecursive: true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ch1 := make(chan any)
			ch2 := make(chan any, 1)
			ch3 := make(chan any)

			orCh := Or(tc.isRecursive, ch1, ch2, ch3)

			ch2 <- "signal"

			waitClosed(t, orCh)
		})
	}
}

func waitClosed(t *testing.T, ch <-chan any) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for merged channel to close")
	}
}
