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

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPublisher(t *testing.T) {
	Convey("Test Publisher", t, func() {
		Convey("Error while publishing", func() {
			pp := publisherProxy{
				pluginProxy: *newPluginProxy(newMockPublisher()),
				plugin:      newMockErrPublisher(),
			}
			_, err := pp.Publish(context.Background(), &rpc.PubProcArg{})
			So(err, ShouldNotBeNil)
		})
		Convey("Succeed while publishing", func() {
			pp := publisherProxy{
				pluginProxy: *newPluginProxy(newMockPublisher()),
				plugin:      newMockPublisher(),
			}

			input, err := getTestData()
			So(err, ShouldBeNil)

			_, err = pp.Publish(context.Background(), &rpc.PubProcArg{Metrics: input})
			So(err, ShouldBeNil)
		})
	})
}

func getTestData() ([]*rpc.Metric, error) {
	input := []*rpc.Metric{}

	mp := getMetricData()
	for _, v := range mp {
		m, err := toProtoMetric(v)
		if err != nil {
			return nil, err
		}
		input = append(input, m)
	}
	return input, nil
}
