#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2016 Intel Corporation
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH = $(shell uname -m)

default:
	$(MAKE) deps
	$(MAKE) examples
deps:
	bash -c "./scripts/deps.sh"
test:
	bash -c "./scripts/test.sh $(SNAP_TEST_TYPE)"
test-small:
	bash -c "./scripts/test.sh small"
test-medium:
	bash -c "./scripts/test.sh medium"
test-large:
	bash -c "./scripts/test.sh large"
test-all:	
	$(MAKE) test-medium
	$(MAKE) test-small
	$(MAKE) test-large
clean:
	rm -rf build

# Build only example plugins
examples:
	bash -c "./scripts/build_examples.sh"

.PHONY: examples
