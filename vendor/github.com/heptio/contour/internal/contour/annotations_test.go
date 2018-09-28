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
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestParseAnnotationTimeout(t *testing.T) {
	tests := map[string]struct {
		a    map[string]string
		want time.Duration
		ok   bool
	}{
		"nada": {
			a:    nil,
			want: 0,
			ok:   false,
		},
		"empty": {
			a:    map[string]string{annotationRequestTimeout: ""}, // not even sure this is possible via the API
			want: 0,
			ok:   false,
		},
		"infinity": {
			a:    map[string]string{annotationRequestTimeout: "infinity"},
			want: 0,
			ok:   true,
		},
		"10 seconds": {
			a:    map[string]string{annotationRequestTimeout: "10s"},
			want: 10 * time.Second,
			ok:   true,
		},
		"invalid": {
			a:    map[string]string{annotationRequestTimeout: "10"}, // 10 what?
			want: 0,
			ok:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := parseAnnotationTimeout(tc.a, annotationRequestTimeout)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("parseAnnotationTimeout(%q): want: %v, %v, got: %v, %v", tc.a, tc.want, tc.ok, got, ok)
			}
		})
	}
}

func TestParseAnnotationUInt32(t *testing.T) {
	tests := map[string]struct {
		a     map[string]string
		want  uint32
		isNil bool
	}{
		"nada": {
			a:     nil,
			isNil: true,
		},
		"empty": {
			a:     map[string]string{annotationRequestTimeout: ""}, // not even sure this is possible via the API
			isNil: true,
		},
		"smallest": {
			a:    map[string]string{annotationRequestTimeout: "0"},
			want: 0,
		},
		"middle value": {
			a:    map[string]string{annotationRequestTimeout: "20"},
			want: 20,
		},
		"biggest": {
			a:    map[string]string{annotationRequestTimeout: "4294967295"},
			want: math.MaxUint32,
		},
		"invalid": {
			a:     map[string]string{annotationRequestTimeout: "10seconds"}, // not a duration
			isNil: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseAnnotationUInt32(tc.a, annotationRequestTimeout)

			if ((got == nil) != tc.isNil) || (got != nil && got.Value != tc.want) {
				t.Fatalf("parseAnnotationUInt32(%q): want: %v, isNil: %v, got: %v", tc.a, tc.want, tc.isNil, got)
			}
		})
	}
}

func TestWebsocketRoutes(t *testing.T) {
	tests := map[string]struct {
		a    *v1beta1.Ingress
		want map[string]*types.BoolValue
	}{
		"empty": {
			a: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{annotationWebsocketRoutes: ""},
				},
			},
			want: map[string]*types.BoolValue{},
		},
		"empty with spaces": {
			a: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{annotationWebsocketRoutes: ", ,"},
				},
			},
			want: map[string]*types.BoolValue{},
		},
		"single value": {
			a: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{annotationWebsocketRoutes: "/ws1"},
				},
			},
			want: map[string]*types.BoolValue{
				"/ws1": {Value: true},
			},
		},
		"multiple values": {
			a: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{annotationWebsocketRoutes: "/ws1,/ws2"},
				},
			},
			want: map[string]*types.BoolValue{
				"/ws1": {Value: true},
				"/ws2": {Value: true},
			},
		},
		"multiple values with spaces and invalid entries": {
			a: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{annotationWebsocketRoutes: " /ws1, , /ws2 "},
				},
			},
			want: map[string]*types.BoolValue{
				"/ws1": {Value: true},
				"/ws2": {Value: true},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := websocketRoutes(tc.a)
			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("websocketRoutes(%q): want: %v, got: %v", tc.a, tc.want, got)
			}
		})
	}
}

func TestHttpAllowed(t *testing.T) {
	tests := map[string]struct {
		i     *v1beta1.Ingress
		valid bool
	}{
		"basic ingress": {
			i: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "simple",
					Namespace: "default",
				},
				Spec: v1beta1.IngressSpec{
					TLS: []v1beta1.IngressTLS{{
						Hosts:      []string{"whatever.example.com"},
						SecretName: "secret",
					}},
					Backend: backend("backend", intstr.FromInt(80)),
				},
			},
			valid: true,
		},
		"kubernetes.io/ingress.allow-http: \"false\"": {
			i: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "simple",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.allow-http": "false",
					},
				},
				Spec: v1beta1.IngressSpec{
					TLS: []v1beta1.IngressTLS{{
						Hosts:      []string{"whatever.example.com"},
						SecretName: "secret",
					}},
					Backend: backend("backend", intstr.FromInt(80)),
				},
			},
			valid: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := httpAllowed(tc.i)
			want := tc.valid
			if got != want {
				t.Fatalf("got: %v, want: %v", got, want)
			}
		})
	}
}

func backend(name string, port intstr.IntOrString) *v1beta1.IngressBackend {
	return &v1beta1.IngressBackend{
		ServiceName: name,
		ServicePort: port,
	}
}
