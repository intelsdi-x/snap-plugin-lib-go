## Snap Plugin Go Library Examples
Here you will find example plugins that cover the basics for writing collector, processor, and publisher plugins.

## Build Plugins & Use with Snap

To get these example collector, processor, and publisher plugins to build properly and work with Snap you will need to have [glide](https://glide.sh/) installed in your $PATH. You should also add snapctl and snapd in your $PATH. 

To test these plugins with Snap, you will need to have [Snap](https://github.com/intelsdi-x/snap) installed, check out these docs for [Snap setup details](https://github.com/intelsdi-x/snap/blob/master/docs/BUILD_AND_TEST.md#getting-started).

###1. Get the plugin library repo:
`go get github.com/intelsdi-x/snap-plugin-lib-go/...` will add the repo to your $GOPATH

###2. Go to the snap-plugin-lib-go folder and update to the newest versions of the package with `glide up`:

```
$ cd snap-plugin-lib-go
$ glide up
[INFO]	Downloading dependencies. Please wait...
[INFO]	--> Fetching updates for github.com/jtolds/gls.
[INFO]	--> Fetching updates for github.com/golang/protobuf.
[INFO]	--> Fetching updates for github.com/smartystreets/assertions.
[INFO]	--> Fetching updates for github.com/smartystreets/goconvey.
[INFO]	--> Fetching updates for google.golang.org/grpc.
[INFO]	--> Fetching updates for github.com/gopherjs/gopherjs.
[INFO]	--> Fetching updates for golang.org/x/net.
[INFO]	--> Setting version for github.com/golang/protobuf to 888eb0692c857ec880338addf316bd662d5e630e.
[INFO]	--> Setting version for github.com/smartystreets/assertions to 443d812296a84445c202c085f19e18fc238f8250.
[INFO]	--> Setting version for github.com/smartystreets/goconvey to 995f5b2e021c69b8b028ba6d0b05c1dd500783db.
[INFO]	--> Setting version for github.com/gopherjs/gopherjs to 4b53e1bddba0e2f734514aeb6c02db652f4c6fe8.
[INFO]	--> Setting version for github.com/jtolds/gls to 8ddce2a84170772b95dd5d576c48d517b22cac63.
[INFO]	--> Setting version for google.golang.org/grpc to 0032a855ba5c8a3c8e0d71c2deef354b70af1584.
[INFO]	--> Setting version for golang.org/x/net to 154d9f9ea81208afed560f4cf27b4860c8ed1904.
[INFO]	Resolving imports
[INFO]	Downloading dependencies. Please wait...
[INFO]	Setting references for remaining imports
[INFO]	Exporting resolved dependencies...
[INFO]	--> Exporting github.com/gopherjs/gopherjs
[INFO]	--> Exporting github.com/golang/protobuf
[INFO]	--> Exporting github.com/jtolds/gls
[INFO]	--> Exporting github.com/smartystreets/assertions
[INFO]	--> Exporting github.com/smartystreets/goconvey
[INFO]	--> Exporting google.golang.org/grpc
[INFO]	--> Exporting golang.org/x/net
[INFO]	Replacing existing vendor dependencies
[INFO]	Project relies on 7 dependencies.
```

###3. Build the collector, processor, and/or publisher plugins in the examples folder.
    Use the `go build` command to generate the example binary files for the collector, processor, and publisher.
    option -o outputs the binary to the specified name 

```
$ go build -o example-collector examples/collector/main.go
$ go build -o example-processor examples/processor/main.go 
$ go build -o example-publisher examples/publisher/main.go 
```

###4. Run Snap, Load Plugins, and Run Tasks
You can now try [running Snap](https://github.com/intelsdi-x/snap#running-snap), [loading plugins](https://github.com/intelsdi-x/snap#load-plugins), and [running tasks](https://github.com/intelsdi-x/snap#running-tasks). You'll find a sample task.yml file for these examples at the bottom of this README.md.  

Below are some sample commands to try:

```
$ snapctl plugin load example-collector
Plugin loaded
Name: test-rand-collector
Version: 1
Type: collector
Signed: false
Loaded Time: Fri, 23 Sep 2016 17:41:44 PDT

$ snapctl plugin list
NAME 			 VERSION 	 TYPE 		 SIGNED 	 STATUS 	 LOADED TIME
test-rand-collector 	 1 		 collector 	 false 		 loaded 	 Fri, 23 Sep 2016 17:41:44 PDT

$ snapctl metric list
NAMESPACE 		 VERSIONS
/random/float 		 1
/random/integer 	 1
/random/string 		 1

$ snapctl plugin load example-processor
Plugin loaded
Name: test-reverse-processor
Version: 1
Type: processor
Signed: false
Loaded Time: Fri, 23 Sep 2016 17:44:12 PDT

$ snapctl plugin load example-publisher
Plugin loaded
Name: test-file-publisher
Version: 1
Type: publisher
Signed: false
Loaded Time: Fri, 23 Sep 2016 17:44:23 PDT

$ snapctl plugin list
NAME 			 VERSION 	 TYPE 		 SIGNED 	 STATUS 	 LOADED TIME
test-rand-collector 	 1 		 collector 	 false 		 loaded 	 Fri, 23 Sep 2016 17:41:44 PDT
test-reverse-processor 	 1 		 processor 	 false 		 loaded 	 Fri, 23 Sep 2016 17:44:12 PDT
test-file-publisher 	 1 		 publisher 	 false 		 loaded 	 Fri, 23 Sep 2016 17:44:23 PDT

```

### Create a task.
You can create a task.yml file for these examples using the following code. 
Discover [how tasks work](https://github.com/intelsdi-x/snap/blob/master/docs/TASKS.md).

task.yml

```---
  version: 1
  schedule:
    type: "simple"
    interval: "1s"
  max-failures: 10
  workflow:
    collect:
      metrics:
        /random/float: {}
        /random/integer: {}
        /random/string: {}
      config:
      process:
        -
          plugin_name: "test-reverse-processor"
          process: null
          publish:
            -
              plugin_name: "test-file-publisher"
              config:
                file: "/tmp/snap_published_grpc_file.log"

```
