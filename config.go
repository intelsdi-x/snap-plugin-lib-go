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

type Config map[string]interface{}

func (c Config) String(key string) (string, error) {
	var (
		val    interface{}
		strout string
		ok     bool
	)
	if val, ok = c[key]; !ok {
		return strout, fmt.Errorf("config item %s not found", key)
	}
	if strout, ok = val.(string); !ok {
		return strout, fmt.Errorf("config item %s is not a string", key)
	}
	return strout, nil
}

type ConfigTree struct{}

type ConfigPolicy struct{}
