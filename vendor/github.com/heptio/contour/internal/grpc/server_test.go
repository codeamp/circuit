// Copyright © 2017 Heptio
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

package grpc

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/heptio/contour/internal/contour"
	"github.com/heptio/contour/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGRPCStreaming(t *testing.T) {
	var l net.Listener

	// tr and et is recreated before the start of each test.
	var et *contour.EndpointsTranslator
	var reh *contour.ResourceEventHandler

	newClient := func(t *testing.T) *grpc.ClientConn {
		cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
		check(t, err)
		return cc
	}

	tests := map[string]func(*testing.T){
		"StreamClusters": func(t *testing.T) {
			reh.OnAdd(&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "simple",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						"app": "simple",
					},
					Ports: []v1.ServicePort{{
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.FromInt(6502),
					}},
				},
			})

			cc := newClient(t)
			defer cc.Close()
			sds := v2.NewClusterDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			stream, err := sds.StreamClusters(ctx)
			check(t, err)
			sendreq(t, stream, clusterType) // send initial notification
			checkrecv(t, stream)            // check we receive one notification
			checktimeout(t, stream)         // check that the second receive times out
		},
		"StreamEndpoints": func(t *testing.T) {
			et.OnAdd(&v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-scheduler",
					Namespace: "kube-system",
				},
				Subsets: []v1.EndpointSubset{{
					Addresses: []v1.EndpointAddress{{
						IP: "130.211.139.167",
					}},
					Ports: []v1.EndpointPort{{
						Port: 80,
					}, {
						Port: 443,
					}},
				}},
			})

			cc := newClient(t)
			defer cc.Close()
			eds := v2.NewEndpointDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			stream, err := eds.StreamEndpoints(ctx)
			check(t, err)
			sendreq(t, stream, endpointType) // send initial notification
			checkrecv(t, stream)             // check we receive one notification
			checktimeout(t, stream)          // check that the second receive times out
		},
		"StreamListeners": func(t *testing.T) {
			cc := newClient(t)
			defer cc.Close()
			lds := v2.NewListenerDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			// add an ingress, which will create a non tls listener
			reh.OnAdd(&v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httpbin-org",
					Namespace: "default",
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{{
						Host: "httpbin.org",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{{
									Backend: v1beta1.IngressBackend{
										ServiceName: "httpbin-org",
										ServicePort: intstr.FromInt(80),
									},
								}},
							},
						},
					}},
				},
			})
			stream, err := lds.StreamListeners(ctx)
			check(t, err)
			sendreq(t, stream, listenerType) // send initial notification
			checkrecv(t, stream)             // check we receive one notification
			checktimeout(t, stream)          // check that the second receive times out
		},
		"StreamRoutes": func(t *testing.T) {
			reh.OnAdd(&v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httpbin-org",
					Namespace: "default",
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{{
						Host: "httpbin.org",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{{
									Backend: v1beta1.IngressBackend{
										ServiceName: "httpbin-org",
										ServicePort: intstr.FromInt(80),
									},
								}},
							},
						},
					}},
				},
			})

			cc := newClient(t)
			defer cc.Close()
			rds := v2.NewRouteDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			stream, err := rds.StreamRoutes(ctx)
			check(t, err)
			sendreq(t, stream, routeType) // send initial notification
			checkrecv(t, stream)          // check we receive one notification
			checktimeout(t, stream)       // check that the second receive times out
		},
	}

	log := testLogger(t)
	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			et = &contour.EndpointsTranslator{
				FieldLogger: log,
			}
			ch := contour.CacheHandler{
				Metrics: metrics.NewMetrics(prometheus.NewRegistry()),
			}
			reh = &contour.ResourceEventHandler{
				Notifier: &ch,
				Metrics:  ch.Metrics,
			}
			srv := NewAPI(log, map[string]Cache{
				clusterType:  &ch.ClusterCache,
				routeType:    &ch.RouteCache,
				listenerType: &ch.ListenerCache,
				endpointType: et,
			})
			var err error
			l, err = net.Listen("tcp", "127.0.0.1:0")
			check(t, err)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				srv.Serve(l)
			}()
			defer func() {
				srv.Stop()
				wg.Wait()
				l.Close()
			}()
			fn(t)
		})
	}
}

func TestGRPCFetching(t *testing.T) {
	var l net.Listener

	newClient := func(t *testing.T) *grpc.ClientConn {
		cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
		check(t, err)
		return cc
	}

	tests := map[string]func(*testing.T){
		"FetchClusters": func(t *testing.T) {
			cc := newClient(t)
			defer cc.Close()
			sds := v2.NewClusterDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			req := &v2.DiscoveryRequest{
				TypeUrl: clusterType,
			}
			_, err := sds.FetchClusters(ctx, req)
			check(t, err)
		},
		"FetchEndpoints": func(t *testing.T) {
			cc := newClient(t)
			defer cc.Close()
			eds := v2.NewEndpointDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			req := &v2.DiscoveryRequest{
				TypeUrl: endpointType,
			}
			_, err := eds.FetchEndpoints(ctx, req)
			check(t, err)
		},
		"FetchListeners": func(t *testing.T) {
			cc := newClient(t)
			defer cc.Close()
			lds := v2.NewListenerDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			req := &v2.DiscoveryRequest{
				TypeUrl: listenerType,
			}
			_, err := lds.FetchListeners(ctx, req)
			check(t, err)
		},
		"FetchRoutes": func(t *testing.T) {
			cc := newClient(t)
			defer cc.Close()
			rds := v2.NewRouteDiscoveryServiceClient(cc)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			req := &v2.DiscoveryRequest{
				TypeUrl: routeType,
			}
			_, err := rds.FetchRoutes(ctx, req)
			check(t, err)
		},
	}

	log := logrus.New()
	log.Out = &testWriter{t}
	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			et := &contour.EndpointsTranslator{
				FieldLogger: log,
			}
			ch := contour.CacheHandler{
				Metrics: metrics.NewMetrics(prometheus.NewRegistry()),
			}
			srv := NewAPI(log, map[string]Cache{
				clusterType:  &ch.ClusterCache,
				routeType:    &ch.RouteCache,
				listenerType: &ch.ListenerCache,
				endpointType: et,
			})
			var err error
			l, err = net.Listen("tcp", "127.0.0.1:0")
			check(t, err)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				srv.Serve(l)
			}()
			defer func() {
				srv.Stop()
				wg.Wait()
				l.Close()
			}()
			fn(t)
		})
	}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func sendreq(t *testing.T, stream interface {
	Send(*v2.DiscoveryRequest) error
}, typeurl string) {
	t.Helper()
	err := stream.Send(&v2.DiscoveryRequest{
		TypeUrl: typeurl,
	})
	check(t, err)
}

func checkrecv(t *testing.T, stream interface {
	Recv() (*v2.DiscoveryResponse, error)
}) {
	t.Helper()
	_, err := stream.Recv()
	check(t, err)
}

func checktimeout(t *testing.T, stream interface {
	Recv() (*v2.DiscoveryResponse, error)
}) {
	t.Helper()
	_, err := stream.Recv()
	if err == nil {
		t.Fatal("expected timeout")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatal(err)
	}
	if s.Code() != codes.DeadlineExceeded {
		t.Fatalf("expected %q, got %q", codes.DeadlineExceeded, s.Code())
	}
}

func testLogger(t *testing.T) logrus.FieldLogger {
	log := logrus.New()
	log.Out = &testWriter{t}
	return log
}

type testWriter struct {
	*testing.T
}

func (t *testWriter) Write(buf []byte) (int, error) {
	t.Helper()
	t.Logf("%s", buf)
	return len(buf), nil
}
