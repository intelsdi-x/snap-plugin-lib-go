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
See the License for the rpecific language governing permissions and
limitations under the License.
*/

package reverse

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReverseProcessor(t *testing.T) {
	rp := RProcessor{}

	Convey("Test Processor", t, func() {
		Convey("Process int metric", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("x", "y", "z"),
					Config:    map[string]interface{}{"pw": "123aB"},
					Data:      345678,
					Tags:      map[string]string{"hello": "world"},
					Unit:      "int",
					Timestamp: time.Now(),
				},
			}
			mts, err := rp.Process(metrics, plugin.Config{})
			So(mts, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, 876543)
		})
		Convey("Process int32 metric", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("x", "y", "z"),
					Config:    map[string]interface{}{"pw": "123aB"},
					Data:      int32(345678),
					Tags:      map[string]string{"hello": "world"},
					Unit:      "int",
					Timestamp: time.Now(),
				},
			}
			mts, err := rp.Process(metrics, plugin.Config{})
			So(mts, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, int32(876543))
		})
		Convey("Process int64 metric", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("x", "y", "z"),
					Config:    map[string]interface{}{"pw": "123aB"},
					Data:      int64(345678999999),
					Tags:      map[string]string{"hello": "world"},
					Unit:      "int",
					Timestamp: time.Now(),
				},
			}
			mts, err := rp.Process(metrics, plugin.Config{})
			So(mts, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, int64(999999876543))
		})
		Convey("Process float metric", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("x", "y", "z"),
					Config:    map[string]interface{}{"pw": 123},
					Data:      42.42,
					Tags:      map[string]string{"hello": "world"},
					Unit:      "float",
					Timestamp: time.Now(),
				},
			}
			mts, err := rp.Process(metrics, plugin.Config{})
			So(mts, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, 24.24)
		})
		Convey("Process float32 metric", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("x", "y", "z"),
					Config:    map[string]interface{}{"pw": 123},
					Data:      float32(24.24),
					Tags:      map[string]string{"hello": "world"},
					Unit:      "float32",
					Timestamp: time.Now(),
				},
			}
			mts, err := rp.Process(metrics, plugin.Config{})
			So(mts, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, float32(42.42))
		})
		Convey("Process string metric", func() {
			curTime := time.Now()
			metrics := []plugin.Metric{
				{
					Namespace:   plugin.NewNamespace("x", "y", "z"),
					Config:      map[string]interface{}{"pw": 123.78},
					Data:        "luck charm",
					Tags:        map[string]string{"hello": "world"},
					Unit:        "string",
					Timestamp:   curTime,
					Version:     2,
					Description: "世界你好",
				},
			}
			mts, err := rp.Process(metrics, plugin.Config{})
			So(mts, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(mts[0].Data, ShouldEqual, "mrahc kcul")
			So(mts[0].Tags, ShouldResemble, map[string]string{"hello": "world"})
			So(mts[0].Unit, ShouldEqual, "string")
			So(mts[0].Timestamp, ShouldResemble, curTime)
			So(mts[0].Version, ShouldEqual, 2)
			So(mts[0].Description, ShouldEqual, "世界你好")
		})

		Convey("Test GetConfigPolicy", func() {
			rp := RProcessor{}
			_, err := rp.GetConfigPolicy()

			Convey("No error returned", func() {
				So(err, ShouldBeNil)
			})

		})
	})
}
