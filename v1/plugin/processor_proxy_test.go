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
	"testing"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestProcessor(t *testing.T) {
	Convey("Test Processor", t, func() {
		Convey("Error while processing", func() {
			pp := processorProxy{
				pluginProxy: *newPluginProxy(newMockProcessor()),
				plugin:      newMockErrProcessor(),
			}
			_, err := pp.Process(context.Background(), &rpc.PubProcArg{})
			So(err, ShouldNotBeNil)
		})
		Convey("Succeed while processing", func() {
			pp := processorProxy{
				pluginProxy: *newPluginProxy(newMockProcessor()),
				plugin:      newMockProcessor(),
			}

			input, err := getTestData()
			So(err, ShouldBeNil)

			cfg := rpc.ConfigMap{FloatMap: map[string]float64{"xyz": 2.2}}
			reply, err := pp.Process(context.Background(), &rpc.PubProcArg{Metrics: input, Config: &cfg})
			So(err, ShouldBeNil)
			So(len(reply.GetMetrics()), ShouldEqual, 5)
		})
	})
}
