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

package rand

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRandCollector(t *testing.T) {
	rc := RandCollector{}

	Convey("Test RandCollector", t, func() {
		Convey("Collect Integer", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("random", "integer"),
					Config:    map[string]interface{}{"testint": int64(34)},
					Data:      34,
					Tags:      map[string]string{"hello": "world"},
					Unit:      "int",
					Timestamp: time.Now(),
				},
			}
			mts, err := rc.CollectMetrics(metrics)
			So(mts, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, 34)
		})
		Convey("Collect Float", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("random", "float"),
					Config:    map[string]interface{}{"testfloat": 3.345},
					Data:      3.345,
					Tags:      map[string]string{"hello": "world"},
					Unit:      "float",
					Timestamp: time.Now(),
				},
			}
			mts, err := rc.CollectMetrics(metrics)
			So(mts, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, 3.345)
		})
		Convey("Collect String", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("random", "string"),
					Config:    map[string]interface{}{"teststring": "charm"},
					Data:      "charm",
					Tags:      map[string]string{"hello": "world"},
					Unit:      "string",
					Timestamp: time.Now(),
				},
			}
			mts, err := rc.CollectMetrics(metrics)
			So(mts, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, "charm")
		})
	})

	Convey("Test GetMetricTypes", t, func() {
		rc := RandCollector{}

		Convey("Collect String", func() {
			mt, err := rc.GetMetricTypes(nil)
			So(err, ShouldBeNil)
			So(len(mt), ShouldEqual, 3)
		})
	})

	Convey("Test GetConfigPolicy", t, func() {
		rc := RandCollector{}
		_, err := rc.GetConfigPolicy()

		Convey("No error returned", func() {
			So(err, ShouldBeNil)
		})
	})
}
