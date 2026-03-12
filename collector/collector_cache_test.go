// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type noOpCollector struct {
	id int
}

func (n noOpCollector) Update(context.Context, chan<- prometheus.Metric) error {
	return nil
}

func TestNewElasticsearchCollector_UsesSharedCacheByDefault(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	u, err := url.Parse("http://example:9200")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	originalFactories := factories
	originalCollectorState := collectorState
	originalInitiatedCollectors := initiatedCollectors
	originalForcedCollectors := forcedCollectors
	t.Cleanup(func() {
		factories = originalFactories
		collectorState = originalCollectorState
		initiatedCollectors = originalInitiatedCollectors
		forcedCollectors = originalForcedCollectors
	})

	enabled := true
	factoryCalls := 0
	factories = map[string]factoryFunc{
		"test": func(_ *slog.Logger, _ *url.URL, _ *http.Client) (Collector, error) {
			factoryCalls++
			return &noOpCollector{id: factoryCalls}, nil
		},
	}
	collectorState = map[string]*bool{"test": &enabled}
	initiatedCollectors = make(map[string]Collector)
	forcedCollectors = map[string]bool{}

	first, err := NewElasticsearchCollector(logger, nil, WithElasticsearchURL(u), WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatalf("unexpected error creating first collector: %v", err)
	}
	second, err := NewElasticsearchCollector(logger, nil, WithElasticsearchURL(u), WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatalf("unexpected error creating second collector: %v", err)
	}

	if factoryCalls != 1 {
		t.Fatalf("expected factory to be called once with cache enabled, got %d", factoryCalls)
	}
	if first.Collectors["test"] != second.Collectors["test"] {
		t.Fatalf("expected collector instance to be reused from cache")
	}
}

func TestNewElasticsearchCollector_SkipsCacheWhenRequested(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	u, err := url.Parse("http://example:9200")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	originalFactories := factories
	originalCollectorState := collectorState
	originalInitiatedCollectors := initiatedCollectors
	originalForcedCollectors := forcedCollectors
	t.Cleanup(func() {
		factories = originalFactories
		collectorState = originalCollectorState
		initiatedCollectors = originalInitiatedCollectors
		forcedCollectors = originalForcedCollectors
	})

	enabled := true
	factoryCalls := 0
	factories = map[string]factoryFunc{
		"test": func(_ *slog.Logger, _ *url.URL, _ *http.Client) (Collector, error) {
			factoryCalls++
			return &noOpCollector{id: factoryCalls}, nil
		},
	}
	collectorState = map[string]*bool{"test": &enabled}
	initiatedCollectors = make(map[string]Collector)
	forcedCollectors = map[string]bool{}

	first, err := NewElasticsearchCollector(logger, nil, WithElasticsearchURL(u), WithHTTPClient(http.DefaultClient), WithSkipCache(true))
	if err != nil {
		t.Fatalf("unexpected error creating first collector: %v", err)
	}
	second, err := NewElasticsearchCollector(logger, nil, WithElasticsearchURL(u), WithHTTPClient(http.DefaultClient), WithSkipCache(true))
	if err != nil {
		t.Fatalf("unexpected error creating second collector: %v", err)
	}

	if factoryCalls != 2 {
		t.Fatalf("expected factory to be called twice with cache disabled, got %d", factoryCalls)
	}
	if first.Collectors["test"] == second.Collectors["test"] {
		t.Fatalf("expected distinct collector instances when cache is disabled")
	}
	if len(initiatedCollectors) != 0 {
		t.Fatalf("expected no collectors to be stored in shared cache when cache is disabled")
	}
}
