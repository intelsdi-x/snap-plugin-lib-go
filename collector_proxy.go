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
	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-go/rpc"
)

// TODO(danielscottt): plugin panics

type collectorProxy struct {
	pluginProxy

	plugin Collector
}

func (c *collectorProxy) collectMetrics(ctx context.Context, arg *rpc.MetricsArg) (*rpc.MetricsReply, error) {
	metrics, err := g.plugin.CollectMetrics(arg.Metrics)
	if err != nil {
		return &rpc.CollectMetricsReply{
			Error: err.Error(),
		}, nil
	}
	reply := &rpc.CollectMetricsReply{
		Metrics: metrics,
	}
	return reply, nil
}

func (cp *collectorProxy) getMetricTypes(ctx context.Context, arg *rpc.GetMetricTypesArg) (*rpc.MetricsReply, error) {
	metricTypes, err := g.Plugin.GetMetricTypes(cfgmap)
	if err != nil {
		return &rpc.GetMetricTypesReply{
			Error: err.Error(),
		}, nil
	}
	reply := &rpc.GetMetricTypesReply{
		Metrics: metricTypes,
	}
	return reply, nil
}