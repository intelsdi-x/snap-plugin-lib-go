// +build small medium

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
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	metricMap = getMetricData()
)

func TestPlugin(t *testing.T) {
	Convey("Test Metrics", t, func() {
		i := StartCollector(newMockCollector(), "collector", 0, Exclusive(true), RoutingStrategy(1))
		So(i, ShouldEqual, 0)

		j := StartProcessor(newMockProcessor(), "processor", 1, Exclusive(false))
		So(j, ShouldEqual, 0)

		k := StartPublisher(newMockPublisher(), "publisher", 2, Exclusive(false))
		So(k, ShouldEqual, 0)
	})

}

type mockPlugin struct {
	err error
}

func newMockPlugin() *mockPlugin {
	return &mockPlugin{}
}

func newMockErrPlugin() *mockPlugin {
	return &mockPlugin{err: errors.New("error")}
}

func (mp *mockPlugin) GetConfigPolicy() (ConfigPolicy, error) {
	if mp.err != nil {
		return ConfigPolicy{}, errors.New("error")
	}
	cp := NewConfigPolicy()

	cp.AddBoolRule([]string{"log"}, boolRule{Key: "logLevel", Required: true, Default: true, HasDefault: true})
	cp.AddBoolRule([]string{"cache"}, boolRule{Key: "cacheTime", Required: true, Default: false, HasDefault: true})

	cp.AddFloatRule([]string{"float"}, floatRule{Key: "low", Required: true, Default: 32.1, HasDefault: true})
	cp.AddFloatRule([]string{"cache"}, floatRule{Key: "high", Required: true, Default: 2399.58, HasDefault: true})

	cp.AddIntRule([]string{"xyz"}, integerRule{Key: "logLevel", Required: false, Default: 30, HasDefault: true})
	cp.AddIntRule([]string{"abc"}, integerRule{Key: "cacheTime", Required: true, Default: 50, HasDefault: true})

	cp.AddStringRule([]string{"log"}, stringRule{Key: "logLevel", Required: true, Default: "123", HasDefault: true})
	cp.AddStringRule([]string{"cache"}, stringRule{Key: "cacheTime", Required: true, Default: "tyty", HasDefault: true})

	return (*cp), nil
}

type mockCollector struct {
	mockPlugin
	err error
}

func newMockCollector() *mockCollector {
	return &mockCollector{}
}

func newMockErrCollector() *mockCollector {
	return &mockCollector{err: errors.New("empty")}
}

func (mc *mockCollector) GetMetricTypes(cfg Config) ([]Metric, error) {
	if mc.err != nil {
		return nil, errors.New("error")
	}

	mts := []Metric{}
	for _, v := range metricMap {
		mts = append(mts, v)
	}
	return mts, nil
}

func (mc *mockCollector) CollectMetrics(mts []Metric) ([]Metric, error) {
	if mc.err != nil {
		return nil, errors.New("error")
	}
	return mts, nil
}

type mockProcessor struct {
	mockPlugin
	err error
}

func newMockProcessor() *mockProcessor {
	return &mockProcessor{}
}

func newMockErrProcessor() *mockProcessor {
	return &mockProcessor{err: errors.New("error")}
}

func (mp *mockProcessor) Process(mts []Metric, cfg Config) ([]Metric, error) {
	if mp.err != nil {
		return nil, mp.err
	}

	metrics := []Metric{}
	for _, m := range mts {
		if m.Version%2 == 0 {
			metrics = append(metrics, m)
		}
	}
	return metrics, nil
}

type mockPublisher struct {
	mockPlugin
	err error
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{}
}

func newMockErrPublisher() *mockPublisher {
	return &mockPublisher{err: errors.New("error")}
}

func (mpb *mockPublisher) Publish(mts []Metric, cfg Config) error {
	if mpb.err != nil {
		return mpb.err
	}
	return nil
}

func getMetricData() map[string]Metric {
	mm := map[string]Metric{}
	for i := 0; i < 10; i++ {
		m := Metric{
			Namespace: NewNamespace("a", "b", "c"),
			Version:   int64(i),
			Config:    map[string]interface{}{"key": 123},
			Data:      i,
			Tags:      map[string]string{fmt.Sprintf("tag%d", i): fmt.Sprintf("value%d", i)},
			Unit:      "int",
			Timestamp: time.Now(),
		}
		idx := fmt.Sprintf("%s.%d", m.Namespace, m.Version)
		mm[idx] = m
	}
	return mm
}

type mockServer struct {
}

func (ms *mockServer) Serve(net.Listener) error {
	return errors.New("error")
}
