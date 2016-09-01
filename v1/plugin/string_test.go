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

func TestStringRule(t *testing.T) {
	tc := stringPassTestCases()
	Convey("Test stringRule Passes", t, func() {
		var sr stringRule
		var err error
		for _, c := range tc {
			Convey(fmt.Sprintf("Test stringRule %+v", c.input.key), func() {
				if c.input.opts != nil {
					sr, err = NewStringRule(c.input.key, c.input.req, c.input.opts)
				} else {
					sr, err = NewStringRule(c.input.key, c.input.req)
				}
				So(err, ShouldBeNil)
				So(sr, ShouldResemble, c.expected)
			})
		}
	})

	tc = stringErrTestCases()
	Convey("Test stringRule Fails", t, func() {
		var sr stringRule
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				sr, err = NewStringRule(c.input.key, c.input.req, c.input.opts)
			} else {
				sr, err = NewStringRule(c.input.key, c.input.req)
			}
			So(err, ShouldNotBeNil)
			So(sr, ShouldResemble, c.expected)
		}
	})
}

type stringInput struct {
	key  string
	req  bool
	opts stringRuleOpt
}

type testCaseString struct {
	expected stringRule
	input    stringInput
}

func stringPassTestCases() []testCaseString {
	tc := []testCaseString{
		testCaseString{
			expected: stringRule{Key: "xyz", Required: true},
			input:    stringInput{key: "xyz", req: true},
		},
		testCaseString{
			expected: stringRule{Key: "abc", Required: false},
			input:    stringInput{key: "abc", req: false},
		},
		testCaseString{
			expected: stringRule{Key: "deer", Required: true, HasDefault: true, Default: "123"},
			input:    stringInput{key: "deer", req: true, opts: SetDefaultString("123")},
		},
		testCaseString{
			expected: stringRule{Key: "racoon", Required: false, HasDefault: true, Default: "aaa"},
			input:    stringInput{key: "racoon", req: false, opts: SetDefaultString("aaa")},
		},
	}
	return tc
}

func stringErrTestCases() []testCaseString {
	tc := []testCaseString{
		testCaseString{
			expected: stringRule{},
			input:    stringInput{req: true},
		},
		testCaseString{
			expected: stringRule{},
			input:    stringInput{key: "", req: false},
		},
		testCaseString{
			expected: stringRule{},
			input:    stringInput{key: "", req: true, opts: SetDefaultString("xyz")},
		},
	}
	return tc
}
