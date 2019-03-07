/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPortNameMapUnion(t *testing.T) {
	testcases := []struct {
		desc        string
		p1          PortNameMap
		p2          PortNameMap
		expectedMap PortNameMap
	}{
		{
			"empty map union empty map",
			PortNameMap{},
			PortNameMap{},
			PortNameMap{},
		},
		{
			"empty map union a non-empty map is the non-empty map",
			PortNameMap{},
			PortNameMap{80: "namedport", 443: "3000"},
			PortNameMap{80: "namedport", 443: "3000"},
		},
		{
			"union of two non-empty maps",
			PortNameMap{443: "3000", 5000: "6000"},
			PortNameMap{80: "namedport", 8080: "9000"},
			PortNameMap{80: "namedport", 443: "3000", 5000: "6000", 8080: "9000"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			result := tc.p1.Union(tc.p2)
			if !reflect.DeepEqual(result, tc.expectedMap) {
				t.Errorf("Expected p1.Union(p2) to equal: %v; got: %v", tc.expectedMap, result)
			}
		})
	}
}

func TestPortNameMapDifference(t *testing.T) {
	testcases := []struct {
		desc        string
		p1          PortNameMap
		p2          PortNameMap
		expectedMap PortNameMap
	}{
		{
			"empty map difference empty map",
			PortNameMap{},
			PortNameMap{},
			PortNameMap{},
		},
		{
			"empty map difference a non-empty map is empty map",
			PortNameMap{},
			PortNameMap{80: "namedport", 443: "3000"},
			PortNameMap{},
		},
		{
			"non-empty map difference a non-empty map is the non-empty map",
			PortNameMap{80: "namedport", 443: "3000"},
			PortNameMap{},
			PortNameMap{80: "namedport", 443: "3000"},
		},
		{
			"difference of two non-empty maps with the same elements",
			PortNameMap{80: "namedport", 443: "3000"},
			PortNameMap{80: "namedport", 443: "3000"},
			PortNameMap{},
		},
		{
			"difference of two non-empty maps with no elements in common returns p1",
			PortNameMap{443: "3000", 5000: "6000"},
			PortNameMap{80: "namedport", 8080: "9000"},
			PortNameMap{443: "3000", 5000: "6000"},
		},
		{
			"difference of two non-empty maps with elements in common",
			PortNameMap{80: "namedport", 443: "3000", 5000: "6000", 8080: "9000"},
			PortNameMap{80: "namedport", 8080: "9000"},
			PortNameMap{443: "3000", 5000: "6000"},
		},
		{
			"difference of two non-empty maps with a key in common but different in value",
			PortNameMap{80: "namedport"},
			PortNameMap{80: "8080", 8080: "9000"},
			PortNameMap{80: "namedport"},
		},
		{
			"difference of two non-empty maps with 2 keys in common but different in values",
			PortNameMap{80: "namedport", 443: "8443"},
			PortNameMap{80: "8080", 443: "9443"},
			PortNameMap{80: "namedport", 443: "8443"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			result := tc.p1.Difference(tc.p2)
			if !reflect.DeepEqual(result, tc.expectedMap) {
				t.Errorf("Expected p1.Difference(p2) to equal: %v; got: %v", tc.expectedMap, result)
			}
		})
	}
}

func TestParseNegStatus(t *testing.T) {
	for _, tc := range []struct {
		desc            string
		status          string
		expectNegStatus *NegStatus
		expectError     error
	}{
		{
			desc:            "Test empty string",
			status:          "",
			expectNegStatus: &NegStatus{},
			expectError:     fmt.Errorf("unexpected end of JSON input"),
		},
		{
			desc:            "Test basic status",
			status:          `{"network_endpoint_groups":{"80":"neg-name"},"zones":["us-central1-a"]}`,
			expectNegStatus: &NegStatus{NetworkEndpointGroups: PortNameMap{80: "neg-name"}, Zones: []string{"us-central1-a"}},
			expectError:     nil,
		},
		{
			desc:            "Incorrect fields",
			status:          `{"network_endpoint_group":{"80":"neg-name"},"zone":["us-central1-a"]}`,
			expectNegStatus: &NegStatus{},
			expectError:     nil,
		},
		{
			desc:            "Invalid annotation",
			status:          `{"network_endpoint_groups":{"80":"neg-name"},"zones":"us-central1-a"]}`,
			expectNegStatus: &NegStatus{},
			expectError:     fmt.Errorf("invalid character ']' after object key:value pair"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			negStatus, err := ParseNegStatus(tc.status)

			if fmt.Sprintf("%q", err) != fmt.Sprintf("%q", tc.expectError) {
				t.Errorf("Test case %q: Expect error to be %q, but got: %q", tc.desc, tc.expectError, err)
			}

			if !reflect.DeepEqual(*tc.expectNegStatus, negStatus) {
				t.Errorf("Expected NegStatus to be %v, got %v instead", tc.expectNegStatus, negStatus)
			}
		})
	}
}
