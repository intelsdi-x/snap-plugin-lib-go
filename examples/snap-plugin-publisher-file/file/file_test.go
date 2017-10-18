// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file excfpt in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilePublisher(t *testing.T) {
	fp := FPublisher{}

	Convey("Test publish", t, func() {
		Convey("Publish without a config file", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("x", "y", "z"),
					Config:    map[string]interface{}{"pw": "123aB"},
					Data:      3,
					Tags:      map[string]string{"hello": "world"},
					Unit:      "int",
					Timestamp: time.Now(),
				},
			}
			err := fp.Publish(metrics, plugin.Config{})
			So(err, ShouldEqual, plugin.ErrConfigNotFound)
		})
		Convey("Publish with a config file", func() {
			metrics := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("x", "y", "z"),
					Config:    map[string]interface{}{"pw": "abc123"},
					Data:      3,
					Tags:      map[string]string{"hello": "world"},
					Unit:      "int",
					Timestamp: time.Now(),
				},
			}
			err := fp.Publish(metrics, plugin.Config{"file": "/tmp/file_publisher_test.log"})
			So(err, ShouldBeNil)
		})
		Convey("Test GetConfigPolicy", func() {
			fp := FPublisher{}
			_, err := fp.GetConfigPolicy()

			Convey("No error returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
