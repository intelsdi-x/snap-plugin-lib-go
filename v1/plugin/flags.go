package plugin

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli"
)

var (
	flConfig = cli.StringFlag{
		Name:  "config",
		Usage: "config to use in JSON format",
	}
	// If no port was provided we let the OS select a port for us.
	// This is safe as address is returned in the Response and keep
	// alive prevents unattended plugins.
	flPort = cli.StringFlag{
		Name:  "port",
		Usage: "port GRPC will listen on",
	}
	flLogLevel = cli.IntFlag{
		Name:  "log-level",
		Usage: "log level - 0:panic 1:fatal 2:error 3:warn 4:info 5:debug",
		Value: 2,
	}
	flPprof = cli.BoolFlag{
		Name:  "pprof",
		Usage: "enable pprof",
	}
	flTLS = cli.BoolFlag{
		Name:  "tls",
		Usage: "enable TLS",
	}
	flCertPath = cli.StringFlag{
		Name:  "cert-path",
		Usage: "necessary to provide when TLS enabled",
	}
	flKeyPath = cli.StringFlag{
		Name:  "key-path",
		Usage: "necessary to provide when TLS enabled",
	}
	flRootCertPaths = cli.StringFlag{
		Name:  "root-cert-paths",
		Usage: fmt.Sprintf("root paths separated by '%c'", filepath.ListSeparator),
	}
	flStandAlone = cli.BoolFlag{
		Name:  "stand-alone",
		Usage: "enable stand alone plugin",
	}
	flHTTPPort = cli.IntFlag{
		Name:  "stand-alone-port",
		Usage: "specify http port when stand-alone is set",
		Value: 8181,
	}
)
