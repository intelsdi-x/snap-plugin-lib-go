package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"net/http/pprof"

	logger "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
)

var (
	pprofPort  = "0"
	configIn   = ""
	standAlone = false
	httpPort   = 0
	arg        = Arg{
		LogLevel:            uint8(2),
		PingTimeoutDuration: time.Millisecond * 1500,
		ListenPort:          "0",
		Pprof:               false,
		CertPath:            "",
		KeyPath:             "",
		TLSEnabled:          false,
	}
)

// Arg represents arguments passed to startup of Plugin
type Arg struct {
	// Plugin log level, see logrus.Loglevel
	LogLevel uint8
	// Ping timeout duration
	PingTimeoutDuration time.Duration

	// The listen port
	ListenPort string

	// enable pprof
	Pprof bool

	// Path to TLS certificate file for a TLS server
	CertPath string

	// Path to TLS private key file for a TLS server
	KeyPath string

	// Flag requesting server to establish TLS channel
	TLSEnabled bool
}

// getArgs returns plugin args or default ones
func getArgs() (*Arg, error) {
	osArgs := libInputOutput.readOSArgs()
	// default parameters - can be parsed as JSON
	paramStr := "{}"
	if len(osArgs) > 1 && osArgs[1] != "" {
		paramStr = osArgs[1]
	}
	err := json.Unmarshal([]byte(paramStr), &arg)
	logger.Errorf("!!!!!!!!! timeout: %v", arg.PingTimeoutDuration)

	logger.Errorf("!!!!!!!!! err=%v", err)

	if arg.Pprof {
		return &arg, getPort()
	}

	return &arg, nil
}

func getPort() error {
	router := httprouter.New()
	router.GET("/debug/pprof/", index)
	router.GET("/debug/pprof/block", index)
	router.GET("/debug/pprof/goroutine", index)
	router.GET("/debug/pprof/heap", index)
	router.GET("/debug/pprof/threadcreate", index)
	router.GET("/debug/pprof/cmdline", cmdline)
	router.GET("/debug/pprof/profile", profile)
	router.GET("/debug/pprof/symbol", symbol)
	router.GET("/debug/pprof/trace", trace)
	addr, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	pprofPort = fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)

	go func() {
		log.Fatal(http.Serve(l, router))
	}()

	return nil
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Index(w, r)
}

func cmdline(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Cmdline(w, r)
}

func profile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Profile(w, r)
}

func symbol(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Symbol(w, r)
}

func trace(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Trace(w, r)
}
