// +build medium

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
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	// Filter out go test flags
	getOSArgs = func() []string {
		args := []string{}
		for _, v := range os.Args {
			if !strings.HasPrefix(v, "-test") {
				args = append(args, v)
			}
		}
		return args
	}
}

const GrpcTimeoutDefault = 2 * time.Second

var (
	grpcTestMetricTypesMap = map[string]interface{}{
		"a/b": int64(0), "a/b/c": int64(0), "d": int64(0),
	}

	grpcTestCollectMetricsMap = map[string]interface{}{
		"a/b": int64(2), "a/b/c": int64(3), "d": int64(1),
	}

	grpcTestProcessMetricsMap = map[string]interface{}{
		"a/b": int64(-2), "a/b/c": int64(-3), "d": int64(-1),
	}

	grpcTestPublishMetricsMap = map[string]interface{}{
		"a/string": "test", "a/float32": float32(-1.2345),
		"a/float64": float64(-2.3456), "an/int32": int32(-345678),
		"an/int64": int64(-456789), "an/uint32": uint32(567890),
		"an/uint64": uint64(678901), "a/[]byte": []byte{78, 90, 12},
		"a/bool": false, "a/nil": nil,
	}

	grpcTestPublishConfigMapInput = map[string]Config{
		"proper/string":  {"string": "ok"},
		"proper/bool":    {"bool": true},
		"proper/float":   {"float": float64(-1.2345)},
		"proper/int":     {"int": int64(-234567)},
		"missing/string": {},
		"missing/bool":   {},
		"missing/float":  {},
		"missing/int":    {},
		"invalid/string": {"string": false},
		"invalid/bool":   {"bool": ""},
		"invalid/float":  {"float": ""},
		"invalid/int":    {"int": ""},
	}

	grpcTestPublishConfigMapOutput = map[string]string{
		"proper/string":  "ok",
		"proper/bool":    "true",
		"proper/float":   "-1.2345",
		"proper/int":     "-234567",
		"missing/string": "missing",
		"missing/bool":   "missing",
		"missing/float":  "missing",
		"missing/int":    "missing",
		"invalid/string": "invalid",
		"invalid/bool":   "invalid",
		"invalid/float":  "invalid",
		"invalid/int":    "invalid",
	}
)

type test struct {
	t *testing.T

	ctx    context.Context
	cancel context.CancelFunc

	srv     *grpc.Server
	srvAddr string

	cc            *grpc.ClientConn
	pluginBuilder func(pt pluginType) Plugin

	halt chan struct{}
}

func (tt *test) tearDown() {
	if tt.cancel != nil {
		tt.cancel()
		tt.cancel = nil
	}

	if tt.cc != nil {
		tt.cc.Close()
		tt.cc = nil
	}

	if tt.srv != nil {
		tt.srv.Stop()
	}
}

func newTest(t *testing.T) *test {
	tt := &test{
		t: t,
		pluginBuilder: func(pt pluginType) Plugin {
			switch pt {
			case collectorType:
				return newMockCollector()
			case processorType:
				return newMockProcessor()
			case publisherType:
				return newMockPublisher()
			}
			panic(fmt.Errorf("unsupported plugin type: %v", pt))
		},
	}

	tt.ctx, tt.cancel = context.WithCancel(context.Background())

	return tt
}

func (tt *test) startServer(pt pluginType) {
	tt.t.Logf("Starting server...")
	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(2)}

	la := "localhost:0"
	lis, err := net.Listen("tcp", la)
	if err != nil {
		tt.t.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(sopts...)
	tt.srv = s
	pluginInst := tt.pluginBuilder(pt)
	switch pt {
	case collectorType:
		collectProxy := &collectorProxy{
			pluginProxy: *newPluginProxy(pluginInst),
			plugin:      pluginInst.(Collector),
		}
		rpc.RegisterCollectorServer(s, collectProxy)
		tt.halt = collectProxy.halt
	case processorType:
		processorProxy := &processorProxy{
			pluginProxy: *newPluginProxy(pluginInst),
			plugin:      pluginInst.(Processor),
		}
		rpc.RegisterProcessorServer(s, processorProxy)
		tt.halt = processorProxy.halt
	case publisherType:
		publisherProxy := &publisherProxy{
			pluginProxy: *newPluginProxy(pluginInst),
			plugin:      pluginInst.(Publisher),
		}
		rpc.RegisterPublisherServer(s, publisherProxy)
		tt.halt = publisherProxy.halt
	}

	addr := la
	_, port, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		tt.t.Fatalf("Failed to parse listener address: %v", err)
	}
	addr = "localhost:" + port

	go s.Serve(lis)
	tt.srvAddr = addr
}

func (tt *test) clientConn() *grpc.ClientConn {
	if tt.cc != nil {
		return tt.cc
	}

	var err error
	tt.cc, err = grpc.Dial(tt.srvAddr, grpc.WithInsecure())
	if err != nil {
		tt.t.Fatalf("Dial(%q) = %v", tt.srvAddr, err)
	}
	return tt.cc
}

func TestCollectorGrpc(t *testing.T) {
	tt := newTest(t)
	tt.startServer(collectorType)
	defer tt.tearDown()
	testCollectorGrpcBackend(t, tt)
}

func testCollectorGrpcBackend(t *testing.T, tt *test) {
	cc := tt.clientConn()
	tc := rpc.NewCollectorClient(cc)

	Convey("Test Collector Client", t, func() {
		Convey("Test GetConfigPolicy", func() {
			reply, err := tc.GetConfigPolicy(tt.ctx, &rpc.Empty{})
			So(err, ShouldBeNil)
			So(reply, ShouldNotBeNil)
		})
		Convey("Test GetMetricTypes", func() {
			reply, err := tc.GetMetricTypes(tt.ctx, &rpc.GetMetricTypesArg{})
			So(err, ShouldBeNil)
			So(reply, ShouldNotBeNil)
		})
		Convey("Test CollectMetrics", func() {
			reply, err := tc.CollectMetrics(tt.ctx, &rpc.MetricsArg{})
			So(err, ShouldBeNil)
			So(reply, ShouldNotBeNil)
		})

		var err error
		Convey("Test Collector Ping", func() {
			if _, err := tc.Ping(tt.ctx, &rpc.Empty{}); err != nil {
				tt.t.Fatalf("failed to ping %v", err)
			}
			So(err, ShouldBeNil)
		})

	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		Convey("Test Collector Client Kill", t, func() {
			defer wg.Done()
			var err error
			if _, err := tc.Kill(tt.ctx, &rpc.KillArg{Reason: "test"}); err != nil {
				tt.t.Fatalf("failed to kill %v", err)
			}
			So(err, ShouldBeNil)
		})
	}()
	<-tt.halt
	wg.Wait()
}

func TestPublisherGrpc(t *testing.T) {
	tt := newTest(t)
	tt.startServer(publisherType)
	defer tt.tearDown()
	testPublisherGrpcBackend(t, tt)
}

func testPublisherGrpcBackend(t *testing.T, tt *test) {
	cc := tt.clientConn()
	tc := rpc.NewPublisherClient(cc)

	Convey("Test Publisher Client Publish", t, func() {
		Convey("Test Publish", func() {
			reply, err := tc.Publish(tt.ctx, &rpc.PubProcArg{})
			So(err, ShouldBeNil)
			So(reply, ShouldNotBeNil)
		})

		var err error
		Convey("Test GetConfigPolicy", func() {
			if _, err = tc.GetConfigPolicy(tt.ctx, &rpc.Empty{}); err != nil {
				tt.t.Fatalf("failed to get config policy %v", err)
			}
			So(err, ShouldBeNil)
		})

		Convey("Test Publisher Ping", func() {
			if _, err = tc.Ping(tt.ctx, &rpc.Empty{}); err != nil {
				tt.t.Fatalf("failed to ping %v", err)
			}
			So(err, ShouldBeNil)
		})
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		Convey("Test Publisher Client Kill", t, func() {
			defer wg.Done()
			var err error
			if _, err := tc.Kill(tt.ctx, &rpc.KillArg{Reason: "test"}); err != nil {
				tt.t.Fatalf("failed to kill %v", err)
			}
			So(err, ShouldBeNil)
		})
	}()
	<-tt.halt
	wg.Wait()
}

func TestProcessorGrpc(t *testing.T) {
	tt := newTest(t)
	tt.startServer(processorType)
	defer tt.tearDown()
	testProcessorGrpcBackend(t, tt)
}

func testProcessorGrpcBackend(t *testing.T, tt *test) {
	cc := tt.clientConn()
	tc := rpc.NewProcessorClient(cc)

	Convey("Test Processor Client", t, func() {

		Convey("Test Process", func() {
			reply, err := tc.Process(tt.ctx, &rpc.PubProcArg{})
			So(err, ShouldBeNil)
			So(reply, ShouldNotBeNil)
		})

		var err error
		Convey("Test GetConfigPolicy", func() {
			if _, err := tc.GetConfigPolicy(tt.ctx, &rpc.Empty{}); err != nil {
				tt.t.Fatalf("failed to get config policy %v", err)
			}
			So(err, ShouldBeNil)
		})

		Convey("Test Processor Ping", func() {
			if _, err := tc.Ping(tt.ctx, &rpc.Empty{}); err != nil {
				tt.t.Fatalf("failed to ping %v", err)
			}
			So(err, ShouldBeNil)
		})
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var err error
		Convey("Test Processor Client Kill", t, func() {
			defer wg.Done()
			if _, err := tc.Kill(tt.ctx, &rpc.KillArg{Reason: "test"}); err != nil {
				tt.t.Fatalf("failed to kill %v", err)
			}
			So(err, ShouldBeNil)
		})
	}()
	<-tt.halt
	wg.Wait()
}

func TestCollectorFlow(t *testing.T) {
	var mockCollector *mockCollector
	Convey("Having Collector communicate over GRPC", t, func() {
		tt := newTest(t)
		tt.pluginBuilder = func(pt pluginType) Plugin {
			if pt != collectorType {
				panic(fmt.Errorf("unsupported plugin type: %v, expected: %v", pt, collectorType))
			}
			return mockCollector
		}
		mockCollector = newMockCollector()
		Convey("errors from GetMetricTypes should be propagated to caller", func() {
			mockCollector.doGetMetricTypes = func(_ Config) ([]Metric, error) {
				return nil, fmt.Errorf("doGetMetricTypes")
			}
			tt.startServer(collectorType)
			cc := tt.clientConn()
			tc := rpc.NewCollectorClient(cc)
			_, err := tc.GetMetricTypes(tt.ctx, &rpc.GetMetricTypesArg{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "doGetMetricTypes")
		})
		Convey("errors from CollectMetrics should be propagated to caller", func() {
			mockCollector.doCollectMetrics = func([]Metric) ([]Metric, error) {
				return nil, fmt.Errorf("doCollectMetrics")
			}
			tt.startServer(collectorType)
			cc := tt.clientConn()
			tc := rpc.NewCollectorClient(cc)
			_, err := tc.CollectMetrics(tt.ctx, &rpc.MetricsArg{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "doCollectMetrics")
		})
		Convey("metrics from GetMetricTypes should be forwarded to caller", func() {
			mockCollector.doGetMetricTypes = func(_ Config) (dst []Metric, err error) {
				for k := range grpcTestMetricTypesMap {
					ns := NewNamespace(strings.Split(k, "/")...)
					dst = append(dst, Metric{Namespace: ns, Data: int64(0)})
				}
				return dst, nil
			}
			tt.startServer(collectorType)
			cc := tt.clientConn()
			tc := rpc.NewCollectorClient(cc)
			reply, err := tc.GetMetricTypes(tt.ctx, &rpc.GetMetricTypesArg{})
			So(err, ShouldBeNil)
			mts := metricsFromRpc(reply.Metrics)
			actMap := metricsToMap(mts)
			So(actMap, ShouldResemble, grpcTestMetricTypesMap)
		})
		Convey("metrics from CollectMetrics should be forwarded to caller", func() {
			mockCollector.doCollectMetrics = func(mts []Metric) (dst []Metric, err error) {
				for _, m := range mts {
					n := Metric{}
					n.Data = len(m.Namespace.Strings())
					n.Namespace = CopyNamespace(m.Namespace)
					dst = append(dst, n)
				}
				return dst, nil
			}
			tt.startServer(collectorType)
			cc := tt.clientConn()
			tc := rpc.NewCollectorClient(cc)
			metricTypes := metricsFromMap(grpcTestMetricTypesMap)
			reply, err := tc.CollectMetrics(tt.ctx, &rpc.MetricsArg{Metrics: metricsToRpc(metricTypes)})
			So(err, ShouldBeNil)
			mts := metricsFromRpc(reply.Metrics)
			actMap := metricsToMap(mts)
			So(actMap, ShouldResemble, grpcTestCollectMetricsMap)
		})
		Reset(func() {
			tt.tearDown()
		})
	})
}

func TestProcessorFlow(t *testing.T) {
	var mockProcessor *mockProcessor
	Convey("Having Processor communicate over GRPC", t, func() {
		tt := newTest(t)
		tt.pluginBuilder = func(pt pluginType) Plugin {
			if pt != processorType {
				panic(fmt.Errorf("unsupported plugin type: %v, expected: %v", pt, processorType))
			}
			return mockProcessor
		}
		mockProcessor = newMockProcessor()
		Convey("errors from Process should be propagated to caller", func() {
			mockProcessor.doProcess = func(_ []Metric, _ Config) ([]Metric, error) {
				return nil, fmt.Errorf("doProcess")
			}
			tt.startServer(processorType)
			cc := tt.clientConn()
			tc := rpc.NewProcessorClient(cc)
			_, err := tc.Process(tt.ctx, &rpc.PubProcArg{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "doProcess")
		})
		Convey("results from Process should be propagated to caller", func() {
			mockProcessor.doProcess = func(mts []Metric, _ Config) (dst []Metric, _ error) {
				for _, m := range mts {
					m.Data = -1 * (m.Data.(int64))
					dst = append(dst, m)
				}
				return dst, nil
			}
			tt.startServer(processorType)
			cc := tt.clientConn()
			tc := rpc.NewProcessorClient(cc)
			collMts := metricsFromMap(grpcTestCollectMetricsMap)
			reply, err := tc.Process(tt.ctx, &rpc.PubProcArg{Metrics: metricsToRpc(collMts)})
			So(err, ShouldBeNil)
			mts := metricsFromRpc(reply.Metrics)
			actMap := metricsToMap(mts)
			So(actMap, ShouldResemble, grpcTestProcessMetricsMap)
		})
		Reset(func() {
			tt.tearDown()
		})
	})
}

func TestPublisherFlow(t *testing.T) {
	var mockPublisher *mockPublisher
	Convey("Having Publisher communicate over GRPC", t, func() {
		tt := newTest(t)
		tt.pluginBuilder = func(pt pluginType) Plugin {
			if pt != publisherType {
				panic(fmt.Errorf("unsupported plugin type: %v, expected: %v", pt, publisherType))
			}
			return mockPublisher
		}
		mockPublisher = newMockPublisher()
		Convey("errors from Publish should be propagated to caller", func() {
			mockPublisher.doPublish = func(_ []Metric, _ Config) error {
				return fmt.Errorf("doPublish")
			}
			tt.startServer(publisherType)
			cc := tt.clientConn()
			tc := rpc.NewPublisherClient(cc)
			reply, _ := tc.Publish(tt.ctx, &rpc.PubProcArg{})
			So(reply.Error, ShouldContainSubstring, "doPublish")
		})
		Convey("metrics passed to Publish should be delivered unchanged", func() {
			reportMux := sync.Mutex{}
			var report []Metric
			mockPublisher.doPublish = func(mts []Metric, _ Config) error {
				reportMux.Lock()
				defer reportMux.Unlock()
				report = append(report, mts...)
				return nil
			}
			tt.startServer(publisherType)
			cc := tt.clientConn()
			tc := rpc.NewPublisherClient(cc)
			inMts := metricsFromMap(grpcTestPublishMetricsMap)
			_, err := tc.Publish(tt.ctx, &rpc.PubProcArg{Metrics: metricsToRpc(inMts)})
			So(err, ShouldBeNil)
			reportMux.Lock()
			defer reportMux.Unlock()
			actMap := metricsToMap(report)
			So(actMap, ShouldResemble, grpcTestPublishMetricsMap)
		})
		Convey("config passed to Publish should be delivered unchanged", func() {
			reportMux := sync.Mutex{}
			report := map[string]string{}
			getConfigAsString := func(c Config, k string) string {
				var v interface{}
				var err error
				switch k {
				case "string":
					v, err = c.GetString(k)
				case "bool":
					v, err = c.GetBool(k)
				case "float":
					v, err = c.GetFloat(k)
				case "int":
					v, err = c.GetInt(k)
				}
				if err == ErrConfigNotFound {
					return "missing"
				}
				if err == ErrNotABool || err == ErrNotAString ||
					err == ErrNotAFloat || err == ErrNotAnInt {
					return "invalid"
				}
				return fmt.Sprintf("%v", v)
			}
			mockPublisher.doPublish = func(mts []Metric, _ Config) error {
				reportMux.Lock()
				defer reportMux.Unlock()
				for _, m := range mts {
					k := strings.Join(m.Namespace.Strings(), "/")
					v := getConfigAsString(m.Config, m.Namespace[len(m.Namespace)-1].Value)
					report[k] = v
				}
				return nil
			}
			tt.startServer(publisherType)
			cc := tt.clientConn()
			tc := rpc.NewPublisherClient(cc)
			inMts := metricsWithConfigFromMap(grpcTestPublishConfigMapInput)
			_, err := tc.Publish(tt.ctx, &rpc.PubProcArg{Metrics: metricsToRpc(inMts)})
			So(err, ShouldBeNil)
			reportMux.Lock()
			defer reportMux.Unlock()
			So(report, ShouldResemble, grpcTestPublishConfigMapOutput)
		})
		Reset(func() {
			tt.tearDown()
		})
	})
}

func metricsFromMap(src map[string]interface{}) (dst []Metric) {
	for k, v := range src {
		ns := NewNamespace()
		ns = ns.AddStaticElements(strings.Split(k, "/")...)
		m := Metric{
			Namespace: ns,
			Data:      v,
		}
		dst = append(dst, m)
	}
	return dst
}

func metricsWithConfigFromMap(src map[string]Config) (dst []Metric) {
	for k, cfg := range src {
		ns := NewNamespace()
		for _, k := range strings.Split(k, "/") {
			ns = ns.AddStaticElement(k)
		}
		m := Metric{
			Namespace: ns,
			Data:      nil,
			Config:    cfg,
		}
		dst = append(dst, m)
	}
	return dst
}

func metricsToMap(src []Metric) (dst map[string]interface{}) {
	dst = map[string]interface{}{}
	for _, m := range src {
		k := strings.Join(m.Namespace.Strings(), "/")
		dst[k] = m.Data
	}
	return dst
}

func metricsFromRpc(src []*rpc.Metric) (dst []Metric) {
	for _, m := range src {
		dst = append(dst, fromProtoMetric(m))
	}
	return dst
}

func metricsToRpc(src []Metric) (dst []*rpc.Metric) {
	for _, m := range src {
		if r, err := toProtoMetric(m); err != nil {
			panic(err)
		} else {
			r.Config = configToRpc(m.Config)
			dst = append(dst, r)
		}
	}
	return dst
}

func configToRpc(cfg Config) (rpcConfig *rpc.ConfigMap) {
	if len(cfg) == 0 {
		return nil
	}
	rpcConfig = &rpc.ConfigMap{
		IntMap:    map[string]int64{},
		StringMap: map[string]string{},
		FloatMap:  map[string]float64{},
		BoolMap:   map[string]bool{},
	}
	for k, v := range cfg {
		t := reflect.TypeOf(v).Kind()
		switch t {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
			rpcConfig.IntMap[k] = v.(int64)
		case reflect.Float32, reflect.Float64:
			rpcConfig.FloatMap[k] = v.(float64)
		case reflect.String:
			rpcConfig.StringMap[k] = v.(string)
		case reflect.Bool:
			rpcConfig.BoolMap[k] = v.(bool)
		}
	}
	return rpcConfig
}
