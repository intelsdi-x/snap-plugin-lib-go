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

// boolRule defines a type to contain Bool specific rule data.
type boolRule struct {
	Key        string
	Default    bool
	HasDefault bool
	Required   bool
}

type boolRuleOpt func(*boolRule)

// SetDefaultBool Allows easy setting of the Default value for an boolRule.
// Usage:
//		NewBoolRule(key, req, config.SetDefaultBool(default))
func SetDefaultBool(in bool) boolRuleOpt {
	return func(i *boolRule) {
		i.Default = in
		i.HasDefault = true
	}
}

// NewBoolRule returns a new boolRule with the specified args.
// The required arguments are key(string), req(bool)
// and optionally:
//		config.SetDefaultBool(bool)
func NewBoolRule(key string, req bool, opts ...boolRuleOpt) (boolRule, error) {
	if key == "" {
		return boolRule{}, fmt.Errorf(errEmptyKey)
	}
	rule := boolRule{
		Key:      key,
		Required: req,
	}

	for _, opt := range opts {
		opt(&rule)
	}

	return rule, nil
}
