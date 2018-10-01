// Copyright © 2018 Heptio
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
)

type testMetric struct {
	metric string
	want   []*io_prometheus_client.Metric
}

func TestWriteDAGTimestampMetric(t *testing.T) {
	tests := map[string]struct {
		timestampMetric testMetric
		value           int64
	}{
		"simple": {
			value: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC).Unix(),
			timestampMetric: testMetric{
				metric: IngressRouteDAGRebuildGauge,
				want: []*io_prometheus_client.Metric{
					{
						Gauge: &io_prometheus_client.Gauge{
							Value: func() *float64 { i := float64(1.258490098e+09); return &i }(),
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := prometheus.NewRegistry()
			m := NewMetrics(r)
			m.SetDAGRebuiltMetric(tc.value)

			gatherers := prometheus.Gatherers{
				r,
				prometheus.DefaultGatherer,
			}

			gathering, err := gatherers.Gather()
			if err != nil {
				t.Fatal(err)
			}

			gotTimestamp := []*io_prometheus_client.Metric{}
			for _, mf := range gathering {
				if mf.GetName() == tc.timestampMetric.metric {
					gotTimestamp = mf.Metric
				}
			}

			if !reflect.DeepEqual(gotTimestamp, tc.timestampMetric.want) {
				t.Fatalf("write metric timestamp metric failed, want: %v got: %v", tc.timestampMetric.want, gotTimestamp)
			}
		})
	}
}

func TestWriteIngressRouteMetric(t *testing.T) {
	tests := map[string]struct {
		irMetrics IngressRouteMetric
		total     testMetric
		valid     testMetric
		invalid   testMetric
		orphaned  testMetric
		root      testMetric
	}{
		"simple": {
			irMetrics: IngressRouteMetric{
				Total: map[Meta]int{
					{Namespace: "testns"}: 6,
					{Namespace: "foons"}:  3,
				},
				Valid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 3,
				},
				Invalid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 2,
				},
				Orphaned: map[Meta]int{
					{Namespace: "testns"}: 1,
				},
				Root: map[Meta]int{
					{Namespace: "testns"}: 4,
				},
			},
			total: testMetric{
				metric: IngressRouteTotalGauge,
				want: []*io_prometheus_client.Metric{
					{
						Label: []*io_prometheus_client.LabelPair{{
							Name:  func() *string { i := "namespace"; return &i }(),
							Value: func() *string { i := "foons"; return &i }(),
						}},
						Gauge: &io_prometheus_client.Gauge{
							Value: func() *float64 { i := float64(3); return &i }(),
						},
					},
					{
						Label: []*io_prometheus_client.LabelPair{{
							Name:  func() *string { i := "namespace"; return &i }(),
							Value: func() *string { i := "testns"; return &i }(),
						}},
						Gauge: &io_prometheus_client.Gauge{
							Value: func() *float64 { i := float64(6); return &i }(),
						},
					},
				},
			},
			orphaned: testMetric{
				metric: IngressRouteOrphanedGauge,
				want: []*io_prometheus_client.Metric{
					{
						Label: []*io_prometheus_client.LabelPair{{
							Name:  func() *string { i := "namespace"; return &i }(),
							Value: func() *string { i := "testns"; return &i }(),
						}},
						Gauge: &io_prometheus_client.Gauge{
							Value: func() *float64 { i := float64(1); return &i }(),
						},
					},
				},
			},
			valid: testMetric{
				metric: IngressRouteValidGauge,
				want: []*io_prometheus_client.Metric{
					{
						Label: []*io_prometheus_client.LabelPair{{
							Name:  func() *string { i := "namespace"; return &i }(),
							Value: func() *string { i := "testns"; return &i }(),
						}, {
							Name:  func() *string { i := "vhost"; return &i }(),
							Value: func() *string { i := "foo.com"; return &i }(),
						}},
						Gauge: &io_prometheus_client.Gauge{
							Value: func() *float64 { i := float64(3); return &i }(),
						},
					},
				},
			},
			invalid: testMetric{
				metric: IngressRouteInvalidGauge,
				want: []*io_prometheus_client.Metric{
					{
						Label: []*io_prometheus_client.LabelPair{{
							Name:  func() *string { i := "namespace"; return &i }(),
							Value: func() *string { i := "testns"; return &i }(),
						}, {
							Name:  func() *string { i := "vhost"; return &i }(),
							Value: func() *string { i := "foo.com"; return &i }(),
						}},
						Gauge: &io_prometheus_client.Gauge{
							Value: func() *float64 { i := float64(2); return &i }(),
						},
					},
				},
			},
			root: testMetric{
				metric: IngressRouteRootTotalGauge,
				want: []*io_prometheus_client.Metric{
					{
						Label: []*io_prometheus_client.LabelPair{{
							Name:  func() *string { i := "namespace"; return &i }(),
							Value: func() *string { i := "testns"; return &i }(),
						}},
						Gauge: &io_prometheus_client.Gauge{
							Value: func() *float64 { i := float64(4); return &i }(),
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := prometheus.NewRegistry()
			m := NewMetrics(r)
			m.SetIngressRouteMetric(tc.irMetrics)

			gatherers := prometheus.Gatherers{
				r,
				prometheus.DefaultGatherer,
			}

			gathering, err := gatherers.Gather()
			if err != nil {
				t.Fatal(err)
			}

			gotTotal := []*io_prometheus_client.Metric{}
			gotValid := []*io_prometheus_client.Metric{}
			gotInvalid := []*io_prometheus_client.Metric{}
			gotOrphaned := []*io_prometheus_client.Metric{}
			gotRoot := []*io_prometheus_client.Metric{}
			for _, mf := range gathering {
				if mf.GetName() == tc.total.metric {
					gotTotal = mf.Metric
				} else if mf.GetName() == tc.valid.metric {
					gotValid = mf.Metric
				} else if mf.GetName() == tc.invalid.metric {
					gotInvalid = mf.Metric
				} else if mf.GetName() == tc.orphaned.metric {
					gotOrphaned = mf.Metric
				} else if mf.GetName() == tc.root.metric {
					gotRoot = mf.Metric
				}
			}

			if !reflect.DeepEqual(gotTotal, tc.total.want) {
				t.Fatalf("write metric total metric failed, want: %v got: %v", tc.total.want, gotTotal)
			}
			if !reflect.DeepEqual(gotValid, tc.valid.want) {
				t.Fatalf("write metric valid metric failed, want: %v got: %v", tc.valid.want, gotValid)
			}
			if !reflect.DeepEqual(gotInvalid, tc.invalid.want) {
				t.Fatalf("write metric invalid metric failed, want: %v got: %v", tc.invalid.want, gotInvalid)
			}
			if !reflect.DeepEqual(gotOrphaned, tc.orphaned.want) {
				t.Fatalf("write metric orphaned metric failed, want: %v got: %v", tc.orphaned.want, gotOrphaned)
			}
			if !reflect.DeepEqual(gotRoot, tc.root.want) {
				t.Fatalf("write metric orphaned metric failed, want: %v got: %v", tc.root.want, gotRoot)
			}
		})
	}
}
func TestRemoveMetric(t *testing.T) {

	total := testMetric{
		metric: IngressRouteTotalGauge,
		want: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{{
					Name:  func() *string { i := "namespace"; return &i }(),
					Value: func() *string { i := "foons"; return &i }(),
				}},
				Gauge: &io_prometheus_client.Gauge{
					Value: func() *float64 { i := float64(3); return &i }(),
				},
			},
			{
				Label: []*io_prometheus_client.LabelPair{{
					Name:  func() *string { i := "namespace"; return &i }(),
					Value: func() *string { i := "testns"; return &i }(),
				}},
				Gauge: &io_prometheus_client.Gauge{
					Value: func() *float64 { i := float64(6); return &i }(),
				},
			},
		},
	}

	orphaned := testMetric{
		metric: IngressRouteOrphanedGauge,
		want: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{{
					Name:  func() *string { i := "namespace"; return &i }(),
					Value: func() *string { i := "testns"; return &i }(),
				}},
				Gauge: &io_prometheus_client.Gauge{
					Value: func() *float64 { i := float64(1); return &i }(),
				},
			},
		},
	}

	valid := testMetric{
		metric: IngressRouteValidGauge,
		want: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{{
					Name:  func() *string { i := "namespace"; return &i }(),
					Value: func() *string { i := "testns"; return &i }(),
				}, {
					Name:  func() *string { i := "vhost"; return &i }(),
					Value: func() *string { i := "foo.com"; return &i }(),
				}},
				Gauge: &io_prometheus_client.Gauge{
					Value: func() *float64 { i := float64(3); return &i }(),
				},
			},
		},
	}

	invalid := testMetric{
		metric: IngressRouteInvalidGauge,
		want: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{{
					Name:  func() *string { i := "namespace"; return &i }(),
					Value: func() *string { i := "testns"; return &i }(),
				}, {
					Name:  func() *string { i := "vhost"; return &i }(),
					Value: func() *string { i := "foo.com"; return &i }(),
				}},
				Gauge: &io_prometheus_client.Gauge{
					Value: func() *float64 { i := float64(2); return &i }(),
				},
			},
		},
	}

	root := testMetric{
		metric: IngressRouteRootTotalGauge,
		want: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{{
					Name:  func() *string { i := "namespace"; return &i }(),
					Value: func() *string { i := "testns"; return &i }(),
				}},
				Gauge: &io_prometheus_client.Gauge{
					Value: func() *float64 { i := float64(4); return &i }(),
				},
			},
		},
	}

	tests := map[string]struct {
		irMetrics        IngressRouteMetric
		irMetricsUpdated IngressRouteMetric
		totalWant        []*io_prometheus_client.Metric
		validWant        []*io_prometheus_client.Metric
		invalidWant      []*io_prometheus_client.Metric
		orphanedWant     []*io_prometheus_client.Metric
		rootWant         []*io_prometheus_client.Metric
	}{
		"orphan is resolved": {
			irMetrics: IngressRouteMetric{
				Total: map[Meta]int{
					{Namespace: "testns"}: 6,
					{Namespace: "foons"}:  3,
				},
				Valid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 3,
				},
				Invalid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 2,
				},
				Orphaned: map[Meta]int{
					{Namespace: "testns"}: 1,
				},
				Root: map[Meta]int{
					{Namespace: "testns"}: 4,
				},
			},
			irMetricsUpdated: IngressRouteMetric{
				Total: map[Meta]int{
					{Namespace: "testns"}: 6,
					{Namespace: "foons"}:  3,
				},
				Valid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 3,
				},
				Invalid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 2,
				},
				Orphaned: map[Meta]int{},
				Root: map[Meta]int{
					{Namespace: "testns"}: 4,
				},
			},
			totalWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "foons"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(3); return &i }(),
					},
				},
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(6); return &i }(),
					},
				},
			},
			orphanedWant: []*io_prometheus_client.Metric{},
			validWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}, {
						Name:  func() *string { i := "vhost"; return &i }(),
						Value: func() *string { i := "foo.com"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(3); return &i }(),
					},
				},
			},
			invalidWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}, {
						Name:  func() *string { i := "vhost"; return &i }(),
						Value: func() *string { i := "foo.com"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(2); return &i }(),
					},
				},
			},
			rootWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(4); return &i }(),
					},
				},
			},
		},
		"root IngressRoute is deleted": {
			irMetrics: IngressRouteMetric{
				Total: map[Meta]int{
					{Namespace: "testns"}: 6,
					{Namespace: "foons"}:  3,
				},
				Valid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 3,
				},
				Invalid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 2,
				},
				Orphaned: map[Meta]int{
					{Namespace: "testns"}: 1,
				},
				Root: map[Meta]int{
					{Namespace: "testns"}: 4,
				},
			},
			irMetricsUpdated: IngressRouteMetric{
				Total: map[Meta]int{
					{Namespace: "testns"}: 6,
				},
				Valid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 3,
				},
				Invalid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 2,
				},
				Orphaned: map[Meta]int{},
				Root: map[Meta]int{
					{Namespace: "testns"}: 4,
				},
			},
			totalWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(6); return &i }(),
					},
				},
			},
			orphanedWant: []*io_prometheus_client.Metric{},
			validWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}, {
						Name:  func() *string { i := "vhost"; return &i }(),
						Value: func() *string { i := "foo.com"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(3); return &i }(),
					},
				},
			},
			invalidWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}, {
						Name:  func() *string { i := "vhost"; return &i }(),
						Value: func() *string { i := "foo.com"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(2); return &i }(),
					},
				},
			},
			rootWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(4); return &i }(),
					},
				},
			},
		},
		"valid is deleted from namespace": {
			irMetrics: IngressRouteMetric{
				Total: map[Meta]int{
					{Namespace: "testns"}: 6,
					{Namespace: "foons"}:  3,
				},
				Valid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 3,
				},
				Invalid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 2,
				},
				Orphaned: map[Meta]int{
					{Namespace: "testns"}: 1,
				},
				Root: map[Meta]int{
					{Namespace: "testns"}: 4,
				},
			},
			irMetricsUpdated: IngressRouteMetric{
				Total: map[Meta]int{
					{Namespace: "testns"}: 6,
				},
				Valid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 3,
				},
				Invalid: map[Meta]int{
					{Namespace: "testns", VHost: "foo.com"}: 2,
				},
				Orphaned: map[Meta]int{},
				Root: map[Meta]int{
					{Namespace: "testns"}: 4,
				},
			},
			totalWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(6); return &i }(),
					},
				},
			},
			orphanedWant: []*io_prometheus_client.Metric{},
			validWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}, {
						Name:  func() *string { i := "vhost"; return &i }(),
						Value: func() *string { i := "foo.com"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(3); return &i }(),
					},
				},
			},
			invalidWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}, {
						Name:  func() *string { i := "vhost"; return &i }(),
						Value: func() *string { i := "foo.com"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(2); return &i }(),
					},
				},
			},
			rootWant: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{{
						Name:  func() *string { i := "namespace"; return &i }(),
						Value: func() *string { i := "testns"; return &i }(),
					}},
					Gauge: &io_prometheus_client.Gauge{
						Value: func() *float64 { i := float64(4); return &i }(),
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := prometheus.NewRegistry()
			m := NewMetrics(r)
			m.SetIngressRouteMetric(tc.irMetrics)

			gatherers := prometheus.Gatherers{
				r,
				prometheus.DefaultGatherer,
			}

			gathering, err := gatherers.Gather()
			if err != nil {
				t.Fatal(err)
			}

			gotTotal := []*io_prometheus_client.Metric{}
			gotValid := []*io_prometheus_client.Metric{}
			gotInvalid := []*io_prometheus_client.Metric{}
			gotOrphaned := []*io_prometheus_client.Metric{}
			gotRoot := []*io_prometheus_client.Metric{}
			for _, mf := range gathering {
				if mf.GetName() == total.metric {
					gotTotal = mf.Metric
				} else if mf.GetName() == valid.metric {
					gotValid = mf.Metric
				} else if mf.GetName() == invalid.metric {
					gotInvalid = mf.Metric
				} else if mf.GetName() == orphaned.metric {
					gotOrphaned = mf.Metric
				} else if mf.GetName() == root.metric {
					gotRoot = mf.Metric
				}
			}

			if !reflect.DeepEqual(gotTotal, total.want) {
				t.Fatalf("write metric total metric failed, want: %v got: %v", total.want, gotTotal)
			}
			if !reflect.DeepEqual(gotValid, valid.want) {
				t.Fatalf("write metric valid metric failed, want: %v got: %v", valid.want, gotValid)
			}
			if !reflect.DeepEqual(gotInvalid, invalid.want) {
				t.Fatalf("write metric invalid metric failed, want: %v got: %v", invalid.want, gotInvalid)
			}
			if !reflect.DeepEqual(gotOrphaned, orphaned.want) {
				t.Fatalf("write metric orphaned metric failed, want: %v got: %v", orphaned.want, gotOrphaned)
			}
			if !reflect.DeepEqual(gotRoot, root.want) {
				t.Fatalf("write metric orphaned metric failed, want: %v got: %v", root.want, gotRoot)
			}

			m.SetIngressRouteMetric(tc.irMetricsUpdated)

			// Now validate that metrics got removed
			gatherers = prometheus.Gatherers{
				r,
				prometheus.DefaultGatherer,
			}

			gathering, err = gatherers.Gather()
			if err != nil {
				t.Fatal(err)
			}

			gotTotal = []*io_prometheus_client.Metric{}
			gotValid = []*io_prometheus_client.Metric{}
			gotInvalid = []*io_prometheus_client.Metric{}
			gotOrphaned = []*io_prometheus_client.Metric{}
			gotRoot = []*io_prometheus_client.Metric{}
			for _, mf := range gathering {
				if mf.GetName() == total.metric {
					gotTotal = mf.Metric
				} else if mf.GetName() == valid.metric {
					gotValid = mf.Metric
				} else if mf.GetName() == invalid.metric {
					gotInvalid = mf.Metric
				} else if mf.GetName() == orphaned.metric {
					gotOrphaned = mf.Metric
				} else if mf.GetName() == root.metric {
					gotRoot = mf.Metric
				}
			}

			if !reflect.DeepEqual(gotTotal, tc.totalWant) {
				t.Fatalf("write metric total metric failed, want: %v got: %v", tc.totalWant, gotTotal)
			}
			if !reflect.DeepEqual(gotValid, tc.validWant) {
				t.Fatalf("write metric valid metric failed, want: %v got: %v", tc.validWant, gotValid)
			}
			if !reflect.DeepEqual(gotInvalid, tc.invalidWant) {
				t.Fatalf("write metric invalid metric failed, want: %v got: %v", tc.invalidWant, gotInvalid)
			}
			if !reflect.DeepEqual(gotOrphaned, tc.orphanedWant) {
				t.Fatalf("write metric orphaned metric failed, want: %v got: %v", tc.orphanedWant, gotOrphaned)
			}
			if !reflect.DeepEqual(gotRoot, tc.rootWant) {
				t.Fatalf("write metric orphaned metric failed, want: %v got: %v", tc.rootWant, gotRoot)
			}
		})
	}
}
