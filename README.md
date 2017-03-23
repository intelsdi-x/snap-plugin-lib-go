## Snap Plugin Library for Go

This is a library for writing plugins in Go for the [Snap telemetry framework](https://github.com/intelsdi-x/snap). 

Snap has three different plugin types and for instructions on how to write a plugin check out the [collector](/examples/snap-plugin-collector-rand/README.md), [processor](examples/snap-plugin-processor-reverse/README.md), and [publisher](examples/snap-plugin-publisher-file/README.md) plugin docs.

Before writing a Snap plugin:

* See if one already exists in the [Plugin Catalog](https://github.com/intelsdi-x/snap/blob/master/docs/PLUGIN_CATALOG.md) 
* See if someone mentioned it in the [plugin wishlist](https://github.com/intelsdi-x/snap/blob/master/docs/PLUGIN_CATALOG.md#wishlist)

If you do decide to write a plugin check out the [plugin authoring docs](https://github.com/intelsdi-x/snap/blob/master/docs/PLUGIN_AUTHORING.md#plugin-authoring) and let us know you are working on one!

## Brief Overview of Snap Architecture

Snap is an open and modular telemetry framework designed to simplify the collection, processing and publishing of data through a single HTTP based API. Plugins provide the functionality of collection, processing and publishing and can be loaded/unloaded, upgraded and swapped without requiring a restart of the Snap daemon.

A Snap plugin is a program that responds to a set of well defined [gRPC](http://www.grpc.io/) services with parameters and return types specified as protocol buffer messages (see [plugin.proto](https://github.com/intelsdi-x/snap/blob/master/control/plugin/rpc/plugin.proto)). The Snap daemon handshakes with the plugin over stdout and then communicates over gRPC.


## Snap Plugin Go Library Examples
You will find [example plugins](examples) that cover the basics for writing collector, processor, and publisher plugins in the examples folder.


## Snap Diagnostics
Snap plugins using plugin-lib-go can be run independent of Snap to show their current running diagnostics. This diagnostic information includes:
* Warning if dependencies not met
* Config policy
    * Warning if config items required and not provided
* Collectable metrics
* Metric catalog
* Runtime details
    * Plugin version
    * RPC type and version
    * OS, architecture
    * Golang version
* How long it took to run each of these diagnostics

### Running Diagnostics
Running plugin diagnostics is easy! Simply build the plugin, then run the executable `$./build/darwin/x86_64/examples/snap-plugin-collector-rand`. When ran on its own, it will show a warning if a config is required for the plugin to load. 

### Global Flags
For specific details and to see all the options when running, run the plugin with the `-help` flag. The flag options are:
```--config value       config file to use
   --port value         port to listen on
   --pingTimeout value  how much time must elapse before a lack of Ping results in a timeout (default: 0s)
   --pprof              something about pprof
   --help, -h           show help
   --version, -v        print the version
```

### Config flag
When `-config` is set, it expects a parameter in the form of a JSON. This is of the form `'{}'`. An example config is: `-config '{\"key\":\"kelly\", \"spirit-animal\":\"coatimundi\"}'`.