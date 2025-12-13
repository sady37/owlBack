package aggregator_test

import (
	"context"
	"sync"
	"time"

	agg "wisefido-card-aggregator/internal/aggregator"
)

// fakeKVStore 仅用于单元测试（内存 KV + TTL）
type fakeKVStore struct {
	mu   sync.Mutex
	data map[string]fakeKVItem
}

type fakeKVItem struct {
	value   string
	expires time.Time // zero = no ttl
}

func newFakeKVStore() *fakeKVStore {
	return &fakeKVStore{
		data: make(map[string]fakeKVItem),
	}
}

func (f *fakeKVStore) Get(ctx context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	item, ok := f.data[key]
	if !ok {
		return "", agg.ErrCacheMiss
	}
	if !item.expires.IsZero() && time.Now().After(item.expires) {
		delete(f.data, key)
		return "", agg.ErrCacheMiss
	}
	return item.value, nil
}

func (f *fakeKVStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	f.data[key] = fakeKVItem{value: value, expires: exp}
	return nil
}


