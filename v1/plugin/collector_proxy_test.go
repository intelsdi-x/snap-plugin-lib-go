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

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMetricTypes(t *testing.T) {

	Convey("Test GetMetricTypes", t, func() {
		Convey("valid metric types", func() {
			cp := collectorProxy{
				pluginProxy: *newPluginProxy(newMockCollector()),
				plugin:      newMockCollector(),
			}
			r, err := cp.GetMetricTypes(context.Background(),
				&rpc.GetMetricTypesArg{Config: &rpc.ConfigMap{BoolMap: map[string]bool{"error": false}}})
			So(err, ShouldBeNil)

			for _, m := range r.GetMetrics() {
				tm := fromProtoMetric(m)
				idx := fmt.Sprintf("%s.%d", tm.Namespace, tm.Version)
				So(tm.Namespace.Strings(), ShouldResemble, getMockMetricDataMap()[idx].Namespace.Strings())
				So(tm.Tags, ShouldResemble, getMockMetricDataMap()[idx].Tags)
			}
		})
		Convey("invalid metric types", func() {
			cp := collectorProxy{
				pluginProxy: *newPluginProxy(newMockCollector()),
				plugin:      newMockErrCollector(),
			}
			r, err := cp.GetMetricTypes(context.Background(), &rpc.GetMetricTypesArg{})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})
	})
}

func TestCollectMetrics(t *testing.T) {

	Convey("Test CollectMetrics", t, func() {
		mp := getTestMetricData()
		ms := []*rpc.Metric{}

		for _, v := range mp {
			ms = append(ms, v)
		}
		Convey("Error while collecting", func() {
			cp := collectorProxy{
				pluginProxy: *newPluginProxy(newMockErrCollector()),
				plugin:      newMockErrCollector(),
			}
			reply, err := cp.CollectMetrics(context.Background(), &rpc.MetricsArg{Metrics: ms})
			So(err, ShouldNotBeNil)
			So(reply, ShouldBeNil)
		})
		Convey("Succeed while collecting", func() {
			cp := collectorProxy{
				pluginProxy: *newPluginProxy(newMockErrCollector()),
				plugin:      newMockCollector(),
			}
			reply, err := cp.CollectMetrics(context.Background(), &rpc.MetricsArg{Metrics: ms})
			So(err, ShouldBeNil)
			So(len(reply.GetMetrics()), ShouldEqual, len(ms))

			for _, v := range reply.GetMetrics() {
				m := fromProtoMetric(v)
				So(v.Tags, ShouldEqual, m.Tags)

				var nsArr []string
				ns := v.GetNamespace()
				for i := range ns {
					nsArr = append(nsArr, ns[i].Value)
				}
				Convey(fmt.Sprintf("colleting namespace: %v", m.Namespace.Strings()), func() {
					So(nsArr, ShouldResemble, m.Namespace.Strings())
				})
			}
		})
	})
}

func getTestMetricData() map[string]*rpc.Metric {
	mp := map[string]*rpc.Metric{}

	// test case for string data and StringMap
	for i := 0; i < 2; i++ {
		name := fmt.Sprintf("name%d", i)
		val := fmt.Sprintf("val%d", i)
		desc := fmt.Sprintf("desc%d", i)
		ns := []*rpc.NamespaceElement{
			{Name: name, Value: val, Description: desc},
		}
		m := &rpc.Metric{
			Namespace:          ns,
			Timestamp:          &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			LastAdvertisedTime: &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			Version:            int64(i),
			Tags:               map[string]string{"x": "123"},
			Data:               &rpc.Metric_StringData{StringData: "abc"},
			Unit:               "string",
			Config:             &rpc.ConfigMap{StringMap: map[string]string{"xyz": "123"}},
		}
		mp[val] = m
	}

	// test case for float32 data and FloatMap
	for i := 2; i < 4; i++ {
		name := fmt.Sprintf("name%d", i)
		val := fmt.Sprintf("val%d", i)
		desc := fmt.Sprintf("desc%d", i)
		ns := []*rpc.NamespaceElement{
			{Name: name, Value: val, Description: desc},
		}
		m := &rpc.Metric{
			Namespace:          ns,
			Timestamp:          &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			LastAdvertisedTime: &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			Version:            int64(i),
			Tags:               map[string]string{"x": "float32"},
			Data:               &rpc.Metric_Float32Data{Float32Data: 2.3},
			Unit:               "float32",
			Config:             &rpc.ConfigMap{FloatMap: map[string]float64{"xyz": 3.2}},
		}
		mp[val] = m
	}

	// test case for float64 data and IntMap
	for i := 4; i < 6; i++ {
		name := fmt.Sprintf("name%d", i)
		val := fmt.Sprintf("val%d", i)
		desc := fmt.Sprintf("desc%d", i)
		ns := []*rpc.NamespaceElement{
			{Name: name, Value: val, Description: desc},
		}

		m := &rpc.Metric{
			Namespace:          ns,
			Timestamp:          &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			LastAdvertisedTime: &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			Version:            int64(i),
			Tags:               map[string]string{"x": "123"},
			Data:               &rpc.Metric_Float64Data{Float64Data: 7.8},
			Unit:               "float64",
			Config:             &rpc.ConfigMap{IntMap: map[string]int64{"xyz": 123}},
		}
		mp[val] = m
	}

	// test case for int32 data and FloatMap
	for i := 6; i < 8; i++ {
		name := fmt.Sprintf("name%d", i)
		val := fmt.Sprintf("val%d", i)
		desc := fmt.Sprintf("desc%d", i)
		ns := []*rpc.NamespaceElement{
			{Name: name, Value: val, Description: desc},
		}
		m := &rpc.Metric{
			Namespace:          ns,
			Timestamp:          &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			LastAdvertisedTime: &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			Version:            int64(i),
			Tags:               map[string]string{"x": "123"},
			Data:               &rpc.Metric_Int32Data{Int32Data: 2},
			Unit:               "int32",
			Config:             &rpc.ConfigMap{FloatMap: map[string]float64{"xyz": 123.2}},
		}
		mp[val] = m
	}

	// test case for int64 data and IntMap
	for i := 8; i < 10; i++ {
		name := fmt.Sprintf("name%d", i)
		val := fmt.Sprintf("val%d", i)
		desc := fmt.Sprintf("desc%d", i)
		ns := []*rpc.NamespaceElement{
			{Name: name, Value: val, Description: desc},
		}
		m := &rpc.Metric{
			Namespace:          ns,
			Timestamp:          &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			LastAdvertisedTime: &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			Version:            int64(i),
			Tags:               map[string]string{"x": "123"},
			Data:               &rpc.Metric_Int64Data{Int64Data: 1},
			Unit:               "int64",
			Config:             &rpc.ConfigMap{IntMap: map[string]int64{"xyz": 123}},
		}
		mp[val] = m
	}

	// test case for bytes data and BoolMap
	for i := 10; i < 12; i++ {
		name := fmt.Sprintf("name%d", i)
		val := fmt.Sprintf("val%d", i)
		desc := fmt.Sprintf("desc%d", i)
		ns := []*rpc.NamespaceElement{
			{Name: name, Value: val, Description: desc},
		}
		m := &rpc.Metric{
			Namespace:          ns,
			Timestamp:          &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			LastAdvertisedTime: &rpc.Time{Sec: int64(10), Nsec: int64(9)},
			Version:            int64(i),
			Tags:               map[string]string{"x": "123"},
			Data:               &rpc.Metric_BytesData{BytesData: []byte("123")},
			Unit:               "byte",
			Config:             &rpc.ConfigMap{BoolMap: map[string]bool{"xyz": false}},
		}
		mp[val] = m
	}
	return mp
}
