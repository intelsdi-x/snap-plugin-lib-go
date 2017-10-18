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
	Convey("Test Config applies defaults from config policy", t, func() {
		// create config policy
		mockPolicy := NewConfigPolicy()
		// string rule
		mockPolicy.AddNewStringRule([]string{"static", "string"},
			"teststr",
			false,
			SetDefaultString("some_str"))
		// integer rule
		mockPolicy.AddNewIntRule([]string{"random", "integer"},
			"testint",
			false,
			SetDefaultInt(10))
		// float rule
		mockPolicy.AddNewFloatRule([]string{"random", "float"},
			"testfloat",
			false,
			SetDefaultFloat(11.1))
		// boolean rule
		mockPolicy.AddNewBoolRule([]string{"random"},
			"testbool",
			false,
			SetDefaultBool(true))

		Convey("apply defaults from config policy", func() {
			Convey("when a given config is empty", func() {
				// create a new config
				config := NewConfig()
				So(config, ShouldBeEmpty)
				// update config with defaults from config policy
				config.applyDefaults(*mockPolicy)
				So(config, ShouldNotBeEmpty)
				// 4 configs are expected (`teststr`, `testint`, `testfloat`, `testbool`)
				So(len(config), ShouldEqual, 4)
				Convey("validate  config values", func() {
					So(config["teststr"], ShouldEqual, "some_str")
					So(config["testint"], ShouldEqual, 10)
					So(config["testfloat"], ShouldEqual, 11.1)
					So(config["testbool"], ShouldEqual, true)
				})
			})
			Convey("when a given config overwrite defaults", func() {
				// create a config (non-empty) with provided values different than defaults
				config := Config{
					"teststr":    "some_str2",
					"testint":    int64(123),
					"testfloat":  32.5,
					"anotherstr": "another_str",
				}
				So(config, ShouldNotBeEmpty)
				// 4 configs are given (`teststr`, `testint`, `testfloat`, `anotherstr`)
				So(len(config), ShouldEqual, 4)
				// update config with defaults from config policy
				config.applyDefaults(*mockPolicy)
				// 5 configs are expected after merging with defaults (including `testbool`)
				So(len(config), ShouldEqual, 5)
				Convey("validate  config values", func() {
					So(config["teststr"], ShouldEqual, "some_str2")
					So(config["anotherstr"], ShouldEqual, "another_str")
					So(config["testint"], ShouldEqual, 123)
					So(config["testfloat"], ShouldEqual, 32.5)
					So(config["testbool"], ShouldEqual, true)
				})
			})
		})
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
		{
			expected: "rain",
			input:    "a",
		},
		{
			expected: int64(123),
			input:    "b",
		},
		{
			expected: 32.5,
			input:    "c",
		},
		{
			expected: true,
			input:    "d",
		},
	}
	return tc
}

func configErrTestCases() []testCaseConfig {
	tc := []testCaseConfig{
		{
			expected: true,
			input:    "f",
		},
		{
			expected: true,
			input:    "h",
		},
		{
			expected: 12.4,
			input:    "x",
		},
		{
			expected: 12.1,
			input:    "i",
		},
		{
			expected: "bbb",
			input:    "y",
		},
		{
			expected: "bbb",
			input:    "j",
		},
		{
			expected: int64(12),
			input:    "z",
		},
		{
			expected: int64(13),
			input:    "k",
		},
	}
	return tc
}
