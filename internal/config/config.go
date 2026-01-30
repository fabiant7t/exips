package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	DefaultServiceName = "exips"
	DefaultNamespace   = "exips"
)

type config struct {
	ServiceName string
	Namespace   string
	KubeConfig  string
	Interval    time.Duration
	Resync      time.Duration
	Debug       bool
}

func (cfg *config) Client() (kubernetes.Interface, error) {
	var restConfig *rest.Config
	if cfg.KubeConfig != "" {
		rc, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
		if err != nil {
			slog.Error("error loading kubeconfig", "err", err, "kubeconfig", cfg.KubeConfig)
			return nil, err
		}
		restConfig = rc
	} else {
		rc, err := rest.InClusterConfig()
		if err != nil {
			slog.Error("error loding in cluster config", "err", err)
			return nil, err
		}
		restConfig = rc
	}
	return kubernetes.NewForConfig(restConfig)
}

func New() (*config, error) {
	cfg := &config{
		ServiceName: DefaultServiceName,
		Namespace:   DefaultNamespace,
		Interval:    15 * time.Second,
		Resync:      1 * time.Minute,
		Debug:       false,
	}
	if v := os.Getenv("SERVICENAME"); v != "" {
		cfg.ServiceName = v
	}
	if v := os.Getenv("NAMESPACE"); v != "" {
		cfg.Namespace = v
	}
	if v := os.Getenv("KUBECONFIG"); v != "" {
		cfg.KubeConfig = v
	}
	if v := os.Getenv("INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		cfg.Interval = d
	}
	if v := os.Getenv("RESYNC"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		cfg.Resync = d
	}
	if v := os.Getenv("DEBUG"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		cfg.Debug = b
	}
	return cfg, nil
}
