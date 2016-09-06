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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIntegerRule(t *testing.T) {
	tc := integerPassTestCases()
	Convey("Test integerRule Passes", t, func() {
		var fr integerRule
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				fr, err = NewIntegerRule(c.input.key, c.input.req, c.input.opts...)
			} else {
				fr, err = NewIntegerRule(c.input.key, c.input.req)
			}
			So(err, ShouldBeNil)
			So(fr, ShouldResemble, c.expected)
		}
	})

	tc = integerErrTestCases()
	Convey("Test integerRule Fails", t, func() {
		var fr integerRule
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				fr, err = NewIntegerRule(c.input.key, c.input.req, c.input.opts...)
			} else {
				fr, err = NewIntegerRule(c.input.key, c.input.req)
			}
			So(err, ShouldNotBeNil)
			So(fr, ShouldResemble, c.expected)
		}
	})
}

type integerInput struct {
	key  string
	req  bool
	opts []integerRuleOpt
}

type testCaseInteger struct {
	expected integerRule
	input    integerInput
}

func integerPassTestCases() []testCaseInteger {
	tc := []testCaseInteger{
		testCaseInteger{
			expected: integerRule{Key: "xyz", Required: true},
			input:    integerInput{key: "xyz", req: true},
		},
		testCaseInteger{
			expected: integerRule{Key: "abc", Required: false},
			input:    integerInput{key: "abc", req: false},
		},
		testCaseInteger{
			expected: integerRule{Key: "xyz1", Required: true, HasDefault: true, Default: 30},
			input:    integerInput{key: "xyz1", req: true, opts: []integerRuleOpt{SetDefaultInt(30)}},
		},
		testCaseInteger{
			expected: integerRule{Key: "xyz2", Required: true, Maximum: 64, HasMax: true},
			input:    integerInput{key: "xyz2", req: true, opts: []integerRuleOpt{SetMaxInt(64)}},
		},
		testCaseInteger{
			expected: integerRule{Key: "xyz3", Required: true, Minimum: 5, HasMin: true},
			input:    integerInput{key: "xyz3", req: true, opts: []integerRuleOpt{SetMinInt(5)}},
		},
	}
	return tc
}

func integerErrTestCases() []testCaseInteger {
	tc := []testCaseInteger{
		testCaseInteger{
			expected: integerRule{},
			input:    integerInput{req: true},
		},
		testCaseInteger{
			expected: integerRule{},
			input:    integerInput{key: "", req: false},
		},
		testCaseInteger{
			expected: integerRule{},
			input:    integerInput{key: "", req: false, opts: []integerRuleOpt{SetMinInt(5)}},
		},
	}
	return tc
}
