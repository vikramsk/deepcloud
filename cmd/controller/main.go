package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vikramsk/deepcloud/pkg/controller"
)

var (
	defaultPort = os.Getenv("SVC_PORT")
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "controller: could not start service. err: %+v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flagset := flag.NewFlagSet("ctrl", flag.ExitOnError)
	var (
		port = flagset.String("svc.port", defaultPort, "controller: service port")
	)

	if err := flagset.Parse(args); err != nil {
		return err
	}

	csp, err := controller.InitControllerServiceProvider()
	if err != nil {
		return err
	}

	var logger log.Logger
	{
		logLevel := level.AllowInfo()
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = level.NewFilter(logger, logLevel)
	}

	apiDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "deepcloud",
		Name:      "api_request_duration_seconds",
		Help:      "API request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "status_code"})

	prometheus.MustRegister(apiDuration)
	api := controller.NewAPI(logger, csp, apiDuration)

	apiListener, err := net.Listen("tcp", ":"+*port)

	go interrupt(apiListener)

	mux := http.NewServeMux()
	mux.Handle("/", api)
	mux.Handle("/metrics", promhttp.Handler())
	return nil
}

func interrupt(apiListener net.Listener) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		fmt.Println("received signal: %s", sig)
		apiListener.Close()
	}
}
