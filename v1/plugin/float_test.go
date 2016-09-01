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

func TestFloatRule(t *testing.T) {
	tc := floatPassTestCases()
	Convey("Test floatRule Passes", t, func() {
		var fr floatRule
		var err error
		for _, c := range tc {
			Convey(fmt.Sprintf("Test floatRule %+v", c.input.key), func() {
				if c.input.opts != nil {
					fr, err = NewFloatRule(c.input.key, c.input.req, c.input.opts...)
				} else {
					fr, err = NewFloatRule(c.input.key, c.input.req)
				}

				So(err, ShouldBeNil)
				So(fr, ShouldResemble, c.expected)
			})
		}
	})

	tc = floatErrTestCases()
	Convey("Test floatRule Fails", t, func() {
		var fr floatRule
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				fr, err = NewFloatRule(c.input.key, c.input.req, c.input.opts...)
			} else {
				fr, err = NewFloatRule(c.input.key, c.input.req)
			}

			So(err, ShouldNotBeNil)
			So(fr, ShouldResemble, c.expected)
		}
	})
}

type floatInput struct {
	key  string
	req  bool
	opts []floatRuleOpt
}

type testCaseFloat struct {
	expected floatRule
	input    floatInput
}

func floatErrTestCases() []testCaseFloat {
	tc := []testCaseFloat{
		testCaseFloat{
			expected: floatRule{},
			input:    floatInput{req: true},
		},
		testCaseFloat{
			expected: floatRule{},
			input:    floatInput{key: "", req: false},
		},
	}
	return tc
}

func floatPassTestCases() []testCaseFloat {
	tc := []testCaseFloat{
		testCaseFloat{
			expected: floatRule{Key: "xyz", Required: true},
			input:    floatInput{key: "xyz", req: true},
		},
		testCaseFloat{
			expected: floatRule{Key: "abc", Required: false},
			input:    floatInput{key: "abc", req: false},
		},
		testCaseFloat{
			expected: floatRule{Key: "xyz1", Required: true, HasDefault: true, Default: 30.1},
			input:    floatInput{key: "xyz1", req: true, opts: []floatRuleOpt{SetDefaultFloat(30.1)}},
		},
		testCaseFloat{
			expected: floatRule{Key: "xyz2", Required: true, Maximum: 32.1, HasMax: true},
			input:    floatInput{key: "xyz2", req: true, opts: []floatRuleOpt{SetMaxFloat(32.1)}},
		},
		testCaseFloat{
			expected: floatRule{Key: "xyz3", Required: false, Minimum: 12.1, HasMin: true},
			input:    floatInput{key: "xyz3", req: false, opts: []floatRuleOpt{SetMinFloat(12.1)}},
		},
	}
	return tc
}
