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

// StringRule defines a type to contain String specific rule data.
type StringRule struct {
	Key        string
	Default    string
	HasDefault bool
	Required   bool
}

type stringRuleOpt func(*StringRule)

// SetDefaultString Allows easy setting of the Default value for an StringRule.
// Usage:
//		NewStringRule(key, req, config.SetDefaultString(default))
func SetDefaultString(in string) stringRuleOpt {
	return func(i *StringRule) {
		i.Default = in
		i.HasDefault = true
	}
}

// NewStringRule returns a new StringRule with the specified args.
// The required arguments are key(string), req(bool)
// and optionally:
//		config.SetDefaultString(string)
func NewStringRule(key string, req bool, opts ...stringRuleOpt) (StringRule, error) {
	if key == "" {
		return StringRule{}, fmt.Errorf(errEmptyKey)
	}
	rule := StringRule{
		Key:      key,
		Required: req,
	}

	for _, opt := range opts {
		opt(&rule)
	}

	return rule, nil
}
