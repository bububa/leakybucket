package memory

import (
	"github.com/bububa/leakybucket"
	"time"
)

type bucket struct {
	capacity  uint
	remaining uint
	reset     time.Time
	rate      time.Duration
	updated   time.Time
}

func (b *bucket) Capacity() uint {
	return b.capacity
}

// Remaining space in the bucket.
func (b *bucket) Remaining() uint {
	return b.remaining
}

// Reset returns when the bucket will be drained.
func (b *bucket) Reset() time.Time {
	return b.reset
}

// Add to the bucket.
func (b *bucket) Add(amount uint) (leakybucket.BucketState, error) {
	b.updated = time.Now()
	if time.Now().After(b.reset) {
		b.reset = time.Now().Add(b.rate)
		b.remaining = b.capacity
	}
	if amount > b.remaining {
		return leakybucket.BucketState{b.capacity, b.remaining, b.reset}, leakybucket.ErrorFull
	}
	b.remaining -= amount
	return leakybucket.BucketState{b.capacity, b.remaining, b.reset}, nil
}

func (b *bucket) AddWithTime(amount uint, t time.Time) (leakybucket.BucketState, error) {
	b.updated = time.Now()
	if t.After(b.reset) {
		b.reset = t.Add(b.rate)
		b.remaining = b.capacity
	}
	if t.Before(b.reset.Add(-1 * b.rate)) {
		b.reset = t.Add(b.rate)
	}
	if amount > b.remaining {
		return leakybucket.BucketState{b.capacity, b.remaining, b.reset}, leakybucket.ErrorFull
	}
	b.remaining -= amount
	return leakybucket.BucketState{b.capacity, b.remaining, b.reset}, nil
}

// Storage is a non thread-safe in-memory leaky bucket factory.
type Storage struct {
	buckets map[string]*bucket
}

// New initializes the in-memory bucket store.
func New() *Storage {
	return &Storage{
		buckets: make(map[string]*bucket),
	}
}

// Create a bucket.
func (s *Storage) Create(name string, capacity uint, rate time.Duration) (leakybucket.Bucket, error) {
	b, ok := s.buckets[name]
	if ok {
		return b, nil
	}
	b = &bucket{
		capacity:  capacity,
		remaining: capacity,
		reset:     time.Now().Add(rate),
		rate:      rate,
		updated:   time.Now(),
	}
	s.buckets[name] = b
	return b, nil
}

func (s *Storage) Clean(name string) {
	for name, b := range s.buckets {
		if b.updated.Before(time.Now().Add(-1 * time.Hour)) {
			delete(s.buckets, name)
		}
	}
}
