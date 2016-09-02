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
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigPolicy(t *testing.T) {
	p := NewConfigPolicy()
	tc := configPolicyTestCases()
	Convey("Test Config Policy", t, func() {
		for _, c := range tc {
			Convey(fmt.Sprintf("Test Config Policy %+v", strings.Join(c.input.key, "/")), func() {
				switch c.input.rule.(type) {
				case stringRule:
					srule := c.input.rule.(stringRule)
					p.AddStringRule(c.input.key, srule)
					r := newGetConfigPolicyReply(*p).GetStringPolicy()
					Convey("Test ConfigPolicy stringRule", func() {
						for _, v := range r {
							for kk, vv := range v.GetRules() {
								So(kk, ShouldEqual, srule.Key)
								So(vv.Required, ShouldEqual, srule.Required)
								So(vv.HasDefault, ShouldEqual, srule.HasDefault)
								So(vv.Default, ShouldEqual, srule.Default)
							}
						}
					})
				case boolRule:
					brule := c.input.rule.(boolRule)
					p.AddBoolRule(c.input.key, brule)
					r := newGetConfigPolicyReply(*p).GetBoolPolicy()
					Convey("Test ConfigPolicy boolRule", func() {
						for _, v := range r {
							for kk, vv := range v.GetRules() {
								So(kk, ShouldEqual, brule.Key)
								So(vv.Required, ShouldEqual, brule.Required)
								So(vv.HasDefault, ShouldEqual, brule.HasDefault)
								So(vv.Default, ShouldEqual, brule.Default)
							}
						}
					})
				case integerRule:
					irule := c.input.rule.(integerRule)
					p.AddIntRule(c.input.key, irule)
					r := newGetConfigPolicyReply(*p)
					Convey("Test ConfigPolicy integerRule", func() {
						for _, v := range r.GetIntegerPolicy() {
							for kk, vv := range v.GetRules() {
								So(kk, ShouldEqual, irule.Key)
								So(vv.Required, ShouldEqual, irule.Required)
								So(vv.HasDefault, ShouldEqual, irule.HasDefault)
								So(vv.Default, ShouldEqual, irule.Default)
							}
						}
					})
				case floatRule:
					frule := c.input.rule.(floatRule)
					p.AddFloatRule(c.input.key, frule)
					r := newGetConfigPolicyReply(*p)
					Convey("Test ConfigPolicy floatRule", func() {
						for _, v := range r.GetFloatPolicy() {
							for kk, vv := range v.GetRules() {
								So(kk, ShouldEqual, frule.Key)
								So(vv.Required, ShouldEqual, frule.Required)
								So(vv.HasDefault, ShouldEqual, frule.HasDefault)
								So(vv.Default, ShouldEqual, frule.Default)
							}
						}
					})
				}
			})
		}
	})
}

type inputConfigPolicy struct {
	key  []string
	rule interface{}
}

type testCaseConfigPolicy struct {
	input inputConfigPolicy
}

func configPolicyTestCases() []testCaseConfigPolicy {
	tc := []testCaseConfigPolicy{
		// test stringRule
		testCaseConfigPolicy{
			input: inputConfigPolicy{
				key:  []string{"a", "b", "c"},
				rule: stringRule{Key: "xyz", Required: true, Default: "sss", HasDefault: true},
			},
		},
		// test boolRule
		testCaseConfigPolicy{
			input: inputConfigPolicy{
				key:  []string{"a1", "b1", "c1"},
				rule: boolRule{Key: "xyz", Required: true, Default: true, HasDefault: true},
			},
		},
		// test floatRule
		testCaseConfigPolicy{
			input: inputConfigPolicy{
				key:  []string{"a2", "b2", "c2"},
				rule: floatRule{Key: "xyz", Required: true, Default: 12.1, HasDefault: true},
			},
		},
		// test integerRule
		testCaseConfigPolicy{
			input: inputConfigPolicy{
				key:  []string{"a3", "b3", "c3"},
				rule: integerRule{Key: "xyz", Required: true, Default: 12, HasDefault: true},
			},
		},
	}
	return tc
}
