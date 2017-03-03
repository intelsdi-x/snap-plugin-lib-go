// +build small

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
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

type mockStreamServer struct {
	grpc.ServerStream
	sendChan chan *rpc.CollectReply
	recvChan chan *rpc.CollectArg
}

func (m mockStreamServer) Send(arg *rpc.CollectReply) error {
	m.sendChan <- arg
	return nil
}
func (m mockStreamServer) Recv() (*rpc.CollectArg, error) {
	a := <-m.recvChan
	return a, nil
}

func TestStreamMetrics(t *testing.T) {
	// Call into stream metrics
	Convey("TestStreamMetrics", t, func() {
		Convey("Error calling StreamMetrics", func() {
			sp := StreamProxy{
				pluginProxy: *newPluginProxy(newMockErrStreamer()),
				plugin:      newMockErrStreamer(),
			}
			sendChan := make(chan *rpc.CollectReply)
			recvChan := make(chan *rpc.CollectArg)
			s := mockStreamServer{
				sendChan: sendChan,
				recvChan: recvChan,
			}

			err := sp.StreamMetrics(s)
			So(err, ShouldNotBeNil)
		})

		Convey("Successful Call to StreamMetrics", func() {
			// Make a successful call to stream metrics
			pl := newMockStreamer()
			sp := StreamProxy{
				pluginProxy: *newPluginProxy(newMockStreamer()),
				plugin:      pl,
			}
			sendChan := make(chan *rpc.CollectReply)
			recvChan := make(chan *rpc.CollectArg)
			s := mockStreamServer{
				sendChan: sendChan,
				recvChan: recvChan,
			}
			go func() {
				err := sp.StreamMetrics(s)
				So(err, ShouldBeNil)
			}()
			Convey("Successful call, stream error", func() {
				// plugin returns an error.
			})
			Convey("metrics sent", func() {
				// plugin returns metrics
			})
		})
		Convey("Successfully stream metrics from plugin", func() {
			// Sends a metric down ch every t time
			f := func(ch chan []Metric, t time.Duration) {
				mt := Metric{}
				for {
					time.Sleep(t)
					ch <- []Metric{mt}
				}
			}
			pl := newMockStreamerStream(f)
			sp := StreamProxy{
				pluginProxy: *newPluginProxy(newMockStreamer()),
				plugin:      pl,
			}
			sendChan := make(chan *rpc.CollectReply)
			recvChan := make(chan *rpc.CollectArg)
			s := mockStreamServer{
				sendChan: sendChan,
				recvChan: recvChan,
			}
			go func() {
				err := sp.StreamMetrics(s)
				So(err, ShouldBeNil)
			}()
			// Need to give time for streamMetrics call to propogate
			time.Sleep(time.Millisecond * 100)
			Convey("get metrics through stream proxy", func() {
				So(pl.outMetric, ShouldNotBeNil)
				pl.doAction(time.Millisecond * 100)
				select {
				case mts := <-sendChan:
					// Success! we got something....
					So(mts, ShouldNotBeNil)
					So(mts.Metrics_Reply, ShouldNotBeNil)
					So(mts.Metrics_Reply.Metrics, ShouldNotBeNil)
					So(len(mts.Metrics_Reply.Metrics), ShouldEqual, 1)
				case <-time.After(time.Second):
					t.Fatal("timed out waiting for metrics to go through stream collector")
				}
			})
		})
	})
}
