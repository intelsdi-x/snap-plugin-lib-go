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

package reverse

import (
	"strconv"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

// RProcessor test processor
type RProcessor struct {
}

// Process test process function
func (r RProcessor) Process(mts []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}
	for _, m := range mts {
		switch m.Data.(type) {
		case int:
			m.Data = stringToInt(reverse(intToString(m.Data.(int))))
		case int32:
			i32 := int(m.Data.(int32))
			m.Data = stringToInt(reverse(intToString(i32)))
		case int64:
			i64 := int(m.Data.(int64))
			m.Data = stringToInt(reverse(intToString(i64)))
		case float32:
			f32 := float64(m.Data.(float32))
			m.Data = stringToFloat(reverse(floatToString(f32)))
		case float64:
			m.Data = stringToFloat(reverse(floatToString(m.Data.(float64))))
		case string:
			m.Data = reverse(m.Data.(string))
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

/*
	GetConfigPolicy() returns the configPolicy for your plugin.

	A config policy is how users can provide configuration info to
	plugin. Here you define what sorts of config info your plugin
	needs and/or requires.
*/
func (r RProcessor) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	policy.AddNewBoolRule([]string{"random"},
		"testbool",
		false)

	return *policy, nil
}

func reverse(s string) string {
	r := []rune(s)
	l := len(r)
	for i := 0; i < len(r)/2; i++ {
		r[i], r[l-i-1] = r[l-i-1], r[i]
	}
	return string(r)
}

func floatToString(input float64) string {
	return strconv.FormatFloat(input, 'f', 6, 64)
}

func stringToFloat(input string) float64 {
	f, _ := strconv.ParseFloat(input, 64)
	return f
}

func intToString(input int) string {
	return strconv.Itoa(input)
}

func stringToInt(input string) int {
	it, _ := strconv.Atoi(input)
	return it
}
