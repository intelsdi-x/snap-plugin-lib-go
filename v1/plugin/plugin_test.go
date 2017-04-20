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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	metricMap = getMetricData()
)

func TestPlugin(t *testing.T) {
	Convey("Basing on plugin lib routines", t, func() {
		var mockInputOutput = newMockInputOutput(libInputOutput)
		libInputOutput = mockInputOutput
		Convey("collector plugin should start successfully", func() {
			i := StartCollector(newMockCollector(), "collector", 0, Exclusive(true), RoutingStrategy(1))
			So(i, ShouldEqual, 0)
		})
		Convey("processor plugin should start successfully", func() {
			j := StartProcessor(newMockProcessor(), "processor", 1, Exclusive(false))
			So(j, ShouldEqual, 0)
		})
		Convey("publisher plugin should start successfully", func() {
			k := StartPublisher(newMockPublisher(), "publisher", 2, Exclusive(false))
			So(k, ShouldEqual, 0)
		})
		Reset(func() {
			libInputOutput = mockInputOutput.prevInputOutput
		})
	})

}

func TestParsingArgs(t *testing.T) {
	Convey("With plugin lib parsing command line arguments", t, func() {
		mockInputOutput := newMockInputOutput(libInputOutput)
		libInputOutput = mockInputOutput
		Convey("invalid JSON will be rejected with an error", func() {
			mockInputOutput.mockArgs = strings.Fields("main {::invalid::JSON::}")
			_, err := getArgs()
			So(err, ShouldNotBeNil)
		})
		Convey("ListenPort should be properly parsed", func() {
			mockInputOutput.mockArgs = strings.Fields(`main {"ListenPort":"4414"}`)
			args, err := getArgs()
			So(err, ShouldBeNil)
			So(args.ListenPort, ShouldEqual, "4414")
		})
		Convey("PingTimeoutDuration should be properly parsed", func() {
			mockInputOutput.mockArgs = strings.Fields(`main {"PingTimeoutDuration":3141}`)
			args, err := getArgs()
			So(err, ShouldBeNil)
			So(args.PingTimeoutDuration, ShouldEqual, 3141)
		})
		Reset(func() {
			libInputOutput = mockInputOutput.prevInputOutput
		})
	})
}

func TestPassingPluginMeta(t *testing.T) {
	Convey("With plugin lib transferring plugin meta", t, func() {
		mockInputOutput := newMockInputOutput(libInputOutput)
		libInputOutput = mockInputOutput
		Convey("all meta arguments should be present in plugin response", func() {
			StartPublisher(newMockPublisher(), "mock-publisher-for-meta", 9, Exclusive(true), ConcurrencyCount(11), RoutingStrategy(StickyRouter), CacheTTL(305*time.Millisecond), rpcType(gRPC))
			var response preamble
			err := json.Unmarshal([]byte(mockInputOutput.output[0]), &response)
			if err != nil {
				panic(err)
			}
			var actMeta = response.Meta
			So(actMeta.CacheTTL, ShouldEqual, 305*time.Millisecond)
			So(actMeta.RoutingStrategy, ShouldEqual, StickyRouter)
			So(actMeta.RPCType, ShouldEqual, gRPC)
			So(actMeta.ConcurrencyCount, ShouldEqual, 11)
			So(actMeta.Exclusive, ShouldEqual, true)
			So(actMeta.Version, ShouldEqual, 9)
			So(actMeta.Name, ShouldEqual, "mock-publisher-for-meta")
			So(actMeta.Type, ShouldEqual, publisherType)
		})
		Reset(func() {
			libInputOutput = mockInputOutput.prevInputOutput
		})
	})
}

func TestApplySecurityArgsToMeta(t *testing.T) {
	Convey("With plugin lib accepting security args", t, func() {
		m := newMeta(processorType, "test-processor", 3)
		args := &Arg{}
		Convey("paths to certificate and key files should be properly passed to plugin meta", func() {
			args.CertPath = "some-cert-path"
			args.KeyPath = "some-key-path"
			args.TLSEnabled = true
			err := applySecurityArgsToMeta(m, args)
			So(err, ShouldBeNil)
			So(m.CertPath, ShouldEqual, "some-cert-path")
			So(m.KeyPath, ShouldEqual, "some-key-path")
			So(m.TLSEnabled, ShouldEqual, true)
		})
		Convey("paths to certificate and key files must not be set if TLS is not enabled", func() {
			args.TLSEnabled = false
			err := applySecurityArgsToMeta(m, args)
			So(err, ShouldBeNil)
			So(m.CertPath, ShouldEqual, "")
			So(m.KeyPath, ShouldEqual, "")
			So(m.TLSEnabled, ShouldEqual, false)
		})
		Convey("paths to certificate file should be allowed only when TLS is enabled with a flag", func() {
			args.CertPath = "some-cert-path"
			err := applySecurityArgsToMeta(m, args)
			So(err, ShouldNotBeNil)
		})
		Convey("paths to key file should be allowed only when TLS is enabled with a flag", func() {
			args.KeyPath = "some-key-path"
			err := applySecurityArgsToMeta(m, args)
			So(err, ShouldNotBeNil)
		})
		Convey("enabling TLS with a flag without certificate path is an error", func() {
			args.KeyPath = "some-key-path"
			args.TLSEnabled = true
			err := applySecurityArgsToMeta(m, args)
			So(err, ShouldNotBeNil)
		})
		Convey("enabling TLS with a flag without key path is an error", func() {
			args.CertPath = "some-cert-path"
			args.TLSEnabled = true
			err := applySecurityArgsToMeta(m, args)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMakeTLSConfig(t *testing.T) {
	Convey("Being security-aware", t, func() {
		tlsSetupInstance := &tlsServerDefaultSetup{}
		Convey("plugin lib should use TLS config requiring verified clients and specific cipher suites", func() {
			config := tlsSetupInstance.makeTLSConfig()
			So(config.ClientAuth, ShouldEqual, tls.RequireAndVerifyClientCert)
			So(config.PreferServerCipherSuites, ShouldEqual, true)
			So(config.CipherSuites, ShouldNotBeEmpty)
		})
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

type mockInputOutput struct {
	mockArgs        []string
	output          []string
	doReadOSArgs    func() []string
	doPrintOut      func(string)
	prevInputOutput osInputOutput
}

func (f *mockInputOutput) readOSArgs() []string {
	return f.doReadOSArgs()
}

func (f *mockInputOutput) printOut(data string) {
	f.doPrintOut(data)
}

func newMockInputOutput(prevInputOutput osInputOutput) *mockInputOutput {
	mock := mockInputOutput{mockArgs: strings.Fields("mock {}")}
	mock.prevInputOutput = prevInputOutput
	mock.doPrintOut = func(data string) {
		mock.output = append(mock.output, data)
	}
	mock.doReadOSArgs = func() []string {
		return mock.mockArgs
	}
	return &mock
}

func (mp *mockPlugin) GetConfigPolicy() (ConfigPolicy, error) {
	if mp.err != nil {
		return ConfigPolicy{}, errors.New("error")
	}
	cp := NewConfigPolicy()

	cp.AddNewBoolRule([]string{"log"}, "logLevel", true, SetDefaultBool(true))
	cp.AddNewBoolRule([]string{"cache"}, "cacheTime", true, SetDefaultBool(false))

	cp.AddNewFloatRule([]string{"float"}, "low", true, SetDefaultFloat(32.1))
	cp.AddNewFloatRule([]string{"cache"}, "high", true, SetDefaultFloat(2399.58))

	cp.AddNewIntRule([]string{"xyz"}, "logLevel", false, SetDefaultInt(30))
	cp.AddNewIntRule([]string{"abc"}, "cacheTime", true, SetDefaultInt(50))

	cp.AddNewStringRule([]string{"log"}, "logLevel", true, SetDefaultString("123"))
	cp.AddNewStringRule([]string{"cache"}, "cacheTime", true, SetDefaultString("tyty"))

	return (*cp), nil
}

type mockStreamer struct {
	mockPlugin
	err                error
	maxBuffer          int64
	maxCollectDuration time.Duration
	inMetric           chan []Metric
	outMetric          chan []Metric
	action             func(chan []Metric, time.Duration, []Metric)
}

func newMockStreamer() *mockStreamer {
	return &mockStreamer{}
}

func newMockErrStreamer() *mockStreamer {
	return &mockStreamer{err: errors.New("empty")}
}

func newMockStreamerStream(action func(chan []Metric, time.Duration, []Metric)) *mockStreamer {
	return &mockStreamer{action: action}
}

func (mc *mockStreamer) doAction(t time.Duration, mts []Metric) {
	go mc.action(mc.outMetric, t, mts)
}
func (mc *mockStreamer) GetMetricTypes(cfg Config) ([]Metric, error) {
	if mc.err != nil {
		return nil, errors.New("error")
	}

	mts := []Metric{}
	for _, v := range metricMap {
		mts = append(mts, v)
	}
	return mts, nil
}

func (mc *mockStreamer) StreamMetrics(i chan []Metric, o chan []Metric, _ chan string) error {

	if mc.err != nil {
		return errors.New("error")
	}
	mc.inMetric = i
	mc.outMetric = o
	return nil
}

type mockCollector struct {
	mockPlugin
	err              error
	doCollectMetrics func([]Metric) ([]Metric, error)
	doGetMetricTypes func(Config) ([]Metric, error)
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
	if mc.doGetMetricTypes != nil {
		return mc.doGetMetricTypes(cfg)
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
	if mc.doCollectMetrics != nil {
		return mc.doCollectMetrics(mts)
	}
	return mts, nil
}

type mockProcessor struct {
	mockPlugin
	err       error
	doProcess func([]Metric, Config) ([]Metric, error)
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
	if mp.doProcess != nil {
		return mp.doProcess(mts, cfg)
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
	err       error
	doPublish func([]Metric, Config) error
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
	if mpb.doPublish != nil {
		return mpb.doPublish(mts, cfg)
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
