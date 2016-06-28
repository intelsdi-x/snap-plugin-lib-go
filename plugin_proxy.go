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

// TODO(danielscottt): heartbeats
// TODO(danielscottt): logging
// TODO(danielscottt): plugin panics

type pluginProxy struct {
}

func (*pluginProxy) Ping(ctx context.Context, arg *rpc.Empty) (*rpc.ErrReply, error) {
	return &rpc.ErrReply{}, nil
}

func (*pluginProxy) Kill(ctx context.Context, arg *rpc.KillArg) (*rpc.ErrReply, error) {
	return &rpc.ErrReply{}, nil
}

func (p *pluginProxy) GetConfigPolicy(ctx context.Context, arg *rpc.Empty) (*rpc.GetConfigPolicyReply, error) {
	policy, err := p.plugin.GetConfigPolicy()
	if err != nil {
		return &rpc.GetConfigPolicyReply{
			Error: err.Error(),
		}, nil
	}
	reply, err := rpc.NewGetConfigPolicyReply(policy)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
