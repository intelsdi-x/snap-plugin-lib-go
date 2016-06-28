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

import "time"

// MetaOpt is used to apply optional metadata on a plugin
type MetaOpt func(m *Meta)

// ConcurrencyCount overwrites the default (5) for a Meta's ConcurrencyCount.
func ConcurrencyCount(cc int) MetaOpt {
	return func(m *Meta) {
		m.ConcurrencyCount = cc
	}
}

// Exclusive overwrites the default (false) for a Meta's Exclusive key.
func Exclusive(e bool) MetaOpt {
	return func(m *Meta) {
		m.Exclusive = e
	}
}

// Unsecure overwrites the default (false) for a Meta's Unsecure key.
func Unsecure(e bool) MetaOpt {
	return func(m *Meta) {
		m.Unsecure = e
	}
}

// RoutingStrategy overwrites the default (LRU) for a Meta's RoutingStrategy.
func RoutingStrategy(r RoutingStrategyType) MetaOpt {
	return func(m *Meta) {
		m.RoutingStrategy = r
	}
}

// CacheTTL overwrites the default (500ms) for a Meta's CacheTTL.
func CacheTTL(t time.Duration) MetaOpt {
	return func(m *Meta) {
		m.CacheTTL = t
	}
}

// meta is the metadata for a plugin
type meta struct {
	// A plugin's unique identifier is type:name:version.
	Type    PluginType
	Name    string
	Version int

	// ConcurrencyCount is the max number of concurrent calls the plugin
	// should take.  For example:
	// If there are 5 tasks using the plugin and its concurrency count is 2,
	// snapd will keep 3 plugin instances running.
	ConcurrencyCount int

	// Exclusive == true results in a single instance of the plugin running
	// regardless of the number of tasks using the plugin.
	Exclusive bool

	// Unsecure results in unencrypted communication with this plugin.
	Unsecure bool

	// CacheTTL will override the default cache TTL for the this plugin. snapd
	// caches metrics on the daemon side for a default of 500ms.
	CacheTTL time.Duration

	// RoutingStrategy will override the routing strategy this plugin requires.
	// The default routing strategy is Least Recently Used.
	RoutingStrategy RoutingStrategyType
}

// newMeta sets defaults, applies options, and then returns a Meta struct
func newMeta(plType pluginType, name string, version int, opts ...MetaOpt) meta {
	p := PluginMet{
		Name:             name,
		Version:          version,
		Type:             pluginType,
		ConcurrencyCount: 5,
	}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}
