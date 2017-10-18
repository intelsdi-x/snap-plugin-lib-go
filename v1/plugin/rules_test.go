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

// Boolean tests

func TestBoolRule(t *testing.T) {
	ts := createPassTestCases()
	Convey("Test boolRule Passes", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range ts {
			Convey(fmt.Sprintf("Test boolRule %+v", c.input.key), func() {
				if c.input.opts != nil {
					err = p.AddNewBoolRule(c.input.ns, c.input.key, c.input.req, c.input.opts)
				} else {
					err = p.AddNewBoolRule(c.input.ns, c.input.key, c.input.req)
				}
				So(err, ShouldBeNil)
			})
		}
	})

	ts = createErrTestCases()
	Convey("Test boolRule Fails", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range ts {
			if c.input.opts != nil {
				err = p.AddNewBoolRule(c.input.ns, c.input.key, c.input.req, c.input.opts)
			} else {
				err = p.AddNewBoolRule(c.input.ns, c.input.key, c.input.req)
			}
			So(err, ShouldNotBeNil)
		}
	})
}

type boolInput struct {
	ns   []string
	key  string
	req  bool
	opts boolRuleOpt
}

type testCaseBool struct {
	input boolInput
}

func createPassTestCases() []testCaseBool {
	tc := []testCaseBool{
		{
			input: boolInput{
				key:  "abc",
				req:  true,
				opts: SetDefaultBool(true),
			},
		},
		{
			input: boolInput{
				key:  "abc1",
				req:  false,
				opts: SetDefaultBool(false),
			},
		},
		{
			input: boolInput{ns: []string{"plop"}, key: "xyz", req: true},
		},
		{
			input: boolInput{ns: []string{"plop"}, key: "abc2", req: false},
		},
	}
	return tc
}

func createErrTestCases() []testCaseBool {
	tc := []testCaseBool{
		{
			input: boolInput{req: true},
		},
		{
			input: boolInput{key: "", req: false},
		},
		{
			input: boolInput{
				key: "", req: false,
				opts: SetDefaultBool(false),
			},
		},
	}
	return tc
}

// Float tests

func TestFloatRule(t *testing.T) {
	tc := floatPassTestCases()
	Convey("Test floatRule Passes", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range tc {
			Convey(fmt.Sprintf("Test floatRule %+v", c.input.key), func() {
				if c.input.opts != nil {
					err = p.AddNewFloatRule(c.input.ns, c.input.key, c.input.req, c.input.opts...)
				} else {
					err = p.AddNewFloatRule(c.input.ns, c.input.key, c.input.req)
				}

				So(err, ShouldBeNil)
			})
		}
	})

	tc = floatErrTestCases()
	Convey("Test floatRule Fails", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				err = p.AddNewFloatRule(c.input.ns, c.input.key, c.input.req, c.input.opts...)
			} else {
				err = p.AddNewFloatRule(c.input.ns, c.input.key, c.input.req)
			}

			So(err, ShouldNotBeNil)
		}
	})
}

type floatInput struct {
	ns   []string
	key  string
	req  bool
	opts []floatRuleOpt
}

type testCaseFloat struct {
	input floatInput
}

func floatErrTestCases() []testCaseFloat {
	tc := []testCaseFloat{
		{
			input: floatInput{req: true},
		},
		{
			input: floatInput{key: "", req: false},
		},
	}
	return tc
}

func floatPassTestCases() []testCaseFloat {
	tc := []testCaseFloat{
		{
			input: floatInput{key: "xyz", req: true},
		},
		{
			input: floatInput{key: "abc", req: false},
		},
		{
			input: floatInput{
				key:  "xyz1",
				req:  true,
				opts: []floatRuleOpt{SetDefaultFloat(30.1)}},
		},
		{
			input: floatInput{
				key:  "xyz2",
				req:  true,
				opts: []floatRuleOpt{SetMaxFloat(32.1)}},
		},
		{
			input: floatInput{
				key:  "xyz3",
				req:  false,
				opts: []floatRuleOpt{SetMinFloat(12.1)}},
		},
	}
	return tc
}

// Integer tests

func TestIntegerRule(t *testing.T) {
	tc := integerPassTestCases()
	Convey("Test integerRule Passes", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				err = p.AddNewIntRule(c.input.ns, c.input.key, c.input.req, c.input.opts...)
			} else {
				err = p.AddNewIntRule(c.input.ns, c.input.key, c.input.req)
			}
			So(err, ShouldBeNil)
		}
	})

	tc = integerErrTestCases()
	Convey("Test integerRule Fails", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				err = p.AddNewIntRule(c.input.ns, c.input.key, c.input.req, c.input.opts...)
			} else {
				err = p.AddNewIntRule(c.input.ns, c.input.key, c.input.req)
			}
			So(err, ShouldNotBeNil)
		}
	})
}

type integerInput struct {
	ns   []string
	key  string
	req  bool
	opts []integerRuleOpt
}

type testCaseInteger struct {
	input integerInput
}

func integerPassTestCases() []testCaseInteger {
	tc := []testCaseInteger{
		{
			input: integerInput{key: "xyz", req: true},
		},
		{
			input: integerInput{key: "abc", req: false},
		},
		{
			input: integerInput{
				ns:   []string{"plop"},
				key:  "xyz1",
				req:  true,
				opts: []integerRuleOpt{SetDefaultInt(30)}},
		},
		{
			input: integerInput{
				ns:   []string{"plop"},
				key:  "xyz2",
				req:  true,
				opts: []integerRuleOpt{SetMaxInt(64)}},
		},
		{
			input: integerInput{
				ns:   []string{"plop", "world"},
				key:  "xyz3",
				req:  true,
				opts: []integerRuleOpt{SetMinInt(5)}},
		},
	}
	return tc
}

func integerErrTestCases() []testCaseInteger {
	tc := []testCaseInteger{
		{
			input: integerInput{req: true},
		},
		{
			input: integerInput{key: "", req: false},
		},
		{
			input: integerInput{
				ns:   []string{"plop"},
				key:  "",
				req:  false,
				opts: []integerRuleOpt{SetMinInt(5)}},
		},
	}
	return tc
}

// String tests

func TestStringRule(t *testing.T) {
	tc := stringPassTestCases()
	Convey("Test stringRule Passes", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range tc {
			Convey(fmt.Sprintf("Test stringRule %+v", c.input.key), func() {
				if c.input.opts != nil {
					err = p.AddNewStringRule(c.input.ns, c.input.key, c.input.req, c.input.opts)
				} else {
					err = p.AddNewStringRule(c.input.ns, c.input.key, c.input.req)
				}
				So(err, ShouldBeNil)
			})
		}
	})

	tc = stringErrTestCases()
	Convey("Test stringRule Fails", t, func() {
		p := NewConfigPolicy()
		var err error
		for _, c := range tc {
			if c.input.opts != nil {
				err = p.AddNewStringRule(c.input.ns, c.input.key, c.input.req, c.input.opts)
			} else {
				err = p.AddNewStringRule(c.input.ns, c.input.key, c.input.req)
			}
			So(err, ShouldNotBeNil)
		}
	})
}

type stringInput struct {
	ns   []string
	key  string
	req  bool
	opts stringRuleOpt
}

type testCaseString struct {
	input stringInput
}

func stringPassTestCases() []testCaseString {
	tc := []testCaseString{
		{
			input: stringInput{key: "xyz", req: true},
		},
		{
			input: stringInput{key: "abc", req: false},
		},
		{
			input: stringInput{
				key:  "deer",
				req:  true,
				opts: SetDefaultString("123")},
		},
		{
			input: stringInput{
				key:  "racoon",
				req:  false,
				opts: SetDefaultString("aaa")},
		},
	}
	return tc
}

func stringErrTestCases() []testCaseString {
	tc := []testCaseString{
		{
			input: stringInput{req: true},
		},
		{
			input: stringInput{key: "", req: false},
		},
		{
			input: stringInput{
				key:  "",
				req:  true,
				opts: SetDefaultString("xyz")},
		},
	}
	return tc
}
