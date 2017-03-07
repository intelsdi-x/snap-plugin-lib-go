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
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	Process([]Metric, Config) ([]Metric, error)
}

// Publisher is a sink in the Snap pipeline.  It publishes data into another
// System, completing a Workflow path.
type Publisher interface {
	Plugin

	Publish([]Metric, Config) error
}

var App *cli.App

// StreamCollector is a Collector that can send back metrics on it's own
// defined interval (within configurable limits). These limits are set by the
// SetMaxBuffer and SetMaxCollectionDuration funcs.
type StreamCollector interface {
	Plugin

	// StreamMetrics allows the plugin to send/receive metrics on a channel
	// Arguments are (in order):
	//
	// A channel for metrics into the plugin from Snap -- which
	// are the metric types snap is requesting the plugin to collect.
	//
	// A channel for metrics from the plugin to Snap -- the actual
	// collected metrics from the plugin.
	//
	// A channel for error strings that the library will report to snap
	// as task errors.
	StreamMetrics(chan []Metric, chan []Metric, chan string) error
	GetMetricTypes(Config) ([]Metric, error)
}

// tlsServerSetup offers functions supporting TLS server setup
type tlsServerSetup interface {
	// makeTLSConfig delivers TLS config suitable to use for plugins, excluding
	// setup of certificates (either subject or root CA certificates).
	makeTLSConfig() *tls.Config
	// readRootCAs is a function that delivers root CA certificates for the purpose
	// of TLS initialization
	readRootCAs() (*x509.CertPool, error)
	// updateServerOptions configures any additional options for GRPC server
	updateServerOptions(options ...grpc.ServerOption) []grpc.ServerOption
}

// osInputOutput supports interactions with OS for the plugin lib
type osInputOutput interface {
	// readOSArgs gets command line arguments passed to application
	readOSArgs() []string
	// printOut outputs given data to application standard output
	printOut(data string)
}

// standardInputOutput delivers standard implementation for OS
// interactions
type standardInputOutput struct {
}

// tlsServerDefaultSetup provides default implementation for TLS setup routines
type tlsServerDefaultSetup struct {
}

// tlsSetup holds TLS setup utility for plugin lib
var tlsSetup tlsServerSetup = tlsServerDefaultSetup{}

// libInputOutput holds utility used for OS interactions
var libInputOutput osInputOutput = standardInputOutput{}

// makeTLSConfig provides TLS configuraton template for plugins, setting
// required verification of client cert and preferred server suites.
func (ts tlsServerDefaultSetup) makeTLSConfig() *tls.Config {
	config := tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}
	return &config
}

// readRootCAs delivers a standard source of root CAs from system
func (ts tlsServerDefaultSetup) readRootCAs() (*x509.CertPool, error) {
	return x509.SystemCertPool()
}

// updateServerOptions a standard implementation delivers no additional options
func (ts tlsServerDefaultSetup) updateServerOptions(options ...grpc.ServerOption) []grpc.ServerOption {
	return options
}

// readOSArgs implementation that returns application args passed by OS
func (io standardInputOutput) readOSArgs() []string {
	return os.Args
}

// printOut implementation that emits data into standard output
func (io standardInputOutput) printOut(data string) {
	fmt.Println(data)
}

// makeGRPCCredentials delivers credentials object suitable for setting up gRPC
// server, with TLS optionally turned on.
func makeGRPCCredentials(m *meta) (creds credentials.TransportCredentials, err error) {
	var config *tls.Config
	if !m.TLSEnabled {
		config = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		cert, err := tls.LoadX509KeyPair(m.CertPath, m.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to setup credentials for plugin - loading key pair failed: %v", err.Error())
		}
		config = tlsSetup.makeTLSConfig()
		config.Certificates = []tls.Certificate{cert}
		if config.RootCAs, err = tlsSetup.readRootCAs(); err != nil {
			return nil, fmt.Errorf("unable to read root CAs: %v", err.Error())
		}
	}
	creds = credentials.NewTLS(config)
	return creds, nil
}

// applySecurityArgsToMeta validates plugin runtime arguments from OS, focusing on
// TLS functionality.
func applySecurityArgsToMeta(m *meta, args *Arg) error {
	if !args.TLSEnabled {
		if args.CertPath != "" || args.KeyPath != "" {
			return fmt.Errorf("excessive arguments given - CertPath and KeyPath are unused with TLS not enabled")
		}
		return nil
	}
	if args.CertPath == "" || args.KeyPath == "" {
		return fmt.Errorf("failed to enable TLS for plugin - need both CertPath and KeyPath")
	}
	m.CertPath = args.CertPath
	m.KeyPath = args.KeyPath
	m.TLSEnabled = true
	return nil
}

// buildGRPCServer configures and builds GRPC server ready to server a plugin
// instance
func buildGRPCServer(typeOfPlugin pluginType, name string, version int, opts ...MetaOpt) (server *grpc.Server, m *meta, err error) {
	args, err := getArgs()
	if err != nil {
		fmt.Println("ERROR 1")
		return nil, nil, err
	}
	m = newMeta(typeOfPlugin, name, version, opts...)

	if err := applySecurityArgsToMeta(m, args); err != nil {
		fmt.Println("ERROR 2")
		return nil, nil, err
	}
	creds, err := makeGRPCCredentials(m)
	if err != nil {
		fmt.Println("ERROR 3")
		return nil, nil, err
	}
	if m.TLSEnabled {
		server = grpc.NewServer(tlsSetup.updateServerOptions(grpc.Creds(creds))...)
	} else {
		server = grpc.NewServer(tlsSetup.updateServerOptions()...)
	}
	return server, m, nil
}

// StartCollector is given a Collector implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartCollector(plugin Collector, name string, version int, opts ...MetaOpt) int {
	server, m, err := buildGRPCServer(collectorType, name, version, opts...)
	if err != nil {
		panic(err)
	}
	proxy := &collectorProxy{
		plugin:      plugin,
		pluginProxy: *newPluginProxy(plugin),
	}
	rpc.RegisterCollectorServer(server, proxy)
	return startPlugin(server, *m, &proxy.pluginProxy)
}

// StartProcessor is given a Processor implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartProcessor(plugin Processor, name string, version int, opts ...MetaOpt) int {
	server, m, err := buildGRPCServer(processorType, name, version, opts...)
	if err != nil {
		panic(err)
	}
	proxy := &processorProxy{
		plugin:      plugin,
		pluginProxy: *newPluginProxy(plugin),
	}
	rpc.RegisterProcessorServer(server, proxy)
	return startPlugin(server, *m, &proxy.pluginProxy)
}

// StartPublisher is given a Publisher implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartPublisher(plugin Publisher, name string, version int, opts ...MetaOpt) int {
	server, m, err := buildGRPCServer(publisherType, name, version, opts...)
	if err != nil {
		panic(err)
	}
	proxy := &publisherProxy{
		plugin:      plugin,
		pluginProxy: *newPluginProxy(plugin),
	}
	rpc.RegisterPublisherServer(server, proxy)
	return startPlugin(server, *m, &proxy.pluginProxy)
}

// StartStreamCollector is given a StreamCollector implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartStreamCollector(plugin StreamCollector, name string, version int, opts ...MetaOpt) int {
	opts = append(opts, rpcType(gRPCStream))
	server, m, err := buildGRPCServer(collectorType, name, version, opts...)
	if err != nil {
		panic(err)
	}
	proxy := &StreamProxy{
		plugin:      plugin,
		pluginProxy: *newPluginProxy(plugin),
	}
	rpc.RegisterStreamCollectorServer(server, proxy)
	return startPlugin(server, *m, &proxy.pluginProxy)
}

type server interface {
	Serve(net.Listener) error
}

type preamble struct {
	Meta          meta
	ListenAddress string
	PprofAddress  string
	Type          pluginType
	State         int
	ErrorMessage  string
}

func startPlugin(srv server, m meta, p *pluginProxy) int {
	app := cli.NewApp()
	app.Name = m.Name
	app.Version = strconv.Itoa(m.Version)
	app.Usage = "A Snap " + getPluginType(m.Type) + " plugin"
	//TODO: optional set description field

	app.Flags = []cli.Flag{flConfig, flVerbose, flPort, flPingTimeout, flPprof}

	app.Action = func(c *cli.Context) error {
		if c.NArg() > 0 {
			printPreamble(srv, &m, p)
		} else { //implies run diagnostics
			var c Config
			if config != "" {
				fmt.Println("TODO: apply config")
				fmt.Println("TODO: parse into a Config type")
				//apply config ?
				//flag will default config to "" if nothing passed in

				c = Config{
					"user":       "john",
					"someint":    1234,
					"somefloat":  3.14,
					"somebool":   true,
					"user2":      "jane",
					"someint2":   4321,
					"somefloat2": 4.13,
				}
			}

			switch p.plugin.(type) {
			case Collector:
				showDiagnostics(m, p, c)
			case Processor:
				fmt.Println("Diagnostics not currently available for processor plugins.")
			case Publisher:
				fmt.Println("Diagnostics not currently available for publisher plugins.")
			}
		}
		if Pprof {
			return getPort()
		}

		return nil
	}
	app.Run(os.Args)

	return 0
}

func printPreamble(srv server, m *meta, p *pluginProxy) error {
	l, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		panic("Unable to get open port")
	}
	l.Close()

	addr := fmt.Sprintf("127.0.0.1:%v", l.Addr().(*net.TCPAddr).Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		// TODO(danielscottt): logging
		panic(err)
	}
	go func() {
		err := srv.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	resp := preamble{
		Meta:          *m,
		ListenAddress: addr,
		Type:          m.Type,
		PprofAddress:  pprofPort,
		State:         0, // Hardcode success since panics on err
	}
	preambleJSON, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	libInputOutput.printOut(string(preambleJSON))
	go p.HeartbeatWatch()
	// TODO(danielscottt): exit code
	<-p.halt

	return nil
}

// GetPluginType converts a pluginType to a string
// describing what type of plugin this is
func getPluginType(plType pluginType) string {
	switch plType {
	case collectorType:
		return "collector"

	case publisherType:
		return "publisher"

	case processorType:
		return "processor"
	}
	return ""
}

func showDiagnostics(m meta, p *pluginProxy, c Config) error {
	defer timeTrack(time.Now(), "showDiagnostics")
	if verbose {
		fmt.Println("Show VERBOSE diagnostics!")
	} else {
		fmt.Println("SHOW DIAGNOSTICS!")
		met, err := printMetricTypes(p, c)
		if err != nil {
			return err
		}

		err = printCollectMetrics(p, met)
		if err != nil {
			return err
		}

	}
	return nil
}

func printMetricTypes(p *pluginProxy, conf Config) ([]Metric, error) {
	defer timeTrack(time.Now(), "printMetricTypes")
	met, err := p.plugin.(Collector).GetMetricTypes(conf)
	if err != nil {
		return nil, err
	}
	fmt.Println("Metric Types include: ")
	for _, j := range met {
		fmt.Println("    Type: " + j.Namespace.Element(1).Value)
	}
	return met, nil
}
func printCollectMetrics(p *pluginProxy, m []Metric) error {
	defer timeTrack(time.Now(), "printCollectMetrics")
	cltd, err := p.plugin.(Collector).CollectMetrics(m)
	if err != nil {
		return err
	}
	fmt.Println("Collected Metrics include: ")
	for _, j := range cltd {
		fmt.Printf("    Type: %10T  Value: %v \n", j.Data, j.Data)
	}
	return nil
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("    %s took %s \n\n", name, elapsed)
}
