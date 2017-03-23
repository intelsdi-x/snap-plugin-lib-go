package plugin

import "github.com/urfave/cli"

var (
	flConfig = cli.StringFlag{
		Name:        "config",
		Value:       "",
		Usage:       "config file to use",
		Destination: &config,
	}
	// If no port was provided we let the OS select a port for us.
	// This is safe as address is returned in the Response and keep
	// alive prevents unattended plugins.
	flPort = cli.StringFlag{
		Name:        "port",
		Usage:       "port to listen on",
		Destination: &listenPort,
	}
	// If PingTimeoutDuration was provided we set it
	flPingTimeout = cli.DurationFlag{
		Name:        "pingTimeout",
		Usage:       "how much time must elapse before a lack of Ping results in a timeout",
		Destination: &PingTimeoutDurationDefault,
	}
	flPprof = cli.BoolFlag{
		Name:        "pprof",
		Hidden:      false,
		Usage:       "set pprof",
		Destination: &Pprof,
	}
)
