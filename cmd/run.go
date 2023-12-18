package cmd

import (
	"fmt"
	"log/slog"
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
)

func Run() error {
	registry := prometheus.NewRegistry()
	err := registry.Register(collectors.NewGoCollector())
	if err != nil {
		return err
	}

	logger := buildLogger(logLevel, logFormat)

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

func buildLogger(level string, format string) *slog.Logger {
	programLevel := new(slog.LevelVar)
	switch level {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	options := &slog.HandlerOptions{Level: programLevel}
	switch format {
	case "text":
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, options))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
}
