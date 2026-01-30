package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fabiant7t/exips/internal/config"
	"github.com/fabiant7t/exips/internal/node/registry"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		slog.Error("error in configuration", "err", err)
		os.Exit(1)
	}
	slog.Info("Configuration",
		"service_name", cfg.ServiceName,
		"namespace", cfg.Namespace,
		"kube_config", cfg.KubeConfig,
		"resync", cfg.Resync,
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

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				externalIPs := reg.ParseExternalIPs()
				log.Println(externalIPs)
			}
		}
	}()

	wg.Wait()
}
