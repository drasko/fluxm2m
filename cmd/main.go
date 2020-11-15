package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/fluxm2m/lwm2m"
	"github.com/mainflux/fluxm2m/lwm2m/api"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/plgd-dev/go-coap/v2"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	defPort      = "5683"
	defLogLevel  = "error"
	defClientTLS = "false"
	defCACerts   = ""

	envPort      = "FLUXM2M_PORT"
	envLogLevel  = "FLUXM2M_LOG_LEVEL"
	envClientTLS = "FLUXM2M_CLIENT_TLS"
	envCACerts   = "FLUXM2M_CA_CERTS"
)

type config struct {
	port      string
	logLevel  string
	clientTLS bool
	caCerts   string
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	svc := lwm2m.New()
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "coap_adapter",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "coap_adapter",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	errs := make(chan error, 2)

	go startHTTPServer(cfg.port, logger, errs)
	go startCoAPServer(cfg, svc, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("CoAP adapter terminated: %s", err))
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	return config{
		port:      mainflux.Env(envPort, defPort),
		logLevel:  mainflux.Env(envLogLevel, defLogLevel),
		clientTLS: tls,
		caCerts:   mainflux.Env(envCACerts, defCACerts),
	}
}

func startHTTPServer(port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("LwM2M service started, exposed port %s", port))
	errs <- http.ListenAndServe(p, api.MakeHTTPHandler())
}

func startCoAPServer(cfg config, svc lwm2m.Service, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.port)
	logger.Info(fmt.Sprintf("LwM2M service started, exposed port %s", cfg.port))
	errs <- coap.ListenAndServe("udp", p, api.MakeCoAPHandler(svc, logger))
}
