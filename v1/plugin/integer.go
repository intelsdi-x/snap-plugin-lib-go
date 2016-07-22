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

import "fmt"

// IntegerRule defines a type to contain Integer specific rule data.
type IntegerRule struct {
	Key string

	Default    int64
	HasDefault bool

	Required bool

	Minimum int64
	HasMin  bool

	Maximum int64
	HasMax  bool
}

type integerRuleOpt func(*IntegerRule)

// SetDefaultInt Allows easy setting of the Default value for an IntegerRule.
// Usage:
//	//	NewIntegerRule(key, req, config.SetDefaultInt(default))
func SetDefaultInt(in int64) integerRuleOpt {
	return func(i *IntegerRule) {
		i.Default = in
		i.HasDefault = true
	}
}

// SetMaxInt Allows easy setting of the Max value for an IntegerRule.
// Usage:
//		NewIntegerRule(key, req, config.SetMaxInt(max))
func SetMaxInt(max int64) integerRuleOpt {
	return func(i *IntegerRule) {
		i.Maximum = max
		i.HasMax = true
	}
}

// SetMinInt Allows easy setting of the Min value for an IntegerRule.
// Usage:
//		NewIntegerRule(key, req, config.SetMinInt(min))
func SetMinInt(min int64) integerRuleOpt {
	return func(i *IntegerRule) {
		i.Minimum = min
		i.HasMin = true
	}
}

// NewIntegerRule returns a new IntegerRule with the specified args.
// The required arguments are key(string), req(bool)
// and optionally:
//		config.SetDefaultInt(int64),
//		config.SetMinInt(int64),
//		config.SetMaxInt(int64),
func NewIntegerRule(key string, req bool, opts ...integerRuleOpt) (IntegerRule, error) {
	if key == "" {
		return IntegerRule{}, fmt.Errorf(errEmptyKey)
	}
	rule := IntegerRule{
		Key:      key,
		Required: req,
	}

	for _, opt := range opts {
		opt(&rule)
	}

	return rule, nil
}
