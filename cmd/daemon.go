package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/qonto/upgrade-manager/config"
	"github.com/qonto/upgrade-manager/internal/app"
	"github.com/qonto/upgrade-manager/internal/build"
	"github.com/qonto/upgrade-manager/internal/infra/http"
	"github.com/qonto/upgrade-manager/internal/infra/kubernetes"
	"go.uber.org/zap"
)

func Run() error {
	registry := prometheus.NewRegistry()
	err := registry.Register(collectors.NewGoCollector())
	if err != nil {
		return err
	}
	zapConfig := zap.NewProductionConfig()
	if debug {
		zapConfig.Level.SetLevel(zap.DebugLevel)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}
	defer func() {
		err = logger.Sync()
	}()

	logger.Info(build.VersionMessage())
	signals := make(chan os.Signal, 1)
	errChan := make(chan error)
	if err != nil {
		return err
	}
	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM)
	config, err := config.Load(configFilePath)
	if err != nil {
		return err
	}

	server, err := http.New(registry, logger, config.HTTP)
	if err != nil {
		return err
	}

	err = server.Start()
	if err != nil {
		return err
	}

	client, err := kubernetes.NewClient(logger)
	if err != nil {
		return err
	}

	a, err := app.New(logger, registry, client, config)
	if err != nil {
		return err
	}

	a.Start()

	go func() {
		for sig := range signals {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				logger.Info(fmt.Sprintf("Received signal %s, shutdown", sig))
				signal.Stop(signals)
				a.Stop()
				err := server.Stop()
				errChan <- err
			}
		}
	}()
	exitErr := <-errChan
	return exitErr
}
