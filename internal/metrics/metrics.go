package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type providerCounters struct {
	deliveryTotal  atomic.Uint64
	deliveryFailed atomic.Uint64
	latencySumMS   atomic.Uint64
	latencySamples atomic.Uint64
}

var countersByProvider sync.Map

func Record(provider string, status string, latencyMS int64) {
	provider = normalizeProvider(provider)
	counters := getOrCreate(provider)

	counters.deliveryTotal.Add(1)
	if status == "failed" {
		counters.deliveryFailed.Add(1)
	}

	if latencyMS < 0 {
		latencyMS = 0
	}

	counters.latencySumMS.Add(uint64(latencyMS))
	counters.latencySamples.Add(1)
}

func PrometheusText() string {
	providers := snapshotProviders()
	b := strings.Builder{}

	b.WriteString("# HELP delivery_total Total number of delivery attempts.\n")
	b.WriteString("# TYPE delivery_total counter\n")
	for _, provider := range providers {
		counters := getOrCreate(provider)
		b.WriteString(fmt.Sprintf("delivery_total{provider=%q} %d\n", provider, counters.deliveryTotal.Load()))
	}

	b.WriteString("# HELP delivery_failed_total Total number of failed delivery attempts.\n")
	b.WriteString("# TYPE delivery_failed_total counter\n")
	for _, provider := range providers {
		counters := getOrCreate(provider)
		b.WriteString(fmt.Sprintf("delivery_failed_total{provider=%q} %d\n", provider, counters.deliveryFailed.Load()))
	}

	b.WriteString("# HELP provider_latency_ms Delivery latency in milliseconds by provider.\n")
	b.WriteString("# TYPE provider_latency_ms summary\n")
	for _, provider := range providers {
		counters := getOrCreate(provider)
		b.WriteString(fmt.Sprintf("provider_latency_ms_sum{provider=%q} %d\n", provider, counters.latencySumMS.Load()))
		b.WriteString(fmt.Sprintf("provider_latency_ms_count{provider=%q} %d\n", provider, counters.latencySamples.Load()))
	}

	return b.String()
}

func getOrCreate(provider string) *providerCounters {
	if current, ok := countersByProvider.Load(provider); ok {
		if typed, ok := current.(*providerCounters); ok {
			return typed
		}
	}

	candidate := &providerCounters{}
	stored, _ := countersByProvider.LoadOrStore(provider, candidate)
	typed, ok := stored.(*providerCounters)
	if ok {
		return typed
	}

	return candidate
}

func normalizeProvider(provider string) string {
	trimmed := strings.TrimSpace(strings.ToLower(provider))
	if trimmed == "" {
		return "unknown"
	}

	return trimmed
}

func snapshotProviders() []string {
	providers := make([]string, 0)
	countersByProvider.Range(func(key any, value any) bool {
		provider, ok := key.(string)
		if ok {
			providers = append(providers, provider)
		}
		return true
	})

	if len(providers) == 0 {
		providers = append(providers, "unknown")
	}

	sort.Strings(providers)
	return providers
}
