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
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/heptio/contour/internal/dag"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testLogger(t *testing.T) logrus.FieldLogger {
	log := logrus.New()
	log.Out = &testWriter{t}
	return log
}

type testWriter struct {
	*testing.T
}

func (t *testWriter) Write(buf []byte) (int, error) {
	t.Logf("%s", buf)
	return len(buf), nil
}

func contents(v interface {
	Values(func(string) bool) []proto.Message
}) []proto.Message {
	return v.Values(func(string) bool { return true })
}

func endpoints(ns, name string, subsets ...v1.EndpointSubset) *v1.Endpoints {
	return &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Subsets: subsets,
	}
}

func addresses(ips ...string) []v1.EndpointAddress {
	var addrs []v1.EndpointAddress
	for _, ip := range ips {
		addrs = append(addrs, v1.EndpointAddress{IP: ip})
	}
	return addrs
}

func ports(ps ...int32) []v1.EndpointPort {
	var ports []v1.EndpointPort
	for _, p := range ps {
		ports = append(ports, v1.EndpointPort{Port: p})
	}
	return ports
}

type nullNotifier int

func (nn *nullNotifier) OnChange(b *dag.Builder) {}
