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

func TestBoolRule(t *testing.T) {
	ts := createPassTestCases()
	Convey("Test boolRule Passes", t, func() {
		var br boolRule
		var err error
		for _, c := range ts {
			Convey(fmt.Sprintf("Test boolRule %+v", c.input.key), func() {
				if c.input.opts != nil {
					br, err = NewBoolRule(c.input.key, c.input.req, c.input.opts)
				} else {
					br, err = NewBoolRule(c.input.key, c.input.req)
				}
				So(err, ShouldBeNil)
				So(br, ShouldResemble, c.expected)
			})
		}
	})

	ts = createErrTestCases()
	Convey("Test boolRule Fails", t, func() {
		var br boolRule
		var err error
		for _, c := range ts {
			if c.input.opts != nil {
				br, err = NewBoolRule(c.input.key, c.input.req, c.input.opts)
			} else {
				br, err = NewBoolRule(c.input.key, c.input.req)
			}
			So(err, ShouldNotBeNil)
			So(br, ShouldResemble, c.expected)
		}
	})
}

type input struct {
	key  string
	req  bool
	opts boolRuleOpt
}

type testCase struct {
	expected boolRule
	input    input
}

func createPassTestCases() []testCase {
	tc := []testCase{
		testCase{
			expected: boolRule{Key: "abc", Required: true, HasDefault: true, Default: true},
			input:    input{key: "abc", req: true, opts: SetDefaultBool(true)},
		},
		testCase{
			expected: boolRule{Key: "abc1", Required: false, HasDefault: true},
			input:    input{key: "abc1", req: false, opts: SetDefaultBool(false)},
		},
		testCase{
			expected: boolRule{Key: "xyz", Required: true},
			input:    input{key: "xyz", req: true},
		},
		testCase{
			expected: boolRule{Key: "abc2", Required: false},
			input:    input{key: "abc2", req: false},
		},
	}
	return tc
}

func createErrTestCases() []testCase {
	tc := []testCase{
		testCase{
			expected: boolRule{},
			input:    input{req: true},
		},
		testCase{
			expected: boolRule{},
			input:    input{key: "", req: false},
		},
		testCase{
			expected: boolRule{},
			input:    input{key: "", req: false, opts: SetDefaultBool(false)},
		},
	}
	return tc
}
