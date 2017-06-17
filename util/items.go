package util

import (
	"errors"
	"sync"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	"golang.org/x/net/context"
)

type streamsMgr struct {
	*sync.Mutex
	collection map[rpc.StreamCollector_StreamMetricsServer]context.CancelFunc
}

func New() *streamsMgr {
	return &streamsMgr{
		Mutex:      &sync.Mutex{},
		collection: make(map[rpc.StreamCollector_StreamMetricsServer]context.CancelFunc),
	}
}

func (s *streamsMgr) Add(stream rpc.StreamCollector_StreamMetricsServer, cancel context.CancelFunc) {
	s.Lock()
	defer s.Unlock()
	s.collection[stream] = cancel
}

func (s *streamsMgr) RemoveAndCancel(stream rpc.StreamCollector_StreamMetricsServer) error {
	s.Lock()
	defer s.Unlock()
	cancel, ok := s.collection[stream]
	if !ok {
		return errors.New("stream not found")
	}
	cancel()
	delete(s.collection, stream)
	return nil
}

func (s *streamsMgr) GetAll() []rpc.StreamCollector_StreamMetricsServer {
	s.Lock()
	defer s.Unlock()
	keys := make([]rpc.StreamCollector_StreamMetricsServer, len(s.collection))
	i := 0
	for k := range s.collection {
		keys[i] = k
		i++
	}
	return keys
}

func (s *streamsMgr) Count() int {
	s.Lock()
	defer s.Unlock()
	return len(s.collection)
}
