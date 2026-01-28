package config

import (
	"log/slog"
	"os"
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
	Resync      time.Duration
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
		Resync:      15 * time.Second,
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
	if v := os.Getenv("RESYNC"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		cfg.Resync = d
	}
	return cfg, nil
}
