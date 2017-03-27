package plugin

import "github.com/urfave/cli"

var (
	flConfig = cli.StringFlag{
		Name:        "configInput",
		Value:       "",
		Usage:       "config file to use",
		Destination: &configIn,
	}
	// If no port was provided we let the OS select a port for us.
	// This is safe as address is returned in the Response and keep
	// alive prevents unattended plugins.
	flPort = cli.StringFlag{
		Name:        "port",
		Usage:       "port to listen on",
		Destination: &arg.ListenPort,
	}
	// If PingTimeoutDuration was provided we set it
	flPingTimeout = cli.DurationFlag{
		Name:        "pingTimeout",
		Usage:       "how much time must elapse before a lack of Ping results in a timeout",
		Destination: &arg.PingTimeoutDuration,
	}
	flPprof = cli.BoolFlag{
		Name:        "pprof",
		Hidden:      false,
		Usage:       "set pprof",
		Destination: &arg.Pprof,
	}
	flTLS = cli.BoolFlag{
		Name:        "tls",
		Hidden:      false,
		Usage:       "enable TLS",
		Destination: &arg.TLSEnabled,
	}
	flCertPath = cli.StringFlag{
		Name:        "certPath",
		Usage:       "necessary to provide when TLS enabled",
		Destination: &arg.CertPath,
	}
	flKeyPath = cli.StringFlag{
		Name:        "keyPath",
		Usage:       "necessary to provide when TLS enabled",
		Destination: &arg.KeyPath,
	}
)
