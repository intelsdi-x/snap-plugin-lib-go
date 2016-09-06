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
	"net"
	"sync"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

type test struct {
	t *testing.T

	ctx    context.Context
	cancel context.CancelFunc

	srv     *grpc.Server
	srvAddr string

	cc *grpc.ClientConn

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

	switch pt {
	case collectorType:
		collectProxy := &collectorProxy{
			pluginProxy: *newPluginProxy(&mockPlugin{}),
			plugin:      newMockCollector(),
		}
		rpc.RegisterCollectorServer(s, collectProxy)
		tt.halt = collectProxy.halt
	case processorType:
		processorProxy := &processorProxy{
			pluginProxy: *newPluginProxy(&mockPlugin{}),
			plugin:      newMockProcessor(),
		}
		rpc.RegisterProcessorServer(s, processorProxy)
		tt.halt = processorProxy.halt
	case publisherType:
		publisherProxy := &publisherProxy{
			pluginProxy: *newPluginProxy(&mockPlugin{}),
			plugin:      newMockPublisher(),
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
