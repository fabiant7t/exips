package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/fabiant7t/exips/internal/config"
	"github.com/fabiant7t/exips/internal/node/registry"
	"github.com/fabiant7t/exips/internal/service"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuiltTime = "unknown"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		slog.Error("error in configuration", "err", err)
		os.Exit(1)
	}
	if cfg.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	slog.Info("exips",
		"version", Version,
		"author", "Fabian Topfstedt",
		"url", "https://github.com/fabiant7t/exips",
		"licence", "MIT",
		"commit", Commit,
		"built_time", BuiltTime,
	)
	slog.Info("Configuration",
		"service_name", cfg.ServiceName,
		"service_namespace", cfg.ServiceNamespace,
		"kube_config", cfg.KubeConfig,
		"interval", cfg.Interval,
		"resync", cfg.Resync,
		"debug", cfg.Debug,
	)

	client, err := cfg.Client()
	if err != nil {
		slog.Error("error creating kubernetes client", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reg := registry.New()

	var wg sync.WaitGroup
	wg.Go(func() {
		if err := reg.Run(ctx, client, cfg.Resync); err != nil {
			slog.Error("error syncing registry", "err", err)
			return
		}
	})
	wg.Go(func() {
		ticker := time.NewTicker(cfg.Interval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				externalIPs := reg.ParseExternalIPs()
				externalIPStrings := make([]string, len(externalIPs))
				for i, ip := range externalIPs {
					externalIPStrings[i] = ip.String()
				}

				existingSvc, err := service.Get(ctx, client, cfg.ServiceName, cfg.ServiceNamespace)
				if err != nil {
					slog.Error("error getting service", "err", err, "name", cfg.ServiceName, "namespace", cfg.ServiceNamespace)
					return
				}
				if existingSvc == nil { // create service
					svc := service.New(cfg.ServiceName, externalIPStrings)
					if err := service.Apply(ctx, client, svc, cfg.ServiceNamespace); err == nil {
						slog.Info("Service created", "name", cfg.ServiceName, "namespace", cfg.ServiceNamespace, "external_ips", svc.Spec.ExternalIPs)
					} else {
						slog.Error("error creating service", "err", err, "name", cfg.ServiceName, "namespace", cfg.ServiceNamespace)
					}
				} else { // service exists, may require update
					if upToDate := slices.Equal(existingSvc.Spec.ExternalIPs, externalIPStrings); upToDate {
						slog.Debug("Service is already up to date", "name", cfg.ServiceName, "namespace", cfg.ServiceNamespace, "external_ips", existingSvc.Spec.ExternalIPs)
					} else { // must update
						svc := service.New(cfg.ServiceName, externalIPStrings)
						if err := service.Apply(ctx, client, svc, cfg.ServiceNamespace); err == nil {
							slog.Info("Service updated", "name", cfg.ServiceName, "namespace", cfg.ServiceNamespace, "external_ips", svc.Spec.ExternalIPs)
						} else {
							slog.Error("error updating service", "err", err, "name", cfg.ServiceName, "namespace", cfg.ServiceNamespace)
						}
					}
				}
			}
		}
	})

	wg.Wait()
}
