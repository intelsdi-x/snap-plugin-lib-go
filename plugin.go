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
	"encoding/json"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/intelsdi-x/snap-plugin-lib-go/rpc"
)

// Plugin is the base plugin type. All plugins must implement GetConfigPolicy.
type Plugin interface {
	GetConfigPolicy() (ConfigPolicy, error)
}

// Collector is a plugin which is the source of new data in the Snap pipeline.
type Collector interface {
	Plugin

	GetMetricTypes(Config) ([]Metric, error)
	CollectMetrics([]Metric) ([]Metric, error)
}

// Processor is a plugin which filters, agregates, or decorates data in the
// Snap pipeline.
type Processor interface {
	Plugin

	Process([]Metric) ([]Metric, error)
}

// Publisher is a sink in the Snap pipeline.  It publishes data into another
// System, completing a Workflow path.
type Publisher interface {
	Plugin

	Publisher([]Metric) ([]Metric, error)
}

// StartCollector is given a Collector implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartCollector(plugin Collector, name string, version int, opts ...MetaOpt) int {
	m := newMeta(collectorType, name, version, opts...)
	server := grpc.NewServer()
	// TODO(danielscottt) SSL
	proxy := &collectorProxy{
		plugin:      plugin,
		pluginProxy: *newPluginProxy(plugin),
	}
	rpc.RegisterCollectorServer(server, proxy)
	return startPlugin(server, m, proxy.pluginProxy.halt)
}

// StartProcessor is given a Processor implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartProcessor(plugin Processor, name string, version int, opts ...MetaOpt) int {
	m := newMeta(processorType, name, version, opts...)
	server := grpc.NewServer()
	// TODO(danielscottt) SSL
	proxy := &processorProxy{
		plugin:      plugin,
		pluginProxy: *newPluginProxy(plugin),
	}
	rpc.RegisterProcessorServer(server, proxy)
	return startPlugin(server, m, proxy.pluginProxy.halt)
}

// StartPublisher is given a Publisher implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartPublisher(plugin Publisher, name string, version int, opts ...MetaOpt) int {
	m := newMeta(publisherType, name, version, opts...)
	server := grpc.NewServer()
	// TODO(danielscottt) SSL
	proxy := &publisherProxy{
		plugin:      plugin,
		pluginProxy: *newPluginProxy(plugin),
	}
	rpc.RegisterPublisherServer(server, proxy)
	return startPlugin(server, m, proxy.pluginProxy.halt)
}

type server interface {
	Serve(net.Listener) error
}

func startPlugin(srv server, m meta, halt <-chan struct{}) int {
	//TODO(danielscottt): listen port
	l, err := net.Listen("tcp", "127.0.0.1:9998")
	if err != nil {
		// TODO(danielscottt): logging
		panic(err)
	}
	go func() {
		err := srv.Serve(l)
		if err != nil {
			panic(err) //TODO(danielscottt): panic?
		}
	}()
	// TODO(danielscottt): Resp generation
	// Output response to stdout
	metaJson, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(metaJson))
	// TODO(danielscottt): heartbeats
	// TODO(danielscottt): exit code
	<-halt
	return 0
}
