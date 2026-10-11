package lockstripe

import (
	"context"
	"errors"
	"hash/fnv"
	"testing"
	"time"
)

func TestIndex(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want int
	}{
		{"empty key", "", int(fnv.New32a().Sum32() % Stripes)},
		{"user id", "9f1c2b7e-0000-4000-8000-000000000001", func() int {
			h := fnv.New32a()
			_, _ = h.Write([]byte("9f1c2b7e-0000-4000-8000-000000000001"))
			return int(h.Sum32() % Stripes)
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Index(tc.key); got != tc.want {
				t.Errorf("Index(%q) = %d, want %d", tc.key, got, tc.want)
			}
		})
	}
	if Stripes != 256 {
		t.Errorf("Stripes = %d, want 256", Stripes)
	}
}

func TestLock_SameKeySerialisesAndHonoursContext(t *testing.T) {
	s := New()
	unlock, err := s.Lock(context.Background(), "k")
	if err != nil {
		t.Fatalf("first Lock: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := s.Lock(ctx, "k"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second Lock err = %v, want deadline exceeded", err)
	}

	unlock()
	unlock2, err := s.Lock(context.Background(), "k")
	if err != nil {
		t.Fatalf("Lock after unlock: %v", err)
	}
	unlock2()
}

func TestLock_DifferentStripesDoNotBlock(t *testing.T) {
	s := New()
	a := "a"
	b := ""
	for i := 0; ; i++ {
		b = string(rune('b' + i))
		if Index(a) != Index(b) {
			break
		}
	}
	unlock, err := s.Lock(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock2, err := s.Lock(ctx, b)
	if err != nil {
		t.Fatalf("Lock on other stripe: %v", err)
	}
	unlock2()
}
