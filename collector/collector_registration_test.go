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
	"net/http"
	"net/url"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestRegisterNodesAndIndicesDoesNotConflict(t *testing.T) {
	u, err := url.Parse("http://127.0.0.1:9200")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	reg := prometheus.NewRegistry()
	nodes := NewNodes(promslog.NewNopLogger(), http.DefaultClient, u, true, "_local")
	indices := NewIndices(promslog.NewNopLogger(), http.DefaultClient, u, false, false)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("registering nodes and indices collectors should not panic, got: %v", r)
		}
	}()

	reg.MustRegister(nodes)
	reg.MustRegister(indices)
}
