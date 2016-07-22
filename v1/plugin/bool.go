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

// BoolRule defines a type to contain Bool specific rule data.
type BoolRule struct {
	Key        string
	Default    bool
	HasDefault bool
	Required   bool
}

type boolRuleOpt func(*BoolRule)

// SetDefaultBool Allows easy setting of the Default value for an BoolRule.
// Usage:
//		NewBoolRule(key, req, config.SetDefaultBool(default))
func SetDefaultBool(in bool) boolRuleOpt {
	return func(i *BoolRule) {
		i.Default = in
		i.HasDefault = true
	}
}

// NewBoolRule returns a new BoolRule with the specified args.
// The required arguments are key(string), req(bool)
// and optionally:
//		config.SetDefaultBool(bool)
func NewBoolRule(key string, req bool, opts ...boolRuleOpt) (BoolRule, error) {
	if key == "" {
		return BoolRule{}, fmt.Errorf(errEmptyKey)
	}
	rule := BoolRule{
		Key:      key,
		Required: req,
	}

	for _, opt := range opts {
		opt(&rule)
	}

	return rule, nil
}
