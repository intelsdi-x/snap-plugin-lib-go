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
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	maxPingTimeoutLimit = 65535
	tlsTestCA           = "libtest-CA"
	tlsTestSrv          = "libtest-srv"
	tlsTestCli          = "libtest-cli"
	crtFileExt          = ".crt"
	keyFileExt          = ".key"
	badCrtFileExt       = "-BAD.crt"
)

var (
	mockInputOutputInUse *mockInputOutput
	testTLSSetupInUse    *testServerSetup
	prevPingTimeoutLimit int
	grpcOptsBuilderInUse *grpcOptsBuilder
	testFilesToRemove    []string
)

type testServerSetup struct {
	prevServerSetup tlsServerSetup
	caCertPath      string
}

func newTestServerSetup(prevServerSetup tlsServerSetup) *testServerSetup {
	return &testServerSetup{prevServerSetup: prevServerSetup}
}

// makeTLSConfig implementation that supports injecting CA certificate for
// verification of TLS client certs.
func (m *testServerSetup) makeTLSConfig() *tls.Config {
	tlsConfig := m.prevServerSetup.makeTLSConfig()
	if m.caCertPath == "" {
		return tlsConfig
	}
	b, err := ioutil.ReadFile(m.caCertPath)
	if err != nil {
		panic(err)
	}
	tlsConfig.ClientCAs = x509.NewCertPool()
	tlsConfig.ClientCAs.AppendCertsFromPEM(b)
	return tlsConfig
}

func (m *testServerSetup) readRootCAs() (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	return rootCAs, nil
}

func (m *testServerSetup) updateServerOptions(options ...grpc.ServerOption) []grpc.ServerOption {
	opts := m.prevServerSetup.updateServerOptions(options...)
	opts = append(opts, grpc.MaxConcurrentStreams(2))
	return opts
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

func TestMain(m *testing.M) {
	setUpTestMain()
	retCode := m.Run()
	tearDownTestMain()
	os.Exit(retCode)
}

func TestIncorrectPluginArgsFail(t *testing.T) {
	FocusConvey("Intending to start secure plugin server", t, func() {
		setUpSecureTestcase(true, true)
		Convey("omitting Cert Path from arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArgs = strings.Fields(fmt.Sprintf(`mock
				{"KeyPath":"%s","TLSEnabled":true}`,
				tlsTestSrv+keyFileExt))
			So(func() {
				startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
			}, ShouldPanic)
		})
		Convey("omitting Key Path from arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArgs = strings.Fields(fmt.Sprintf(`mock
				{"CertPath":"%s","TLSEnabled":true}`,
				tlsTestSrv+crtFileExt))
			So(func() {
				startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
			}, ShouldPanic)
		})
		Convey("omitting TLSEnabled flag from arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArgs = strings.Fields(fmt.Sprintf(`mock
				{"CertPath":"%s","KeyPath":"%s"}`,
				tlsTestSrv+crtFileExt, tlsTestSrv+keyFileExt))
			So(func() {
				startSecureGrpcPlugin(t, &mockCollector{}, collectorType, "mock-coll")
			}, ShouldPanic)
		})
		Convey("adding mismatched certificate and key in arguments will make plugin fail", func() {
			mockInputOutputInUse.mockArgs = strings.Fields(fmt.Sprintf(`mock
				{"CertPath":"%s","KeyPath":"%s","TLSEnabled":true}`,
				tlsTestSrv+crtFileExt, tlsTestCli+keyFileExt))
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
		Convey("when plugin server is started without client CA cert", func() {
			setUpSecureTestcase(true, true)
			testTLSSetupInUse.caCertPath = tlsTestCA + badCrtFileExt
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

func setUpTestMain() {
	rand.Seed(time.Now().Unix())
	if tlsTestFiles, err := buildTLSCerts(tlsTestCA, tlsTestSrv, tlsTestCli); err != nil {
		panic(err)
	} else {
		testFilesToRemove = append(testFilesToRemove, tlsTestFiles...)
	}
}

func tearDownTestMain() {
	for _, fn := range testFilesToRemove {
		os.Remove(fn)
	}
}

func setUpSecureTestcase(serverTLSUp, clientTLSUp bool) {
	mockInputOutputInUse = newMockInputOutput(libInputOutput)
	libInputOutput = mockInputOutputInUse
	testTLSSetupInUse = newTestServerSetup(tlsSetup)
	tlsSetup = testTLSSetupInUse
	grpcOptsBuilderInUse = newGrpcOptsBuilder()
	if serverTLSUp {
		mockInputOutputInUse.mockArgs = strings.Fields(fmt.Sprintf(`mock
			{"CertPath":"%s","KeyPath":"%s","TLSEnabled":true}`,
			tlsTestSrv+crtFileExt, tlsTestSrv+keyFileExt))
		testTLSSetupInUse.caCertPath = tlsTestCA + crtFileExt
	}
	if clientTLSUp {
		grpcOptsBuilderInUse.
			setCACertPath(tlsTestCA+crtFileExt).
			setClientCertKeyPath(tlsTestCli+crtFileExt, tlsTestCli+keyFileExt).
			setSecure(true)
	}
}

func tearDownSecureTestcase() {
	tlsSetup = testTLSSetupInUse.prevServerSetup
	libInputOutput = mockInputOutputInUse.prevInputOutput
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

// buildTLSCerts builds a set of certificates and private keys for testing TLS.
// Generated files include: CA certificate, server certificate and private key,
// client certificate and private key, and alternate (BAD) CA certificate.
// Certificate and key files are named after given common names (e.g.: srvCN).
func buildTLSCerts(caCN, srvCN, cliCN string) (resFiles []string, err error) {
	ou := fmt.Sprintf("%06x", rand.Intn(1<<24))
	u := &certTestUtil{}
	caCertTpl, caCert, caPrivKey, err := u.makeCACertKeyPair(caCN, ou, defaultKeyValidPeriod)
	if err != nil {
		return nil, err
	}
	caCertFn := caCN + crtFileExt
	if err := u.writePEMFile(caCertFn, certificatePEMHeader, caCert); err != nil {
		return nil, err
	}
	resFiles = append(resFiles, caCertFn)
	_, caBadCert, _, err := u.makeCACertKeyPair(caCN, ou, defaultKeyValidPeriod)
	if err != nil {
		return resFiles, err
	}
	badCaCertFn := caCN + badCrtFileExt
	if err := u.writePEMFile(badCaCertFn, certificatePEMHeader, caBadCert); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, badCaCertFn)
	srvCert, srvPrivKey, err := u.makeSubjCertKeyPair(srvCN, ou, defaultKeyValidPeriod, caCertTpl, caPrivKey)
	if err != nil {
		return resFiles, err
	}
	srvCertFn := srvCN + crtFileExt
	srvKeyFn := srvCN + keyFileExt
	if err := u.writePEMFile(srvCertFn, certificatePEMHeader, srvCert); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, srvCertFn)
	if err := u.writePEMFile(srvKeyFn, rsaKeyPEMHeader, x509.MarshalPKCS1PrivateKey(srvPrivKey)); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, srvKeyFn)
	cliCert, cliPrivKey, err := u.makeSubjCertKeyPair(cliCN, ou, defaultKeyValidPeriod, caCertTpl, caPrivKey)
	if err != nil {
		return resFiles, err
	}
	cliCertFn := cliCN + crtFileExt
	cliKeyFn := cliCN + keyFileExt
	if err := u.writePEMFile(cliCertFn, certificatePEMHeader, cliCert); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, cliCertFn)
	if err := u.writePEMFile(cliKeyFn, rsaKeyPEMHeader, x509.MarshalPKCS1PrivateKey(cliPrivKey)); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, cliKeyFn)
	return resFiles, nil
}
