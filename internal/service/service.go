package service

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// New returns a Service that acts as a sidekick to the ingress controller,
// exposing the cluster’s external IPs. The Service is intentionally created
// without a selector so it is not backed by any Pods.
func New(name string, externalIPs []string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "dummy",
					Port:       6942,
					TargetPort: intstr.FromInt(6942),
				},
			},
			ExternalIPs: externalIPs,
		},
	}
}

// Apply creates or updates the given service in the given namespace.
func Apply(ctx context.Context, client kubernetes.Interface, svc *corev1.Service, namespace string) error {
	data, err := json.Marshal(svc)
	if err != nil {
		return fmt.Errorf("error marshaling service to JSON: %w ", err)
	}

	yes := true
	_, err = client.CoreV1().Services(namespace).Patch(
		ctx,
		svc.Name,
		types.ApplyPatchType,
		data,
		metav1.PatchOptions{
			FieldManager: "exips-service",
			Force:        &yes,
		},
	)
	if err != nil {
		return fmt.Errorf("error patching service: %w ", err)
	}
	return nil
}

// Get fetches the Service by name and namespace.
// Returns nil and no error if the Service does not exist.
func Get(ctx context.Context, client kubernetes.Interface, name, namespace string) (*corev1.Service, error) {
	svc, err := client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return svc, nil
}
