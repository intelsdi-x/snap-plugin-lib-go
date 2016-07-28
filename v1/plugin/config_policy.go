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
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

type ConfigPolicy struct {
	IntegerRules map[string]IntegerRule
	BoolRules    map[string]BoolRule
	StringRules  map[string]StringRule
	FloatRules   map[string]FloatRule
}

func NewConfigPolicy() *ConfigPolicy {
	return &ConfigPolicy{
		IntegerRules: map[string]IntegerRule{},
		BoolRules:    map[string]BoolRule{},
		StringRules:  map[string]StringRule{},
		FloatRules:   map[string]FloatRule{},
	}
}

// AddIntRule adds a given IntegerRule to the IntegerRules map.
// This will overwrite any existing entry.
func (c *ConfigPolicy) AddIntRule(key []string, in IntegerRule) {
	k := strings.Join(key, ".") // Method used on daemon side in ctree
	c.IntegerRules[k] = in
}

// AddBoolRule adds a given BoolRule to the BoolRules map.
// This will overwrite any existing entry.
func (c *ConfigPolicy) AddBoolRule(key []string, in BoolRule) {
	k := strings.Join(key, ".") // Method used in daemon/ctree
	c.BoolRules[k] = in
}

// AddFloatRule adds a given FloatRule to the FloatRules map.
// This will overwrite any existing entry.
func (c *ConfigPolicy) AddFloatRule(key []string, in FloatRule) {
	k := strings.Join(key, ".") // Method used in daemon/ctree
	c.FloatRules[k] = in
}

// AddStringRule adds a given StringRule to the StringRules map.
// This will overwrite any existing entry.
func (c *ConfigPolicy) AddStringRule(key []string, in StringRule) {
	k := strings.Join(key, ".") // Method used in daemon/ctree
	c.StringRules[k] = in
}

func newGetConfigPolicyReply(cfg ConfigPolicy) *rpc.GetConfigPolicyReply {
	ret := &rpc.GetConfigPolicyReply{
		BoolPolicy:    map[string]*rpc.BoolPolicy{},
		FloatPolicy:   map[string]*rpc.FloatPolicy{},
		IntegerPolicy: map[string]*rpc.IntegerPolicy{},
		StringPolicy:  map[string]*rpc.StringPolicy{},
	}

	for k, v := range cfg.IntegerRules {
		r := &rpc.IntegerRule{
			Required:   v.Required,
			Default:    v.Default,
			HasDefault: v.HasDefault,
			Minimum:    v.Minimum,
			HasMin:     v.HasMin,
			Maximum:    v.Maximum,
			HasMax:     v.HasMax,
		}

		if ret.IntegerPolicy[k] == nil {
			ret.IntegerPolicy[k] = &rpc.IntegerPolicy{Rules: map[string]*rpc.IntegerRule{}}
		}
		ret.IntegerPolicy[k].Rules[v.Key] = r
	}

	for k, v := range cfg.FloatRules {
		r := &rpc.FloatRule{
			Required:   v.Required,
			Default:    v.Default,
			HasDefault: v.HasDefault,
			Minimum:    v.Minimum,
			HasMin:     v.HasMin,
			Maximum:    v.Maximum,
			HasMax:     v.HasMax,
		}

		if ret.FloatPolicy[k] == nil {
			ret.FloatPolicy[k] = &rpc.FloatPolicy{Rules: map[string]*rpc.FloatRule{}}
		}
		ret.FloatPolicy[k].Rules[v.Key] = r
	}

	for k, v := range cfg.StringRules {
		r := &rpc.StringRule{
			Required:   v.Required,
			Default:    v.Default,
			HasDefault: v.HasDefault,
		}

		if ret.StringPolicy[k] == nil {
			ret.StringPolicy[k] = &rpc.StringPolicy{Rules: map[string]*rpc.StringRule{}}
		}
		ret.StringPolicy[k].Rules[v.Key] = r
	}

	for k, v := range cfg.BoolRules {
		r := &rpc.BoolRule{
			Required:   v.Required,
			Default:    v.Default,
			HasDefault: v.HasDefault,
		}

		if ret.BoolPolicy[k] == nil {
			ret.BoolPolicy[k] = &rpc.BoolPolicy{Rules: map[string]*rpc.BoolRule{}}
		}
		ret.BoolPolicy[k].Rules[v.Key] = r
	}

	return ret
}
