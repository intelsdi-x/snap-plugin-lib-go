package plugin

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

const (
	defaultMaxCollectDuration = 10 * time.Second
	defaultMaxMetricsBuffer   = 0
)

type StreamProxy struct {
	pluginProxy
	plugin StreamCollector

	// maxMetricsBuffer is the maximum number of metrics the plugin is buffering before sending metrics.
	// Defaults to zero what means send metrics immediately.
	maxMetricsBuffer int64

	// maxCollectionDuration sets the maximum duration (always greater than 0s) between collections before metrics are sent.
	// Defaults to 10s what means that after 10 seconds no new metrics are received, the plugin should send
	// whatever data it has in the buffer instead of waiting longer.
	maxCollectDuration time.Duration
}

func (p *StreamProxy) GetMetricTypes(ctx context.Context, arg *rpc.GetMetricTypesArg) (*rpc.MetricsReply, error) {
	cfg := fromProtoConfig(arg.Config)

	r, err := p.plugin.GetMetricTypes(cfg)
	if err != nil {
		return nil, err
	}
	metrics := []*rpc.Metric{}
	for _, mt := range r {
		// We can ignore this error since we are not returning data from GetMetricTypes.
		metric, _ := toProtoMetric(mt)
		metrics = append(metrics, metric)
	}
	reply := &rpc.MetricsReply{
		Metrics: metrics,
	}
	return reply, nil
}

func (p *StreamProxy) StreamMetrics(stream rpc.StreamCollector_StreamMetricsServer) error {
	if stream == nil {
		return errors.New("Stream metrics server is nil")
	}

	// Error channel where we will forward plugin errors to snap where it
	// can report/handle them.
	errChan := make(chan string)
	// Metrics into the plugin from snap.
	inChan := make(chan []Metric)
	// Metrics out of the plugin into snap.
	outChan := make(chan []Metric)

	err := p.plugin.StreamMetrics(inChan, outChan, errChan)
	if err != nil {
		return err
	}

	go p.metricSend(outChan, stream)
	go p.errorSend(errChan, stream)
	p.streamRecv(inChan, stream)

	return nil
}

func (p *StreamProxy) SetConfig(context.Context, *rpc.ConfigMap) (*rpc.ErrReply, error) {
	return nil, nil
}

func (p *StreamProxy) errorSend(errChan chan string, stream rpc.StreamCollector_StreamMetricsServer) {
	for r := range errChan {
		reply := &rpc.CollectReply{
			Error: &rpc.ErrReply{Error: r},
		}
		if err := stream.Send(reply); err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (p *StreamProxy) metricSend(ch chan []Metric, stream rpc.StreamCollector_StreamMetricsServer) {
	metrics := []*rpc.Metric{}

	for {
		select {
		case mts := <-ch:
			if len(mts) == 0 {
				break
			}

			for _, mt := range mts {
				metric, err := toProtoMetric(mt)
				if err != nil {
					fmt.Println(err.Error())
					break
				}
				metrics = append(metrics, metric)

				// send metrics if maxMetricsBuffer is reached
				// (notice it is only possible for maxMetricsBuffer greater than 0)
				if p.maxMetricsBuffer == int64(len(metrics)) {
					sendReply(metrics, stream)
					metrics = []*rpc.Metric{}
				}
			}

			// send all available metrics immediately for maxMetricsBuffer is 0 (defaults)
			if p.maxMetricsBuffer == 0 {
				sendReply(metrics, stream)
				metrics = []*rpc.Metric{}
			}
		case <-time.After(p.maxCollectDuration):
			// send metrics if maxCollectDuration is reached
			sendReply(metrics, stream)
			metrics = []*rpc.Metric{}

		}
	}
}

func (p *StreamProxy) streamRecv(ch chan []Metric, stream rpc.StreamCollector_StreamMetricsServer) {
	for {
		s, err := stream.Recv()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if s != nil {
			if s.MaxMetricsBuffer > 0 {
				p.setMaxMetricsBuffer(s.MaxMetricsBuffer)
			}
			if s.MaxCollectDuration > 0 {
				p.setMaxCollectDuration(time.Duration(s.MaxCollectDuration))
			}
			if s.Metrics_Arg != nil {
				metrics := []Metric{}
				for _, mt := range s.Metrics_Arg.Metrics {
					metric := fromProtoMetric(mt)
					metrics = append(metrics, metric)
				}
				// send requested metrics to be collected into the stream plugin
				ch <- metrics
			}
		}
	}
}

func (p *StreamProxy) setMaxCollectDuration(d time.Duration) {
	p.maxCollectDuration = d
}

func (p *StreamProxy) setMaxMetricsBuffer(i int64) {
	p.maxMetricsBuffer = i
}

func sendReply(metrics []*rpc.Metric, stream rpc.StreamCollector_StreamMetricsServer) {
	if len(metrics) == 0 {
		fmt.Println("No metrics available to send")
		return
	}

	reply := &rpc.CollectReply{
		Metrics_Reply: &rpc.MetricsReply{Metrics: metrics},
	}

	if err := stream.Send(reply); err != nil {
		fmt.Println(err.Error())
	}
}
