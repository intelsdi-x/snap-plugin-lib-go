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

package file

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

// FPublisher is a testing publisher.
type FPublisher struct {
}

/*
	GetConfigPolicy() returns the configPolicy for your plugin.

	A config policy is how users can provide configuration info to
	plugin. Here you define what sorts of config info your plugin
	needs and/or requires.
*/
func (f FPublisher) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	return *policy, nil
}

// Publish test publish function
func (f FPublisher) Publish(mts []plugin.Metric, cfg plugin.Config) error {
	file, err := cfg.GetString("file")
	if err != nil {
		return err
	}
	if val, err := cfg.GetBool("return_error"); err == nil && val {
		return errors.New("Houston we have a problem")
	}
	fileHandle, _ := os.Create(file)
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()

	for _, m := range mts {
		fmt.Fprintf(writer, "%s|%v|%d|%s|%s|%s|%v|%v\n",
			m.Namespace.Strings(),
			m.Data,
			m.Version,
			m.Unit,
			m.Description,
			m.Timestamp,
			m.Tags,
			m.Config,
		)
	}
	writer.Flush()

	return nil
}
