package audit

import (
	"context"
	"sync"
	"testing"
)

func TestMemRecorderRecordAndRead(t *testing.T) {
	r := NewMemRecorder(10, nil)
	for i := 0; i < 3; i++ {
		if err := r.Record(context.Background(), Event{Action: "user.login", Status: "success"}); err != nil {
			t.Fatal(err)
		}
	}
	if got := r.Len(); got != 3 {
		t.Errorf("Len = %d, want 3", got)
	}
	evs := r.Events()
	if len(evs) != 3 {
		t.Fatalf("Events len = %d, want 3", len(evs))
	}
	for i, ev := range evs {
		if ev.Action != "user.login" {
			t.Errorf("event[%d].Action = %q", i, ev.Action)
		}
	}
}

func TestMemRecorderRingWrap(t *testing.T) {
	r := NewMemRecorder(3, nil)
	for i := 0; i < 5; i++ {
		_ = r.Record(context.Background(), Event{Action: string(rune('a' + i))})
	}
	if got := r.Len(); got != 3 {
		t.Errorf("Len after wrap = %d, want 3 (cap)", got)
	}
	// 环形缓冲满后, 顺序应该是 c, d, e (最旧的 a, b 被覆盖)
	evs := r.Events()
	if len(evs) != 3 {
		t.Fatalf("Events len = %d", len(evs))
	}
	want := []string{"c", "d", "e"}
	for i, w := range want {
		if evs[i].Action != w {
			t.Errorf("Events[%d] = %q, want %q", i, evs[i].Action, w)
		}
	}
}

func TestMemRecorderConcurrentSafe(t *testing.T) {
	r := NewMemRecorder(100, nil)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = r.Record(context.Background(), Event{Action: "x"})
		}(i)
	}
	wg.Wait()
	if got := r.Len(); got != 50 {
		t.Errorf("Len = %d, want 50", got)
	}
}

func TestMemRecorderDefaultCap(t *testing.T) {
	r := NewMemRecorder(0, nil)
	_ = r.Record(context.Background(), Event{Action: "x"})
	if r.Len() != 1 {
		t.Errorf("Len = %d, want 1", r.Len())
	}
}
