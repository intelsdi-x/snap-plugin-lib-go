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
	"fmt"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

// Metric contains all info related to a Snap Metric
type Metric struct {
	Namespace   namespace
	Version     int64
	Config      Config
	Data        interface{}
	Tags        map[string]string
	Timestamp   time.Time
	Unit        string
	Description string
	//Unexported but passed through for legacy reasons
	lastAdvertisedTime time.Time
}

// Converts a metric to an protobuf metric.
// Returns an error in the case where the metric.Data is not one of the
// supported types.
func toProtoMetric(mt Metric) (*rpc.Metric, error) {
	if mt.Timestamp == (time.Time{}) {
		//Timestamp is unitialized, set to time.Now()
		mt.Timestamp = time.Now()
	}

	if mt.lastAdvertisedTime == (time.Time{}) {
		// lastAdvertisedTime is unitialized, set to time.Now()
		mt.lastAdvertisedTime = time.Now()
	}
	metric := &rpc.Metric{
		Namespace:   toProtoNamespace(mt.Namespace),
		Version:     mt.Version,
		Tags:        mt.Tags,
		Unit:        mt.Unit,
		Description: mt.Description,
		Timestamp: &rpc.Time{
			Sec:  mt.Timestamp.Unix(),
			Nsec: int64(mt.Timestamp.Nanosecond()),
		},
		LastAdvertisedTime: &rpc.Time{
			Sec:  mt.lastAdvertisedTime.Unix(),
			Nsec: int64(mt.lastAdvertisedTime.Nanosecond()),
		},
	}
	switch t := mt.Data.(type) {
	case string:
		metric.Data = &rpc.Metric_StringData{StringData: t}
	case float64:
		metric.Data = &rpc.Metric_Float64Data{Float64Data: t}
	case float32:
		metric.Data = &rpc.Metric_Float32Data{Float32Data: t}
	case int32:
		metric.Data = &rpc.Metric_Int32Data{Int32Data: t}
	case int:
		metric.Data = &rpc.Metric_Int64Data{Int64Data: int64(t)}
	case int64:
		metric.Data = &rpc.Metric_Int64Data{Int64Data: t}
	case []byte:
		metric.Data = &rpc.Metric_BytesData{BytesData: t}
	case bool:
		metric.Data = &rpc.Metric_BoolData{BoolData: t}
	case nil:
		metric.Data = nil
	default:
		return nil, fmt.Errorf("unsupported type: %s given in metric data", t)
	}
	return metric, nil
}

func fromProtoMetric(mt *rpc.Metric) Metric {
	metric := Metric{
		Namespace:   fromProtoNamespace(mt.Namespace),
		Version:     mt.Version,
		Tags:        mt.Tags,
		Unit:        mt.Unit,
		Description: mt.Description,
		Config:      Config{},
		Timestamp:   time.Unix(mt.Timestamp.Sec, mt.Timestamp.Nsec),
		lastAdvertisedTime: time.Unix(mt.LastAdvertisedTime.Sec,
			mt.LastAdvertisedTime.Nsec),
	}
	metric.Config = fromProtoConfig(mt.Config)

	switch mt.Data.(type) {
	case *rpc.Metric_BytesData:
		metric.Data = mt.GetBytesData()
	case *rpc.Metric_StringData:
		metric.Data = mt.GetStringData()
	case *rpc.Metric_Float64Data:
		metric.Data = mt.GetFloat64Data()
	case *rpc.Metric_Float32Data:
		metric.Data = mt.GetFloat32Data()
	case *rpc.Metric_Int64Data:
		metric.Data = mt.GetInt64Data()
	case *rpc.Metric_Int32Data:
		metric.Data = mt.GetInt32Data()
	case *rpc.Metric_BoolData:
		metric.Data = mt.GetBoolData()
	}

	return metric
}

func fromProtoConfig(config *rpc.ConfigMap) Config {
	cfg := make(Config)
	if config == nil {
		return cfg
	}

	if config.IntMap != nil {
		for k, v := range config.IntMap {
			cfg[k] = v
		}
	}
	if config.StringMap != nil {
		for k, v := range config.StringMap {
			cfg[k] = v
		}
	}
	if config.BoolMap != nil {
		for k, v := range config.BoolMap {
			cfg[k] = v
		}
	}

	if config.FloatMap != nil {
		for k, v := range config.FloatMap {
			cfg[k] = v
		}
	}
	return cfg
}

func toProtoNamespace(ns namespace) []*rpc.NamespaceElement {
	Elements := make([]*rpc.NamespaceElement, 0)
	for _, ele := range ns {
		Element := &rpc.NamespaceElement{
			Value:       ele.Value,
			Description: ele.Description,
			Name:        ele.Name,
		}
		Elements = append(Elements, Element)
	}
	return Elements
}

func fromProtoNamespace(ns []*rpc.NamespaceElement) namespace {
	var nse namespace
	for _, ele := range ns {
		element := namespaceElement{
			Value:       ele.Value,
			Description: ele.Description,
			Name:        ele.Name,
		}
		nse = append(nse, element)
	}
	return nse
}

type namespace []namespaceElement

// String returns the string representation of the namespace with "/" joining
// the elements of the namespace.  A leading "/" is added.
func (n namespace) String() string {
	ns := n.Strings()
	return "/" + strings.Join(ns, "/")
}

// Strings returns an array of strings that represent the elements of the
// namespace.
func (n namespace) Strings() []string {
	var ns []string
	for _, namespaceElement := range n {
		ns = append(ns, namespaceElement.Value)
	}
	return ns
}

// Key returns a string representation of the namespace with "." joining
// the elements of the namespace.
func (n namespace) Key() string {
	return strings.Join(n.Strings(), ".")
}

// IsDynamic returns true if there is any element of the namespace which is
// dynamic.  If the namespace is dynamic the second return value will contain
// an array of namespace elements (indexes) where there are dynamic namespace
// elements. A dynamic component of the namespace are those elements that
// contain variable data.
func (n namespace) IsDynamic() (bool, []int) {
	var idx []int
	ret := false
	for i := range n {
		if n[i].IsDynamic() {
			ret = true
			idx = append(idx, i)
		}
	}
	return ret, idx
}

// Newnamespace takes an array of strings and returns a namespace.  A namespace
// is an array of namespaceElements.  The provided array of strings is used to
// set the corresponding Value fields in the array of namespaceElements.
func NewNamespace(ns ...string) namespace {
	n := make([]namespaceElement, len(ns))
	for i, ns := range ns {
		n[i] = namespaceElement{Value: ns}
	}
	return n
}

// AddDynamicElement adds a dynamic element to the given namespace.  A dynamic
// namespaceElement is defined by having a nonempty Name field.
func (n namespace) AddDynamicElement(name, description string) namespace {
	nse := namespaceElement{Name: name, Description: description, Value: "*"}
	return append(n, nse)
}

// AddStaticElement adds a static element to the given namespace.  A static
// namespaceElement is defined by having an empty Name field.
func (n namespace) AddStaticElement(value string) namespace {
	nse := namespaceElement{Value: value}
	return append(n, nse)
}

// AddStaticElements adds a static elements to the given namespace.  A static
// namespaceElement is defined by having an empty Name field.
func (n namespace) AddStaticElements(values ...string) namespace {
	for _, value := range values {
		n = append(n, namespaceElement{Value: value})
	}
	return n
}

func (n namespace) Element(idx int) namespaceElement {
	if idx >= 0 && idx < len(n) {
		return n[idx]
	}
	return namespaceElement{}
}

// namespaceElement provides meta data related to the namespace.
// This is of particular importance when the namespace contains data.
type namespaceElement struct {
	Value       string
	Description string
	Name        string
}

// NewNamespaceElement tasks a string and returns a namespaceElement where the
// Value field is set to the provided string argument.
func NewNamespaceElement(e string) namespaceElement {
	if e != "" {
		return namespaceElement{Value: e}
	}
	return namespaceElement{}
}

// IsDynamic returns true if the namespace element contains data.  A namespace
// element that has a nonempty Name field is considered dynamic.
func (n *namespaceElement) IsDynamic() bool {
	if n.Name != "" {
		return true
	}
	return false
}
