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

func TestConfigPolicy(t *testing.T) {
	p := NewConfigPolicy()
	tc := configPolicyTestCases()
	Convey("Test Config Policy", t, func() {
		for _, c := range tc {
			Convey("Test Config Policy "+c.input.key, func() {
				exist := false
				switch c.input.opts.(type) {
				case stringRuleOpt:
					sopts := c.input.opts.(stringRuleOpt)
					p.AddNewStringRule(c.input.ns, c.input.key, c.input.req, sopts)
					r := newGetConfigPolicyReply(*p).GetStringPolicy()
					Convey("Test ConfigPolicy stringRule "+c.input.key, func() {
						for _, v := range r {
							for kk, vv := range v.GetRules() {
								if kk == c.output.key {
									exist = true
									So(vv.Required, ShouldEqual, c.output.req)
									So(vv.HasDefault, ShouldEqual, c.output.hasDefault)
									if vv.HasDefault {
										So(vv.Default, ShouldEqual, c.output.def.(string))
									}
								}
							}
						}
						So(exist, ShouldEqual, true)
					})
				case boolRuleOpt:
					bopts := c.input.opts.(boolRuleOpt)
					p.AddNewBoolRule(c.input.ns, c.input.key, c.input.req, bopts)
					r := newGetConfigPolicyReply(*p).GetBoolPolicy()
					Convey("Test ConfigPolicy boolRule "+c.input.key, func() {
						for _, v := range r {
							for kk, vv := range v.GetRules() {
								if kk == c.output.key {
									exist = true
									So(vv.Required, ShouldEqual, c.output.req)
									So(vv.HasDefault, ShouldEqual, c.output.hasDefault)
									if vv.HasDefault {
										So(vv.Default, ShouldEqual, c.output.def.(bool))
									}
								}
							}
						}
						So(exist, ShouldEqual, true)
					})
				case integerRuleOpt:
					iopts := c.input.opts.(integerRuleOpt)
					p.AddNewIntRule(c.input.ns, c.input.key, c.input.req, iopts)
					r := newGetConfigPolicyReply(*p)
					Convey("Test ConfigPolicy integerRule "+c.input.key, func() {
						for _, v := range r.GetIntegerPolicy() {
							for kk, vv := range v.GetRules() {
								if kk == c.output.key {
									exist = true
									So(vv.Required, ShouldEqual, c.output.req)
									So(vv.HasDefault, ShouldEqual, c.output.hasDefault)
									if vv.HasDefault {
										So(vv.Default, ShouldEqual, c.output.def.(int64))
									}
								}
							}
						}
						So(exist, ShouldEqual, true)
					})
				case floatRuleOpt:
					fopts := c.input.opts.(floatRuleOpt)
					p.AddNewFloatRule(c.input.ns, c.input.key, c.input.req, fopts)
					r := newGetConfigPolicyReply(*p)
					Convey("Test ConfigPolicy floatRule "+c.input.key, func() {
						for _, v := range r.GetFloatPolicy() {
							for kk, vv := range v.GetRules() {
								if kk == c.output.key {
									exist = true
									So(vv.Required, ShouldEqual, c.output.req)
									So(vv.HasDefault, ShouldEqual, c.output.hasDefault)
									if vv.HasDefault {
										So(vv.Default, ShouldEqual, c.output.def.(float64))
									}
								}
							}
						}
						So(exist, ShouldEqual, true)
					})
				}
			})
		}
		Convey("Test Config Policy getting defaults", func() {
			// get config policy defaults
			cfg := p.getDefaults()
			So(cfg, ShouldNotBeEmpty)
			So(cfg["StringWithDefault"], ShouldEqual, "sss")
			So(cfg["boolRequired"], ShouldEqual, true)
			So(cfg["float"], ShouldEqual, 12.1)
			So(cfg["integer"], ShouldEqual, 12)

			Convey("config rules without default should be skipped", func() {
				_, exist := cfg["StringWithoutDefault"]
				So(exist, ShouldBeFalse)
			})
		})
	})
}

type inputConfigPolicy struct {
	ns   []string
	key  string
	req  bool
	opts interface{}
}

type expectedConfigPolicy struct {
	ns         []string
	key        string
	req        bool
	hasDefault bool
	def        interface{}
}

type testCaseConfigPolicy struct {
	input  inputConfigPolicy
	output expectedConfigPolicy
}

func configPolicyTestCases() []testCaseConfigPolicy {
	tc := []testCaseConfigPolicy{
		// test stringRule with Default value
		{
			input: inputConfigPolicy{
				ns:   []string{"a", "b", "c"},
				key:  "StringWithDefault",
				req:  true,
				opts: SetDefaultString("sss"),
			},
			output: expectedConfigPolicy{
				ns:         []string{"a", "b", "c"},
				key:        "StringWithDefault",
				req:        true,
				hasDefault: true,
				def:        "sss",
			},
		},
		// test stringRule without Default value
		{
			input: inputConfigPolicy{
				key: "StringWithoutDefault",
				req: true,
			},
			output: expectedConfigPolicy{
				key:        "StringWithoutDefault",
				req:        true,
				hasDefault: false,
			},
		},
		// test boolRule required
		{
			input: inputConfigPolicy{
				ns:   []string{"a1", "b1", "c1"},
				key:  "boolRequired",
				req:  true,
				opts: SetDefaultBool(true),
			},
			output: expectedConfigPolicy{
				ns:         []string{"a1", "b1", "c1"},
				key:        "boolRequired",
				req:        true,
				hasDefault: true,
				def:        true,
			},
		},
		// test boolRule not required
		{
			input: inputConfigPolicy{
				key:  "boolNotRequired",
				req:  false,
				opts: SetDefaultBool(true),
			},
			output: expectedConfigPolicy{
				key:        "boolNotRequired",
				req:        false,
				hasDefault: true,
				def:        true,
			},
		},
		// test floatRule
		{
			input: inputConfigPolicy{
				ns:   []string{"a2", "b2", "c2"},
				key:  "float",
				req:  true,
				opts: SetDefaultFloat(12.1),
			},
			output: expectedConfigPolicy{
				ns:         []string{"a2", "b2", "c2"},
				key:        "float",
				req:        true,
				hasDefault: true,
				def:        12.1,
			},
		},
		// test integerRule
		{
			input: inputConfigPolicy{
				ns:   []string{"a3", "b3", "c3"},
				key:  "integer",
				req:  true,
				opts: SetDefaultInt(12),
			},
			output: expectedConfigPolicy{
				ns:         []string{"a3", "b3", "c3"},
				key:        "integer",
				req:        true,
				hasDefault: true,
				def:        int64(12),
			},
		},
	}
	return tc
}
