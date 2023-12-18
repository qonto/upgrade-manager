package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Host              string `validate:"required"`
	Port              uint32 `validate:"required,gt=1024,lt=65535"`
	WriteTimeout      int    `yaml:"write-timeout" validate:"gt=-1,lt=60"`
	ReadTimeout       int    `yaml:"read-timeout" validate:"gt=-1,lt=60"`
	ReadHeaderTimeout int    `yaml:"read-header-timeout" validate:"gt=-1,lt=60"`
}

type HTTPServer struct {
	server *http.Server
	Engine *gin.Engine
	logger *slog.Logger

	wg               sync.WaitGroup
	registry         *prometheus.Registry
	requestHistogram *prometheus.HistogramVec
	responseCounter  *prometheus.CounterVec
}

func healthz(context *gin.Context) {
	context.JSON(200, "ok")
}

func New(registry *prometheus.Registry, logger *slog.Logger, config Config) (*HTTPServer, error) {
	var defaultTimeout int = 10
	engine := gin.New()
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	if config.WriteTimeout == 0 {
		config.WriteTimeout = defaultTimeout
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = defaultTimeout
	}
	if config.ReadHeaderTimeout == 0 {
		config.ReadHeaderTimeout = defaultTimeout
	}
	server := &http.Server{
		WriteTimeout:      time.Duration(config.WriteTimeout) * time.Second,
		ReadTimeout:       time.Duration(config.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(config.ReadHeaderTimeout) * time.Second,
		Addr:              address,
		Handler:           engine,
	}
	respCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Count the number of HTTP responses.",
		},
		[]string{"method", "status", "rule"})

	buckets := []float64{
		0.05, 0.1, 0.2, 0.4, 0.8, 1,
		1.5, 2, 3, 5,
	}

	reqHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requests_duration_second",
			Help:    "Time to execute http requests",
			Buckets: buckets,
		},
		[]string{"method", "rule"})

	return &HTTPServer{
		server:           server,
		Engine:           engine,
		logger:           logger,
		requestHistogram: reqHistogram,
		responseCounter:  respCounter,
		registry:         registry,
	}, nil
}

func (h *HTTPServer) Start() error {
	h.logger.Info(fmt.Sprintf("Starting HTTP server on %s", h.server.Addr))
	err := h.registry.Register(h.responseCounter)
	if err != nil {
		return err
	}

	err = h.registry.Register(h.requestHistogram)
	if err != nil {
		return err
	}

	go func() {
		defer h.wg.Done()

		err = h.server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed { //nolint
			h.logger.Error(fmt.Sprintf("HTTP server error: %s", err.Error()))
			exitCode := 2
			os.Exit(exitCode)
		}
	}()
	h.wg.Add(1)
	h.Engine.GET("/healthz", healthz)
	var gatherer prometheus.Gatherer = h.registry
	h.Engine.GET("/metrics", gin.WrapH(promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})))
	return nil
}

func (h *HTTPServer) Stop() error {
	h.logger.Info("Stopping HTTP Server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint
	defer cancel()
	if err := h.server.Shutdown(ctx); err != nil {
		h.logger.Error(err.Error())
		return err
	}
	h.wg.Wait()
	return nil
}
