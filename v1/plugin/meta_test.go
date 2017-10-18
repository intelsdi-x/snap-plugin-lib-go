// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package plugin

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMeta(t *testing.T) {
	tc := metaTestCases()

	Convey("Test Meta", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test Meta %+v", c.input.Name), func() {
				So(c.input, ShouldResemble, c.output)
			})
		}
	})
}

type metaTestCase struct {
	input  meta
	output meta
}

func metaTestCases() []metaTestCase {
	tc := []metaTestCase{
		{
			input:  *newMeta(collectorType, "fakeCollector", 0, ConcurrencyCount(0), Exclusive(false), RoutingStrategy(LRURouter), CacheTTL(time.Millisecond*0)),
			output: meta{Type: collectorType, Name: "fakeCollector", Version: 0, ConcurrencyCount: 0, Exclusive: false, RoutingStrategy: 0, RPCType: gRPC, RPCVersion: 1, Unsecure: true, CacheTTL: time.Millisecond * 0},
		},
		{
			input:  *newMeta(processorType, "fakeProcessor", 1, ConcurrencyCount(1), Exclusive(true), RoutingStrategy(StickyRouter), CacheTTL(time.Millisecond*1)),
			output: meta{Type: processorType, Name: "fakeProcessor", Version: 1, ConcurrencyCount: 1, Exclusive: true, RoutingStrategy: 1, RPCType: gRPC, RPCVersion: 1, Unsecure: true, CacheTTL: time.Millisecond * 1},
		},
		{
			input:  *newMeta(publisherType, "fakePublisher", 10, ConcurrencyCount(8), Exclusive(false), RoutingStrategy(ConfigBasedRouter), CacheTTL(time.Millisecond*1)),
			output: meta{Type: publisherType, Name: "fakePublisher", Version: 10, ConcurrencyCount: 8, Exclusive: false, RoutingStrategy: 2, RPCType: gRPC, RPCVersion: 1, Unsecure: true, CacheTTL: time.Millisecond * 1},
		},
	}
	return tc
}
