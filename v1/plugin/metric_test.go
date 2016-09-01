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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMetric(t *testing.T) {
	tc := metricTestCases()

	Convey("Test Metrics", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test Metrics %+v", c.input.Namespace.String()), func() {
				ns := c.input.Namespace
				nsStr := ns.String()
				nsArr := ns.Strings()
				So(strings.TrimLeft(nsStr, "/"), ShouldEqual, strings.Join(nsArr, "/"))
				So(ns.Key(), ShouldEqual, strings.Join(nsArr, "."))

				for i, n := range ns {
					elem := ns.Element(i)
					if n.IsDynamic() {
						So(n.Value, ShouldEqual, "*")
						So(elem.Value, ShouldEqual, "*")
						So(elem.Name, ShouldNotBeEmpty)

						b, idx := c.input.Namespace.IsDynamic()
						So(b, ShouldEqual, true)
						So(len(idx), ShouldBeGreaterThan, 0)
					} else {
						So(elem.IsDynamic(), ShouldEqual, false)
						newElem := NewNamespaceElement(elem.Value)
						So(newElem.Value, ShouldEqual, elem.Value)
					}
				}

				em := ns.Element(len(ns))
				So(em.Value, ShouldBeEmpty)
			})
		}
	})
}

func TestToFromProtoNamespace(t *testing.T) {
	nss := metricNamespaceTestCases()

	Convey("Test ToFromProtoNamespace", t, func() {
		for _, ns := range nss {
			Convey(fmt.Sprintf("Test ToFromProtoNamespace %+v", ns.String()), func() {
				protoNs := toProtoNamespace(ns)
				fromProtoNs := fromProtoNamespace(protoNs)
				So(fromProtoNs, ShouldResemble, ns)
			})
		}
	})
}

func TestToFromProtoMetric(t *testing.T) {
	tc := metricTestCases()

	Convey("Test ToFromProtoMetric", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test ToFromProtoMetric %+v", c.input.Namespace.String()), func() {
				protoMetric, err := toProtoMetric(c.input)
				So(err, ShouldBeNil)

				mts := fromProtoMetric(protoMetric)
				So(mts.Namespace, ShouldResemble, c.input.Namespace)
				So(mts.Tags, ShouldEqual, c.input.Tags)
				So(mts.Unit, ShouldEqual, c.input.Unit)
				So(mts.Version, ShouldEqual, c.input.Version)
				So(mts.Description, ShouldEqual, c.input.Description)

				rv := reflect.ValueOf(c.input.Data)
				switch rv.Kind() {
				case reflect.Slice:
					So(mts.Data, ShouldResemble, c.input.Data)
				default:
					So(mts.Data, ShouldEqual, c.input.Data)
				}
			})
		}
	})
}

func TestMetricConfig(t *testing.T) {
	tc := metricConfigTestCases()

	Convey("Test Metric Config", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test Metric Config %+v", c.input.String()), func() {
				cfg := fromProtoConfig(&c.input)

				for k, v := range c.expect {
					switch v.(type) {
					case int64:
						i, _ := cfg.GetInt(k)
						So(i, ShouldEqual, v)
					case float64:
						f, _ := cfg.GetFloat(k)
						So(f, ShouldEqual, v)
					case string:
						s, _ := cfg.GetString(k)
						So(s, ShouldEqual, v)
					case bool:
						b, _ := cfg.GetBool(k)
						So(b, ShouldEqual, v)
					}
				}
			})
		}
	})
}

func TestMetricTime(t *testing.T) {
	tc := metricNoDefaultTimeTestCases()

	Convey("Test metric has no default time", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test metric has no default time %+v", c.input.Namespace.String()), func() {
				protoMetric, err := toProtoMetric(c.input)
				So(err, ShouldBeNil)

				So(protoMetric.Timestamp.Sec, ShouldNotEqual, int64(-62135596800))
				So(protoMetric.LastAdvertisedTime.Sec, ShouldNotEqual, int64(-62135596800))
			})
		}
	})

	tc = metricHasDefaultTimeTestCases()

	Convey("Test metric has time set", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test metric has time set %+v", c.input.Namespace.String()), func() {
				protoMetric, err := toProtoMetric(c.input)
				So(err, ShouldBeNil)

				mts := fromProtoMetric(protoMetric)
				So(mts.Timestamp, ShouldResemble, c.input.Timestamp)
				So(mts.lastAdvertisedTime, ShouldResemble, c.input.lastAdvertisedTime)
			})
		}
	})
}

type metricInput struct {
	metrics []Metric
}

type testCaseMetric struct {
	input Metric
}

func metricTestCases() []testCaseMetric {
	tc := []testCaseMetric{
		testCaseMetric{
			input: Metric{
				Namespace: NewNamespace("a", "b", "c"),
				Version:   0,
				Config: map[string]interface{}{
					"user": "cindy",
					"pw":   "12345Y",
				},
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace: NewNamespace("a1", "b1", "c1").AddStaticElement("d").AddDynamicElement("charm", "desc").AddStaticElements("x", "y"),
				Data:      "abc",
				Unit:      "string",
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace: NewNamespace("a2", "b2", "c2"),
				Version:   1,
				Timestamp: time.Now(),
				Data:      int32(123),
				Tags:      map[string]string{"label": "abc"},
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("a3", "b3", "c3"),
				Version:            2,
				Timestamp:          time.Now(),
				Data:               int64(123),
				Tags:               map[string]string{"label": "abc"},
				Description:        "desc2",
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("a4", "b4", "c4"),
				Version:            3,
				Timestamp:          time.Now(),
				Data:               true,
				Tags:               map[string]string{"label": "abc"},
				Description:        "desc3",
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("a5", "b5", "c5"),
				Version:            4,
				Timestamp:          time.Now(),
				Data:               float32(123.1),
				Tags:               map[string]string{"label": "abc"},
				Unit:               "float32",
				Description:        "desc4",
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("a6", "b6", "c7"),
				Version:            5,
				Timestamp:          time.Now(),
				Data:               float64(123.3),
				Tags:               map[string]string{"label": "abc"},
				Description:        "desc5",
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("a8", "b8", "c8"),
				Version:            6,
				Timestamp:          time.Now(),
				Data:               []byte("abc"),
				Tags:               map[string]string{"label": "abc"},
				Description:        "desc",
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("a9", "b9", "29"),
				Version:            7,
				Timestamp:          time.Now(),
				Data:               int(1),
				Tags:               map[string]string{"label": "abc"},
				Description:        "desc",
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("a10", "b10", "c10"),
				Version:            9,
				Timestamp:          time.Now(),
				Data:               nil,
				Tags:               map[string]string{"label": "abc"},
				Description:        "desc",
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace(NewNamespaceElement("").Value),
				Version:            10,
				Timestamp:          time.Now(),
				Data:               nil,
				Tags:               map[string]string{"label": "abc"},
				Description:        "desc",
				lastAdvertisedTime: time.Now(),
				Unit:               "object",
			},
		}}
	return tc
}

func metricNoDefaultTimeTestCases() []testCaseMetric {
	tc := []testCaseMetric{
		testCaseMetric{
			input: Metric{
				Namespace: NewNamespace("a", "b", "c"),
				Version:   0,
				Config: map[string]interface{}{
					"user": "cindy",
					"pw":   "12345Y",
				},
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace: NewNamespace("a1", "b1", "c1").AddStaticElement("d").AddDynamicElement("charm", "desc").AddStaticElements("x", "y"),
				Data:      "abc",
				Unit:      "string",
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace: NewNamespace("a2", "b2", "c2"),
				Version:   1,
				Data:      int32(123),
				Tags:      map[string]string{"label": "abc"},
			},
		},
	}
	return tc
}

func metricHasDefaultTimeTestCases() []testCaseMetric {
	tc := []testCaseMetric{
		testCaseMetric{
			input: Metric{
				Namespace: NewNamespace("a", "b", "c"),
				Version:   0,
				Config: map[string]interface{}{
					"user": "cindy",
					"pw":   "12345Y",
				},
				Timestamp:          time.Now(),
				lastAdvertisedTime: time.Now(),
			},
		},
		testCaseMetric{
			input: Metric{
				Namespace:          NewNamespace("x", "y", "z").AddStaticElement("d").AddDynamicElement("charm", "desc").AddStaticElements("r", "s"),
				Data:               "abc",
				Unit:               "string",
				Timestamp:          time.Now(),
				lastAdvertisedTime: time.Now(),
			},
		},
	}
	return tc
}

type testCaseMetricConfig struct {
	expect map[string]interface{}
	input  rpc.ConfigMap
}

func metricConfigTestCases() []testCaseMetricConfig {
	tc := []testCaseMetricConfig{
		testCaseMetricConfig{
			input: rpc.ConfigMap{
				IntMap:    map[string]int64{"abc": 123},
				StringMap: map[string]string{"xyz": "abc"},
				FloatMap:  map[string]float64{"rst": 32.5},
				BoolMap:   map[string]bool{"hasDefault": true},
			},
			expect: map[string]interface{}{
				"abc":        123,
				"xyz":        "abc",
				"rst":        32.5,
				"hasDefault": true,
			},
		},
	}
	return tc
}

func metricNamespaceTestCases() []namespace {
	nss := []namespace{
		namespace{
			namespaceElement{Value: "a"},
			namespaceElement{Value: "b"},
			namespaceElement{Value: "c"},
		},
		namespace{
			namespaceElement{Value: "a"},
			namespaceElement{Name: "rst", Value: "*", Description: "range"},
			namespaceElement{Value: "c"},
			namespaceElement{Name: "party", Value: "*", Description: "lol"},
			namespaceElement{Value: "d"},
		},
	}
	return nss
}
