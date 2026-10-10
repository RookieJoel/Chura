// Package lockstripe provides a fixed-size set of in-process locks selected by
// key hash, so per-key serialisation needs bounded memory.
package lockstripe

import (
	"context"
	"hash/fnv"
)

// Stripes is the number of locks in a Set.
const Stripes = 256

// Set is a striped lock set; create it with New.
type Set struct {
	stripes [Stripes]chan struct{}
}

func New() *Set {
	s := &Set{}
	for i := range s.stripes {
		s.stripes[i] = make(chan struct{}, 1)
	}
	return s
}

// Index returns the stripe for key: fnv32a(key) % Stripes.
func Index(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32() % Stripes)
}

// Lock acquires the stripe for key, waiting until it is free or ctx is done.
// The returned func releases it and must be called exactly once.
func (s *Set) Lock(ctx context.Context, key string) (func(), error) {
	sem := s.stripes[Index(key)]
	select {
	case sem <- struct{}{}:
		return func() { <-sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
