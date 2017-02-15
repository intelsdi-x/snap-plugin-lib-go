package plugin

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

type StreamProxy struct {
	pluginProxy

	plugin StreamCollector
}

func (c *StreamProxy) StreamMetrics(stream rpc.StreamCollector_StreamMetricsServer) error {
	// Error channel where we will forward plugin errors to snap where it
	// can report/handle them.
	errChan := make(chan string)
	// Metrics into the plugin from snap.
	inChan := make(chan []Metric)
	// Metrics out of the plugin into snap.
	outChan := make(chan []Metric)
	err := c.plugin.StreamMetrics(inChan, outChan, errChan)
	if err != nil {
		return err
	}
	go metricSend(c.plugin, outChan, stream)
	go errorSend(c.plugin, errChan, stream)
	streamRecv(c.plugin, inChan, stream)
	return nil
}

func errorSend(
	plugin StreamCollector,
	ch chan string,
	s rpc.StreamCollector_StreamMetricsServer) {
	for r := range ch {
		reply := &rpc.CollectReply{
			Error: &rpc.ErrReply{Error: r},
		}
		if err := s.Send(reply); err != nil {
			fmt.Println(err.Error())
		}

	}
}

func metricSend(
	plugin StreamCollector,
	ch chan []Metric,
	s rpc.StreamCollector_StreamMetricsServer) {
	for r := range ch {
		mts := []*rpc.Metric{}
		for _, mt := range r {
			metric, err := toProtoMetric(mt)
			if err != nil {
				fmt.Println(err.Error())
			}
			mts = append(mts, metric)
		}
		reply := &rpc.CollectReply{
			Metrics_Reply: &rpc.MetricsReply{Metrics: mts},
		}
		if err := s.Send(reply); err != nil {
			fmt.Println(err.Error())
		}
	}

}

func streamRecv(
	plugin StreamCollector,
	ch chan []Metric,
	s rpc.StreamCollector_StreamMetricsServer) {

	for {
		s, err := s.Recv()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if s != nil {
			if s.Metrics_Arg != nil {
				metrics := []Metric{}
				for _, mt := range s.Metrics_Arg.Metrics {
					metric := fromProtoMetric(mt)
					metrics = append(metrics, metric)
				}
				ch <- metrics
			}
		}
	}
}

func (c *StreamProxy) SetConfig(context.Context, *rpc.ConfigMap) (*rpc.ErrReply, error) {
	return nil, nil
}

func (c *StreamProxy) GetMetricTypes(ctx context.Context, arg *rpc.GetMetricTypesArg) (*rpc.MetricsReply, error) {
	cfg := fromProtoConfig(arg.Config)

	r, err := c.plugin.GetMetricTypes(cfg)
	if err != nil {
		return nil, err
	}
	metrics := []*rpc.Metric{}
	for _, mt := range r {
		// We can ignore this error since we are not returning data from
		// GetMetricTypes.
		metric, _ := toProtoMetric(mt)
		metrics = append(metrics, metric)
	}
	reply := &rpc.MetricsReply{
		Metrics: metrics,
	}
	return reply, nil
}
