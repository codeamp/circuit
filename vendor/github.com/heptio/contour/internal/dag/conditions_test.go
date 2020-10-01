// Copyright © 2019 VMware
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

package dag

import (
	"testing"

	projcontour "github.com/projectcontour/contour/apis/projectcontour/v1"
	"github.com/projectcontour/contour/internal/assert"
)

func TestPathCondition(t *testing.T) {
	tests := map[string]struct {
		conditions []projcontour.Condition
		want       Condition
	}{
		"empty condition list": {
			conditions: nil,
			want:       &PrefixCondition{Prefix: "/"},
		},
		"single slash": {
			conditions: []projcontour.Condition{{
				Prefix: "/",
			}},
			want: &PrefixCondition{Prefix: "/"},
		},
		"two slashes": {
			conditions: []projcontour.Condition{{
				Prefix: "/",
			}, {
				Prefix: "/",
			}},
			want: &PrefixCondition{Prefix: "/"},
		},
		"mixed conditions": {
			conditions: []projcontour.Condition{{
				Prefix: "/a/",
			}, {
				Prefix: "/b",
			}},
			want: &PrefixCondition{Prefix: "/a/b"},
		},
		"trailing slash": {
			conditions: []projcontour.Condition{{
				Prefix: "/a/",
			}},
			want: &PrefixCondition{Prefix: "/a/"},
		},
		"trailing slash on second prefix condition": {
			conditions: []projcontour.Condition{{
				Prefix: "/a",
			},
				{
					Prefix: "/b/",
				}},
			want: &PrefixCondition{Prefix: "/a/b/"},
		},
		"nothing but slashes": {
			conditions: []projcontour.Condition{
				{
					Prefix: "///",
				},
				{
					Prefix: "/",
				}},
			want: &PrefixCondition{Prefix: "/"},
		},
		"header condition": {
			conditions: []projcontour.Condition{{
				Header: new(projcontour.HeaderCondition),
			}},
			want: &PrefixCondition{Prefix: "/"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := mergePathConditions(tc.conditions)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHeaderConditions(t *testing.T) {
	tests := map[string]struct {
		conditions []projcontour.Condition
		want       []HeaderCondition
	}{
		"empty condition list": {
			conditions: nil,
			want:       nil,
		},
		"prefix": {
			conditions: []projcontour.Condition{{
				Prefix: "/",
			}},
			want: nil,
		},
		"header condition empty": {
			conditions: []projcontour.Condition{{
				Header: new(projcontour.HeaderCondition),
			}},
			want: nil,
		},
		"header present": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name:    "x-request-id",
					Present: true,
				},
			}},
			want: []HeaderCondition{{
				Name:      "x-request-id",
				MatchType: "present",
			}},
		},
		"header name but missing condition": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name: "x-request-id",
				},
			}},
			// this should be filtered out beforehand, but in case it leaks
			// through the behavior is to ignore the header contains entry.
			want: nil,
		},
		"header contains": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name:     "x-request-id",
					Contains: "abcdef",
				},
			}},
			want: []HeaderCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
			}},
		},
		"header not contains": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name:        "x-request-id",
					NotContains: "abcdef",
				},
			}},
			want: []HeaderCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
				Invert:    true,
			}},
		},
		"header exact": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name:  "x-request-id",
					Exact: "abcdef",
				},
			}},
			want: []HeaderCondition{{
				Name:      "x-request-id",
				MatchType: "exact",
				Value:     "abcdef",
			}},
		},
		"header not exact": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name:     "x-request-id",
					NotExact: "abcdef",
				},
			}},
			want: []HeaderCondition{{
				Name:      "x-request-id",
				MatchType: "exact",
				Value:     "abcdef",
				Invert:    true,
			}},
		},
		"two header contains": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name:     "x-request-id",
					Contains: "abcdef",
				},
			}, {
				Header: &projcontour.HeaderCondition{
					Name:     "x-request-id",
					Contains: "cedfg",
				},
			}},
			want: []HeaderCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
			}, {
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "cedfg",
			}},
		},
		"two header contains different case": {
			conditions: []projcontour.Condition{{
				Header: &projcontour.HeaderCondition{
					Name:     "x-request-id",
					Contains: "abcdef",
				},
			}, {
				Header: &projcontour.HeaderCondition{
					Name:     "X-Request-Id",
					Contains: "abcdef",
				},
			}},
			want: []HeaderCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
			}, {
				Name:      "X-Request-Id",
				MatchType: "contains",
				Value:     "abcdef",
			}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := mergeHeaderConditions(tc.conditions)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPrefixConditionsValid(t *testing.T) {
	tests := map[string]struct {
		conditions []projcontour.Condition
		want       bool
	}{
		"empty condition list": {
			conditions: nil,
			want:       true,
		},
		"valid path condition only": {
			conditions: []projcontour.Condition{{
				Prefix: "/api",
			}},
			want: true,
		},
		"valid path condition with headers": {
			conditions: []projcontour.Condition{{
				Prefix: "/api",
				Header: &projcontour.HeaderCondition{
					Name:     "x-header",
					Contains: "abc",
				},
			}},
			want: true,
		},
		"two prefix conditions": {
			conditions: []projcontour.Condition{{
				Prefix: "/api",
			}, {
				Prefix: "/v1",
			}},
			want: false,
		},
		"two prefix conditions with headers": {
			conditions: []projcontour.Condition{{
				Prefix: "/api",
				Header: &projcontour.HeaderCondition{
					Name:     "x-header",
					Contains: "abc",
				},
			}, {
				Prefix: "/v1",
			}},
			want: false,
		},
		"invalid prefix condition": {
			conditions: []projcontour.Condition{{
				Prefix: "api",
			}},
			want: false,
		},
		"invalid prefix condition with headers": {
			conditions: []projcontour.Condition{{
				Prefix: "api",
				Header: &projcontour.HeaderCondition{
					Name:     "x-header",
					Contains: "abc",
				},
			}},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := pathConditionsValid(tc.conditions)
			assert.Equal(t, tc.want, err == nil)
		})
	}
}

func TestValidateHeaderConditions(t *testing.T) {
	tests := map[string]struct {
		conditions []projcontour.Condition
		wantErr    bool
	}{
		"empty condition list": {
			conditions: nil,
			wantErr:    false,
		},
		"prefix only": {
			conditions: []projcontour.Condition{
				{
					Prefix: "/blog",
				},
			},
			wantErr: false,
		},
		"valid conditions": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				},
			},
			wantErr: false,
		},
		"prefix conditions + valid headers": {
			conditions: []projcontour.Condition{
				{
					Prefix: "/blog",
				}, {
					Header: &projcontour.HeaderCondition{
						Name:        "x-header",
						NotContains: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:        "another-header",
						NotContains: "123",
					},
				},
			},
			wantErr: false,
		},
		"multiple 'exact' conditions for the same header are invalid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:  "x-header",
						Exact: "123",
					},
				},
			},
			wantErr: true,
		},
		"multiple 'exact' conditions for different headers are valid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:  "x-different-header",
						Exact: "123",
					},
				},
			},
			wantErr: false,
		},
		"'exact' and 'notexact' conditions for the same header with the same value are invalid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:     "x-header",
						NotExact: "abc",
					},
				},
			},
			wantErr: true,
		},
		"'exact' and 'notexact' conditions for the same header with different values are valid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:     "x-header",
						NotExact: "def",
					},
				},
			},
			wantErr: false,
		},
		"'exact' and 'notexact' conditions for different headers with the same value are valid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:     "x-another-header",
						NotExact: "abc",
					},
				},
			},
			wantErr: false,
		},
		"'contains' and 'notcontains' conditions for the same header with the same value are invalid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:        "x-header",
						NotContains: "abc",
					},
				},
			},
			wantErr: true,
		},
		"'contains' and 'notcontains' conditions for the same header with different values are valid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:        "x-header",
						NotContains: "def",
					},
				},
			},
			wantErr: false,
		},
		"'contains' and 'notcontains' conditions for different headers with the same value are valid": {
			conditions: []projcontour.Condition{
				{
					Header: &projcontour.HeaderCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				}, {
					Header: &projcontour.HeaderCondition{
						Name:        "x-another-header",
						NotContains: "abc",
					},
				},
			},
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr := headerConditionsValid(tc.conditions)

			if !tc.wantErr && gotErr != nil {
				t.Fatalf("Expected no error, got (%v)", gotErr)
			}
			if tc.wantErr && gotErr == nil {
				t.Fatalf("Expected error, got none")
			}
		})
	}
}
