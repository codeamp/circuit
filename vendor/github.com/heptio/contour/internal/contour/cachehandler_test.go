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

package contour

import (
	"reflect"
	"testing"

	ingressroutev1 "github.com/heptio/contour/apis/contour/v1beta1"
	"github.com/heptio/contour/internal/dag"
	"github.com/heptio/contour/internal/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressRouteMetrics(t *testing.T) {
	// ir1 is a valid ingressroute
	ir1 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: 8080,
				}},
			}, {
				Match: "/prefix",
				Delegate: &ingressroutev1.Delegate{
					Name: "delegated",
				}},
			},
		},
	}

	// ir2 is invalid because it contains a service with negative port
	ir2 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: -80,
				}},
			}},
		},
	}

	// ir3 is invalid because it lives outside the roots namespace
	ir3 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "finance",
			Name:      "example",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foobar",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// ir4 is invalid because its match prefix does not match its parent's (ir1)
	ir4 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "delegated",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/doesnotmatch",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// ir5 is invalid because its service weight is less than zero
	ir5 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "delegated",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name:   "home",
					Port:   8080,
					Weight: -10,
				}},
			}},
		},
	}

	// ir6 is invalid because it delegates to itself, producing a cycle
	ir6 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "self",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "self",
				},
			}},
		},
	}

	// ir7 delegates to ir8, which is invalid because it delegates back to ir7
	ir7 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "child",
				},
			}},
		},
	}

	ir8 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "child",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "parent",
				},
			}},
		},
	}

	// ir9 is invalid because it has a route that both delegates and has a list of services
	ir9 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "child",
				},
				Services: []ingressroutev1.Service{{
					Name: "kuard",
					Port: 8080,
				}},
			}},
		},
	}

	// ir10 delegates to ir11 and ir 12.
	ir10 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "validChild",
				},
			}, {
				Match: "/bar",
				Delegate: &ingressroutev1.Delegate{
					Name: "invalidChild",
				},
			}},
		},
	}

	ir11 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "validChild",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "foo",
					Port: 8080,
				}},
			}},
		},
	}

	// ir12 is invalid because it contains an invalid port
	ir12 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "invalidChild",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/bar",
				Services: []ingressroutev1.Service{{
					Name: "foo",
					Port: 12345678,
				}},
			}},
		},
	}

	// ir13 is invalid because it does not specify and FQDN
	ir13 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "foo",
					Port: 8080,
				}},
			}},
		},
	}

	// ir14 delegates tp ir15 but it is invalid because it is missing fqdn
	ir14 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "invalidParent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &ingressroutev1.VirtualHost{},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "validChild",
				},
			}},
		},
	}

	tests := map[string]struct {
		objs           []*ingressroutev1.IngressRoute
		want           metrics.IngressRouteMetric
		rootNamespaces []string
	}{
		"valid ingressroute": {
			objs: []*ingressroutev1.IngressRoute{ir1},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{},
				Valid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
			},
		},
		"invalid port in service": {
			objs: []*ingressroutev1.IngressRoute{ir2},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Valid:    map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
			},
		},
		"root ingressroute outside of roots namespace": {
			objs: []*ingressroutev1.IngressRoute{ir3},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "finance"}: 1,
				},
				Valid:    map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "finance"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "finance"}: 1,
				},
			},
			rootNamespaces: []string{"foo"},
		},
		"delegated route's match prefix does not match parent's prefix": {
			objs: []*ingressroutev1.IngressRoute{ir1, ir4},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Valid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 2,
				},
			},
		},
		"invalid weight in service": {
			objs: []*ingressroutev1.IngressRoute{ir5},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Valid:    map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
			},
		},
		"root ingressroute does not specify FQDN": {
			objs: []*ingressroutev1.IngressRoute{ir13},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Valid:    map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
			},
		},
		"self-edge produces a cycle": {
			objs: []*ingressroutev1.IngressRoute{ir6},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Valid:    map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
			},
		},
		"child delegates to parent, producing a cycle": {
			objs: []*ingressroutev1.IngressRoute{ir7, ir8},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Valid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 2,
				},
			},
		},
		"route has a list of services and also delegates": {
			objs: []*ingressroutev1.IngressRoute{ir9},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Valid:    map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
			},
		},
		"ingressroute is an orphaned route": {
			objs: []*ingressroutev1.IngressRoute{ir8},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{},
				Valid:   map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Root: map[metrics.Meta]int{},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
			},
		},
		"ingressroute delegates to multiple ingressroutes, one is invalid": {
			objs: []*ingressroutev1.IngressRoute{ir10, ir11, ir12},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 1,
				},
				Valid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 2,
				},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 3,
				},
			},
		},
		"invalid parent orphans children": {
			objs: []*ingressroutev1.IngressRoute{ir14, ir11},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Valid: map[metrics.Meta]int{},
				Orphaned: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 2,
				},
			},
		},
		"multi-parent children is not orphaned when one of the parents is invalid": {
			objs: []*ingressroutev1.IngressRoute{ir14, ir11, ir10},
			want: metrics.IngressRouteMetric{
				Invalid: map[metrics.Meta]int{
					{Namespace: "roots"}: 1,
				},
				Valid: map[metrics.Meta]int{
					{Namespace: "roots", VHost: "example.com"}: 2,
				},
				Orphaned: map[metrics.Meta]int{},
				Root: map[metrics.Meta]int{
					{Namespace: "roots"}: 2,
				},
				Total: map[metrics.Meta]int{
					{Namespace: "roots"}: 3,
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b := dag.Builder{
				KubernetesCache: dag.KubernetesCache{
					IngressRouteRootNamespaces: tc.rootNamespaces,
				},
			}
			for _, o := range tc.objs {
				b.Insert(o)
			}
			dag := b.Build()
			gotMetrics := calculateIngressRouteMetric(dag)
			if !reflect.DeepEqual(tc.want.Root, gotMetrics.Root) {
				t.Fatalf("(metrics-Root) expected to find: %v but got: %v", tc.want.Root, gotMetrics.Root)
			}
			if !reflect.DeepEqual(tc.want.Valid, gotMetrics.Valid) {
				t.Fatalf("(metrics-Valid) expected to find: %v but got: %v", tc.want.Valid, gotMetrics.Valid)
			}
			if !reflect.DeepEqual(tc.want.Invalid, gotMetrics.Invalid) {
				t.Fatalf("(metrics-Invalid) expected to find: %v but got: %v", tc.want.Invalid, gotMetrics.Invalid)
			}
			if !reflect.DeepEqual(tc.want.Orphaned, gotMetrics.Orphaned) {
				t.Fatalf("(metrics-Orphaned) expected to find: %v but got: %v", tc.want.Orphaned, gotMetrics.Orphaned)
			}
			if !reflect.DeepEqual(tc.want.Total, gotMetrics.Total) {
				t.Fatalf("(metrics-Total) expected to find: %v but got: %v", tc.want.Total, gotMetrics.Total)
			}
		})
	}
}
