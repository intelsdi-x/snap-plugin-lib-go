// +build small

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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	data := map[string]interface{}{
		"a": "rain",
		"b": int64(123),
		"c": 32.5,
		"d": true,
		"h": 789,
		"i": 12,
		"j": false,
		"k": true,
	}

	config := mockConfig{config: data}

	var result interface{}
	var err error

	tc := configPassTestCases()
	Convey("Test Config Passes", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test Config Case %+v", c.input), func() {
				switch c.expected.(type) {
				case bool:
					result, err = config.config.GetBool(c.input)
				case float64:
					result, err = config.config.GetFloat(c.input)
				case string:
					result, err = config.config.GetString(c.input)
				case int64:
					result, err = config.config.GetInt(c.input)
				}
				So(err, ShouldBeNil)
				So(result, ShouldEqual, c.expected)
			})
		}
	})

	tc = configErrTestCases()
	Convey("Test Config Fails", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test Config Case %+v", c.input), func() {
				switch c.expected.(type) {
				case bool:
					result, err = config.config.GetBool(c.input)
				case float64:
					result, err = config.config.GetFloat(c.input)
				case string:
					result, err = config.config.GetString(c.input)
				case int64:
					result, err = config.config.GetInt(c.input)
				}
				So(err, ShouldNotBeNil)
			})
		}
	})
}

type mockConfig struct {
	config Config
}

type testCaseConfig struct {
	expected interface{}
	input    string
}

func configPassTestCases() []testCaseConfig {
	tc := []testCaseConfig{
		testCaseConfig{
			expected: "rain",
			input:    "a",
		},
		testCaseConfig{
			expected: int64(123),
			input:    "b",
		},
		testCaseConfig{
			expected: 32.5,
			input:    "c",
		},
		testCaseConfig{
			expected: true,
			input:    "d",
		},
	}
	return tc
}

func configErrTestCases() []testCaseConfig {
	tc := []testCaseConfig{
		testCaseConfig{
			expected: true,
			input:    "f",
		},
		testCaseConfig{
			expected: true,
			input:    "h",
		},
		testCaseConfig{
			expected: 12.4,
			input:    "x",
		},
		testCaseConfig{
			expected: 12.1,
			input:    "i",
		},
		testCaseConfig{
			expected: "bbb",
			input:    "y",
		},
		testCaseConfig{
			expected: "bbb",
			input:    "j",
		},
		testCaseConfig{
			expected: int64(12),
			input:    "z",
		},
		testCaseConfig{
			expected: int64(13),
			input:    "k",
		},
	}
	return tc
}
