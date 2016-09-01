// +build small

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
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginProxy(t *testing.T) {
	p := newPluginProxy(newMockPlugin())
	last := p.LastPing

	Convey("Test Plugin Proxy", t, func() {
		Convey("Succeed while pinging", func() {
			_, err := p.Ping(context.Background(), &rpc.Empty{})
			So(err, ShouldBeNil)
			So(p.LastPing.Sub(last), ShouldBeGreaterThan, 0)
		})
		Convey("Succeed while killing", func() {
			var err error
			go func() {
				_, err = p.Kill(context.Background(), &rpc.KillArg{Reason: "test killing"})
			}()
			<-p.halt
			So(err, ShouldBeNil)
		})
		Convey("Succeed while getting config policy", func() {
			_, err := p.GetConfigPolicy(context.Background(), &rpc.Empty{})
			So(err, ShouldBeNil)
		})
		Convey("Error while getting config policy", func() {
			pp := newPluginProxy(newMockErrPlugin())
			_, err := pp.GetConfigPolicy(context.Background(), &rpc.Empty{})
			So(err, ShouldNotBeNil)
		})
		Convey("Test Heart Beat", func() {
			p.PingTimeoutDuration = time.Microsecond * 200
			p.HeartbeatWatch()
			_, ok := (<-p.halt)
			So(ok, ShouldEqual, false)
		})
	})
}
