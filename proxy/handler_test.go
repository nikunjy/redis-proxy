package proxy

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/nikunjy/redis-proxy/store"
	"github.com/stretchr/testify/require"
)

func requireKey(t *testing.T, proxy *proxyHandler, key string, expected string) {
	val, err := proxy.get(key)
	require.NoError(t, err)
	require.EqualValues(t, expected, val)
}

func requireCacheKey(t *testing.T, proxy *proxyHandler, key string, expected string) {
	val, err := proxy.cachedGet(key)
	require.NoError(t, err)
	require.EqualValues(t, expected, val)
}

func TestLRUEviction(t *testing.T) {
	counts := make(map[string]int)
	getCallback := func(key string) error {
		counts[key]++
		return nil
	}
	local := store.NewLocal()
	local.WithGetCallback(getCallback)
	proxy, err := New(local, WithCacheSize(2))
	require.NoError(t, err)
	require.NoError(t, proxy.put("a", "1"))
	require.NoError(t, proxy.put("b", "2"))
	require.NoError(t, proxy.put("c", "3"))

	// the regular gets will force the store lookups even though
	// put has stored them inside the cache
	requireKey(t, proxy, "a", "1")
	requireKey(t, proxy, "b", "2")
	requireCacheKey(t, proxy, "a", "1")
	requireCacheKey(t, proxy, "b", "2")

	requireCounts(t, map[string]int{
		"a": 1,
		"b": 1,
	}, counts)
	requireCacheKey(t, proxy, "a", "1")

	// b will be knocked out
	requireKey(t, proxy, "c", "3")

	// b will have to be pulled from the actual store again
	requireCacheKey(t, proxy, "b", "2")
	requireCounts(t, map[string]int{
		"a": 1,
		"c": 1,
		"b": 2,
	}, counts)
}

func TestCacheExpiry(t *testing.T) {
	counts := make(map[string]int)
	getCallback := func(key string) error {
		counts[key]++
		return nil
	}
	local := store.NewLocal()
	local.WithGetCallback(getCallback)
	proxy, err := New(local, WithCacheSize(2), WithCacheTTL(time.Second*2))
	fakeClock := clockwork.NewFakeClock()
	proxy.clock = fakeClock

	require.NoError(t, err)
	require.NoError(t, proxy.put("a", "1"))

	fakeClock.Advance(time.Second * 2)
	requireCacheKey(t, proxy, "a", "1")
	// put puts the thing in the cache
	requireCounts(t, nil, counts)

	// cache will be expired
	fakeClock.Advance(time.Second * 2)
	requireCacheKey(t, proxy, "a", "1")
	requireCounts(t, map[string]int{"a": 1}, counts)
}

func requireCounts(t *testing.T, expected map[string]int, actual map[string]int) {
	for k := range expected {
		require.EqualValues(t, expected[k], actual[k])
	}
	require.Equal(t, len(expected), len(actual), "Len of expected and actual different")
}
