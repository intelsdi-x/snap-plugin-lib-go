// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

var (
	mockInputOutputInUse *mockInputOutput
	mockInputRootCerts   []string
	grpcOptsBuilderInUse *grpcOptsBuilder
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

type testProxyCtor struct {
	prevProxyCtor    pluginProxyConstructor
	onCreateCallback func(Plugin, *pluginProxy)
}

func newTestProxyCtor(prevProxyCtor pluginProxyConstructor, onCreateCallback func(Plugin, *pluginProxy)) *testProxyCtor {
	return &testProxyCtor{prevProxyCtor: prevProxyCtor, onCreateCallback: onCreateCallback}
}

func (tc *testProxyCtor) create(plugin Plugin) *pluginProxy {
	proxy := tc.prevProxyCtor(plugin)
	tc.onCreateCallback(plugin, proxy)
	return proxy
}

type grpcOptsBuilder struct {
	caCertPath string
	certPath   string
	keyPath    string
	secure     bool
}

func (gb *grpcOptsBuilder) build() ([]grpc.DialOption, error) {
	grpcDialOpts := []grpc.DialOption{
		grpc.WithTimeout(GrpcTimeoutDefault),
	}
	if !gb.secure {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
		return grpcDialOpts, nil
	}
	tlsConfig := tls.Config{}
	if gb.certPath != "" && gb.keyPath != "" {
		cert, err := tls.LoadX509KeyPair(gb.certPath, gb.keyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if gb.caCertPath != "" {
		b, err := ioutil.ReadFile(gb.caCertPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(b)
	} else {
		var err error
		tlsConfig.RootCAs, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
	}
	creds := credentials.NewTLS(&tlsConfig)
	grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(creds))
	return grpcDialOpts, nil
}

func (gb *grpcOptsBuilder) setSecure(secure bool) *grpcOptsBuilder {
	gb.secure = secure
	return gb
}

func (gb *grpcOptsBuilder) setCACertPath(caCertPath string) *grpcOptsBuilder {
	gb.caCertPath = caCertPath
	return gb
}

func (gb *grpcOptsBuilder) setClientCertKeyPath(certPath, keyPath string) *grpcOptsBuilder {
	gb.certPath = certPath
	gb.keyPath = keyPath
	return gb
}

func newGrpcOptsBuilder() *grpcOptsBuilder {
	return &grpcOptsBuilder{}
}

func TestIncorrectPluginArgsFail(t *testing.T) {
	// The tests below are designed to recover from a panic so we interpret
	// the return code and panic if the grpc server could not be created.
	cli.OsExiter = func(ret int) {
		// a return code of 2 indicates a failure to build the grpc server
		if ret == 2 {
			panic(errors.New("failed to create grpc server in test"))
		}
		os.Exit(ret)
	}
	Convey("Intending to start secure plugin server", t, func() {
		setUpSecureTestcase(true, true)
		Convey("omitting Cert Path from arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArg = fmt.Sprintf(`
				{"KeyPath":"%s","TLSEnabled":true}`,
				tlsTestSrv+keyFileExt)
			So(func() {
				startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
			}, ShouldPanic)
		})
		Convey("omitting Key Path from arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArg = fmt.Sprintf(`
				{"CertPath":"%s","TLSEnabled":true}`,
				tlsTestSrv+crtFileExt)
			So(func() {
				startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
			}, ShouldPanic)
		})
		Convey("omitting TLSEnabled flag from arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArg = fmt.Sprintf(`
				{"CertPath":"%s","KeyPath":"%s"}`,
				tlsTestSrv+crtFileExt, tlsTestSrv+keyFileExt)
			So(func() {
				startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
			}, ShouldPanic)
		})
		Convey("adding mismatched certificate and key in arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArg = fmt.Sprintf(`
				{"CertPath":"%s","KeyPath":"%s","TLSEnabled":true}`,
				tlsTestSrv+crtFileExt, tlsTestCli+keyFileExt)
			So(func() {
				startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
			}, ShouldPanic)
		})
		Reset(func() {
			tearDownSecureTestcase()
		})
	})
}

func TestSecureCollectorGrpc(t *testing.T) {
	setUpSecureTestcase(true, true)
	defer tearDownSecureTestcase()

	tt := startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
	defer tt.tearDown()

	grpcOpts, err := grpcOptsBuilderInUse.build()
	if err != nil {
		panic(err)
	}
	tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
	if err != nil {
		panic(err)
	}
	// verify collector responds over secure channel
	testCollectorGrpcBackend(t, tt)
}

func TestSecureProcessorGrpc(t *testing.T) {
	setUpSecureTestcase(true, true)
	defer tearDownSecureTestcase()

	tt := startSecureGrpcPlugin(t, &mockProcessor{}, processorType, "mock-proc")
	defer tt.tearDown()

	grpcOpts, err := grpcOptsBuilderInUse.build()
	if err != nil {
		panic(err)
	}
	tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
	if err != nil {
		panic(err)
	}
	testProcessorGrpcBackend(t, tt)
}

func TestSecurePublisherGrpc(t *testing.T) {
	setUpSecureTestcase(true, true)
	defer tearDownSecureTestcase()

	tt := startSecureGrpcPlugin(t, &mockPublisher{}, publisherType, "mock-pub")
	defer tt.tearDown()

	grpcOpts, err := grpcOptsBuilderInUse.build()
	if err != nil {
		panic(err)
	}
	tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
	if err != nil {
		panic(err)
	}
	testPublisherGrpcBackend(t, tt)
}

func TestInvalidClientFailsAgainstTLSServer(t *testing.T) {
	Convey("When secure plugin is started", t, func() {
		setUpSecureTestcase(true, false)
		tt := startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
		Convey("insecure client should fail to connect or ping", func() {
			grpcOpts, err := grpcOptsBuilderInUse.build()
			if err != nil {
				panic(err)
			}
			So(func() {
				// connection may not fail immediately even though no valid connection is available so continue and try to ping the plugin
				tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
				if err != nil {
					panic(err)
				}
				cc := tt.clientConn()
				tc := rpc.NewCollectorClient(cc)
				_, err = tc.Ping(tt.ctx, &rpc.Empty{})
				if err != nil {
					panic(err)
				}
			}, ShouldPanic)
		})
		Convey("TLS client not knowing server's CA should fail to connect or ping", func() {
			grpcOptsBuilderInUse.caCertPath = tlsTestCA + badCrtFileExt
			grpcOpts, err := grpcOptsBuilderInUse.build()
			if err != nil {
				panic(err)
			}
			So(func() {
				// connection may not fail immediately even though no valid connection is available so continue and try to ping the plugin
				tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
				if err != nil {
					panic(err)
				}
				cc := tt.clientConn()
				tc := rpc.NewCollectorClient(cc)
				_, err = tc.Ping(tt.ctx, &rpc.Empty{})
				if err != nil {
					panic(err)
				}
			}, ShouldPanic)
		})
		Reset(func() {
			tt.tearDown()
			tearDownSecureTestcase()
		})
	})
}

func TestTLSClientFailsAgainstInvalidServer(t *testing.T) {
	Convey("In the ideal secure world", t, func() {
		Convey("when insecure plugin server is started", func() {
			setUpSecureTestcase(false, true)
			tt := startSecureGrpcPlugin(t, &mockPublisher{}, publisherType, "mock-pub")
			Convey("secure client should fail to connect or ping", func() {
				grpcOpts, err := grpcOptsBuilderInUse.build()
				if err != nil {
					panic(err)
				}
				So(func() {
					tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
					if err != nil {
						panic(err)
					}
					cc := tt.clientConn()
					tc := rpc.NewPublisherClient(cc)
					_, err = tc.Ping(tt.ctx, &rpc.Empty{})
					if err != nil {
						panic(err)
					}
				}, ShouldPanic)
			})
			Reset(func() {
				tt.tearDown()
				tearDownSecureTestcase()
			})
		})
		Convey("when plugin server is started with invalid client CA cert", func() {
			mockInputRootCerts = []string{tlsTestCA + badCrtFileExt}
			setUpSecureTestcase(true, true)
			tt := startSecureGrpcPlugin(t, &mockPublisher{}, publisherType, "mock-pub")
			Convey("secure client should fail to connect or ping", func() {
				grpcOpts, err := grpcOptsBuilderInUse.build()
				if err != nil {
					panic(err)
				}
				So(func() {
					tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
					if err != nil {
						panic(err)
					}
					cc := tt.clientConn()
					tc := rpc.NewPublisherClient(cc)
					_, err = tc.Ping(tt.ctx, &rpc.Empty{})
					if err != nil {
						panic(err)
					}
				}, ShouldPanic)
			})
			Reset(func() {
				tt.tearDown()
				tearDownSecureTestcase()
			})
		})
		Convey("when plugin server is started without client CA cert", func() {
			mockInputRootCerts = []string{""}
			setUpSecureTestcase(true, true)
			tt := startSecureGrpcPlugin(t, &mockPublisher{}, publisherType, "mock-pub")
			Convey("secure client should fail to connect or ping", func() {
				grpcOpts, err := grpcOptsBuilderInUse.build()
				if err != nil {
					panic(err)
				}
				So(func() {
					tt.cc, err = grpcClientConn(tt.srvAddr, grpcOpts)
					if err != nil {
						panic(err)
					}
					cc := tt.clientConn()
					tc := rpc.NewPublisherClient(cc)
					_, err = tc.Ping(tt.ctx, &rpc.Empty{})
					if err != nil {
						panic(err)
					}
				}, ShouldPanic)
			})
			Reset(func() {
				tt.tearDown()
				tearDownSecureTestcase()
			})
		})
	})
}

func setUpSecureTestcase(serverTLSUp, clientTLSUp bool) {
	mockInputOutputInUse = newMockInputOutput(libInputOutput)
	libInputOutput = mockInputOutputInUse
	grpcOptsBuilderInUse = newGrpcOptsBuilder()
	if serverTLSUp {
		var certPaths string
		if len(mockInputRootCerts) > 0 {
			certPaths = strings.Join(mockInputRootCerts, string(filepath.ListSeparator))
		} else {
			certPaths = tlsTestCA + crtFileExt
		}
		rootCertPathsArg := fmt.Sprintf(`"RootCertPaths":"%s"`, certPaths)
		mockInputOutputInUse.mockArg = fmt.Sprintf(`
			{"CertPath":"%s","KeyPath":"%s","TLSEnabled":true,"LogLevel":5,%s}`,
			tlsTestSrv+crtFileExt, tlsTestSrv+keyFileExt, rootCertPathsArg)
	}
	if clientTLSUp {
		grpcOptsBuilderInUse.
			setCACertPath(tlsTestCA+crtFileExt).
			setClientCertKeyPath(tlsTestCli+crtFileExt, tlsTestCli+keyFileExt).
			setSecure(true)
	}
}

func tearDownSecureTestcase() {
	libInputOutput = mockInputOutputInUse.prevInputOutput
	mockInputRootCerts = []string{}
}

func startSecureGrpcPlugin(t *testing.T, plugin Plugin, typeOfPlugin pluginType, pluginName string) *test {
	errChan := make(chan error)
	var errChanValid atomic.Value
	errChanValid.Store(true)
	var prevDoPrintOut = mockInputOutputInUse.doPrintOut
	mockInputOutputInUse.doPrintOut = func(data string) {
		prevDoPrintOut(data)
		errChanValid.Store(false)
		close(errChan)
	}
	defer func() {
		mockInputOutputInUse.doPrintOut = prevDoPrintOut
	}()
	var startupRoutine func()
	switch typeOfPlugin {
	case collectorType:
		startupRoutine = func() {
			StartCollector(plugin.(Collector), pluginName, 1)
		}
	case processorType:
		startupRoutine = func() {
			StartProcessor(plugin.(Processor), pluginName, 1)
		}
	case publisherType:
		startupRoutine = func() {
			StartPublisher(plugin.(Publisher), pluginName, 1)
		}
	}
	// setup new proxy constructor to capture created plugin proxy
	var proxyInUse *pluginProxy
	testProxyCtor := newTestProxyCtor(pluginProxyCtor, func(_ Plugin, proxy *pluginProxy) {
		proxyInUse = proxy
	})
	pluginProxyCtor = testProxyCtor.create
	defer func() {
		pluginProxyCtor = testProxyCtor.prevProxyCtor
	}()
	// start server and wait for preamble to be emitted
	go func() {
		defer func() {
			if r, gotError := recover().(error); gotError && errChanValid.Load().(bool) {
				errChan <- r
			}
		}()
		startupRoutine()
	}()
	pluginErr, stillOpen := <-errChan
	if stillOpen {
		panic(pluginErr)
	}

	var response preamble
	var tt = newTest(t)
	err := json.Unmarshal([]byte(mockInputOutputInUse.output[0]), &response)
	if err != nil {
		panic(err)
	}
	tt.srvAddr = response.ListenAddress
	tt.halt = proxyInUse.halt
	return tt
}

func grpcClientConn(serverAddr string, grpcOpts []grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(serverAddr, grpcOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
