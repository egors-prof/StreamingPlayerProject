package limiter

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimiterNeverExceedsCapacity(t *testing.T) {
	const capacity = 3
	l := New(capacity, time.Second)

	var active, maxSeen atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := l.Acquire(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
			defer release()
			n := active.Add(1)
			for {
				m := maxSeen.Load()
				if n <= m || maxSeen.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			active.Add(-1)
		}()
	}
	wg.Wait()

	if got := maxSeen.Load(); got > capacity {
		t.Fatalf("saw %d concurrent holders, capacity is %d", got, capacity)
	}
	if s := l.Stats(); s.Active != 0 {
		t.Fatalf("all slots must be released, active = %d", s.Active)
	}
}

func TestLimiterRejectsWhenFull(t *testing.T) {
	l := New(1, 20*time.Millisecond)
	release, err := l.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.Acquire(context.Background()); !errors.Is(err, ErrBusy) {
		t.Fatalf("want ErrBusy, got %v", err)
	}
	if l.Stats().Rejected != 1 {
		t.Fatalf("rejected counter = %d, want 1", l.Stats().Rejected)
	}

	release()
	release() // повторный вызов не должен освободить чужой слот
	if _, err := l.Acquire(context.Background()); err != nil {
		t.Fatalf("slot must be free after release: %v", err)
	}
	if l.Stats().Active != 1 {
		t.Fatalf("double release corrupted the counter: active = %d", l.Stats().Active)
	}
}

func TestLimiterWaitsForFreeSlot(t *testing.T) {
	l := New(1, time.Second)
	release, _ := l.Acquire(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		release()
	}()
	if _, err := l.Acquire(context.Background()); err != nil {
		t.Fatalf("second request must get the slot after release: %v", err)
	}
}

func TestLimiterRespectsContextCancel(t *testing.T) {
	l := New(1, time.Second)
	_, _ = l.Acquire(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := l.Acquire(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}
