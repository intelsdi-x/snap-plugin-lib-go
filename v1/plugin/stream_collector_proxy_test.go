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

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

type mockStreamServer struct {
	grpc.ServerStream
	sendChan chan *rpc.CollectReply
	recvChan chan *rpc.CollectArg
}

func (m mockStreamServer) Context() context.Context {
	return context.TODO()
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
	// Send a metric down to ch every t time
	mockStreamAction := func(ch chan []Metric, t time.Duration, mts []Metric) {
		for {
			ch <- mts
			time.Sleep(t)
		}
	}
	Convey("TestStreamMetrics", t, func(c C) {
		Convey("Error calling StreamMetrics", func() {
			sp := StreamProxy{
				pluginProxy:        *newPluginProxy(newMockErrStreamer()),
				plugin:             newMockErrStreamer(),
				maxMetricsBuffer:   defaultMaxMetricsBuffer,
				maxCollectDuration: defaultMaxCollectDuration,
			}
			errChan := make(chan string)
			err := sp.plugin.StreamMetrics(context.Background(), sp.recvChan, sp.sendChan, errChan)
			So(err, ShouldNotBeNil)
		})
		Convey("Successful Call to StreamMetrics", func(c C) {
			// Make a successful call to stream metrics
			pl := newMockStreamer()
			sp := StreamProxy{
				pluginProxy:        *newPluginProxy(newMockStreamer()),
				plugin:             pl,
				maxMetricsBuffer:   defaultMaxMetricsBuffer,
				maxCollectDuration: defaultMaxCollectDuration,
			}
			s := mockStreamServer{}
			go func() {
				err := sp.StreamMetrics(s)
				c.So(err, ShouldBeNil)
			}()
			Convey("Successful call, stream error", func() {
				// plugin returns an error.
			})
			Convey("metrics sent", func() {
				// plugin returns metrics
			})
		})
		Convey("Successfully stream metrics from plugin immediately", func(c C) {
			pl := newMockStreamerStream(mockStreamAction)
			sp := StreamProxy{
				pluginProxy:        *newPluginProxy(newMockStreamer()),
				plugin:             pl,
				maxMetricsBuffer:   defaultMaxMetricsBuffer,
				maxCollectDuration: defaultMaxCollectDuration,
			}
			s := mockStreamServer{}
			go func() {
				err := sp.StreamMetrics(s)
				c.So(err, ShouldBeNil)
			}()
			// Need to give time for streamMetrics call to propagate
			time.Sleep(time.Millisecond * 100)
			Convey("get metrics through stream proxy", func() {
				So(pl.outMetric, ShouldNotBeNil)
				// Create mocked metrics
				metrics := []Metric{
					Metric{},
				}
				// Send metrics down to channel every 100 ms
				pl.doAction(time.Millisecond*100, metrics)
				select {
				case mts := <-sp.sendChan:
					// Success! we got something....
					So(mts, ShouldNotBeNil)
				case <-time.After(1 * time.Second):
					t.Fatal("timed out waiting for metrics to go through stream collector")
				}
			})
		})
		Convey("Successfully stream metrics from plugin", func(c C) {
			pl := newMockStreamerStream(mockStreamAction)

			// Set maxMetricsBuffer to define buffer capacity
			sp := StreamProxy{
				pluginProxy:        *newPluginProxy(newMockStreamer()),
				plugin:             pl,
				maxMetricsBuffer:   5,
				maxCollectDuration: time.Millisecond * 200,
			}
			s := mockStreamServer{}
			go func() {
				err := sp.StreamMetrics(s)
				c.So(err, ShouldBeNil)
			}()
			// Create mocked metrics
			metrics := []Metric{}
			for i := 0; i < 2; i++ {
				metrics = append(metrics, Metric{})
			}
			// Need to give time for streamMetrics call to propagate
			time.Sleep(time.Millisecond * 100)
			Convey("get buffered metrics through stream proxy when maxMetricsBuffer is reached", func() {
				Convey("when maxMetricsBuffer is reached", func() {
					So(pl.outMetric, ShouldNotBeNil)
					// Send metrics down to channel every 20 ms
					pl.doAction(time.Millisecond*20, metrics)
					select {
					case mts := <-sp.sendChan:
						// Success! we got something....
						So(mts, ShouldNotBeNil)
						// Expect to get 5 metrics (see value of maxMetricsBuffer)
					case <-time.After(time.Second):
						t.Fatal("timed out waiting for metrics to go through stream collector")
					}
				})
				Convey("when maxCollectDuration is exceeded", func() {
					So(pl.outMetric, ShouldNotBeNil)
					// Send metrics down to channel every 300 ms
					// notice it is longer than set maxCollectDuration
					pl.doAction(time.Millisecond*300, metrics)
					select {
					case mts := <-sp.sendChan:
						// Success! we got something....
						So(mts, ShouldNotBeNil)
					// Expect to get 2 metrics, so even a buffer is not full (its capacity is 5),
					// data will be send after exceeding maxCollectDuration = 200 ms
					case <-time.After(time.Second):
						t.Fatal("timed out waiting for metrics to go through stream collector")
					}
				})
			})
		})
	})
}
