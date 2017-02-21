/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

package rand

import (
	"fmt"
	"math/rand"
	"time"

	"errors"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

var (
	strs = []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As I see it yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
	}
)

func init() {
	rand.Seed(42)
}

// Mock collector implementation used for testing
type RandCollector struct {
}

/*  CollectMetrics collects metrics for testing.

CollectMetrics() will be called by Snap when a task that collects one of the metrics returned from this plugins
GetMetricTypes() is started. The input will include a slice of all the metric types being collected.

The output is the collected metrics as plugin.Metric and an error.
*/
func (RandCollector) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}
	for idx, mt := range mts {
		mts[idx].Timestamp = time.Now()
		if val, err := mt.Config.GetBool("return_error"); err == nil && val {
			return nil, errors.New("Houston, we have a problem!")
		}
		if val, err := mt.Config.GetBool("testbool"); err == nil && val {
			continue
		}
		if mt.Namespace[len(mt.Namespace)-1].Value == "integer" {
			if val, err := mt.Config.GetInt("testint"); err == nil {
				mts[idx].Data = val
			} else {
				mts[idx].Data = rand.Int31()
			}
			metrics = append(metrics, mts[idx])
		} else if mt.Namespace[len(mt.Namespace)-1].Value == "float" {
			if val, err := mt.Config.GetFloat("testfloat"); err == nil {
				mts[idx].Data = val
			} else {
				mts[idx].Data = rand.Float64()
			}
			metrics = append(metrics, mts[idx])
		} else if mt.Namespace[len(mt.Namespace)-1].Value == "string" {
			if val, err := mt.Config.GetString("teststring"); err == nil {
				mts[idx].Data = val
			} else {
				mts[idx].Data = strs[rand.Intn(len(strs)-1)]
			}
			metrics = append(metrics, mts[idx])
		} else {
			return nil, fmt.Errorf("Invalid metric: %v", mt.Namespace.Strings())
		}
	}
	return metrics, nil
}

/*
	GetMetricTypes returns metric types for testing.
	GetMetricTypes() will be called when your plugin is loaded in order to populate the metric catalog(where snaps stores all
	available metrics).

	Config info is passed in. This config information would come from global config snap settings.

	The metrics returned will be advertised to users who list all the metrics and will become targetable by tasks.
*/
func (RandCollector) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	vals := []string{"integer", "float", "string"}
	for _, val := range vals {
		metric := plugin.Metric{
			Namespace: plugin.NewNamespace("random", val),
			Version:   1,
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

/*
	GetConfigPolicy() returns the configPolicy for your plugin.

	A config policy is how users can provide configuration info to
	plugin. Here you define what sorts of config info your plugin
	needs and/or requires.
*/
func (RandCollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	policy.AddNewIntRule([]string{"random", "integer"},
		"testint",
		false,
		plugin.SetMaxInt(1000),
		plugin.SetMinInt(0))

	policy.AddNewFloatRule([]string{"random", "float"},
		"testfloat",
		false,
		plugin.SetMaxFloat(1000.0),
		plugin.SetMinFloat(0.0))

	policy.AddNewStringRule([]string{"random", "string"},
		"teststring",
		false)

	policy.AddNewBoolRule([]string{"random"},
		"testbool",
		false)
	return *policy, nil
}
