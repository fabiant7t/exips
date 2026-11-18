package main

import (
	"net/netip"
	"slices"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestServiceLoopRecoversAfterGetError(t *testing.T) {
	const name, namespace = "exips", "ingress"
	client := fake.NewClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       corev1.ServiceSpec{ExternalIPs: []string{"192.0.2.99"}},
	})
	gets := 0
	client.PrependReactor("get", "services", func(action clienttesting.Action) (bool, runtime.Object, error) {
		if action.GetNamespace() != namespace || action.(clienttesting.GetAction).GetName() != name {
			t.Fatalf("unexpected Service GET: %v", action)
		}
		gets++
		if gets == 1 {
			return true, nil, apierrors.NewServiceUnavailable("first GET failed")
		}
		return false, nil, nil
	})

	ticks := make(chan time.Time, 2)
	ticks <- time.Time{}
	ticks <- time.Time{}
	close(ticks)
	runServiceLoop(t.Context(), client, name, namespace, func() []netip.Addr {
		return []netip.Addr{netip.MustParseAddr("192.0.2.10")}
	}, ticks)

	if gets != 2 {
		t.Fatalf("got %d Service GETs across two ticks, want 2", gets)
	}
	actions := client.Actions()
	if len(actions) != 3 || actions[0].GetVerb() != "get" || actions[1].GetVerb() != "get" || !actions[2].Matches("patch", "services") {
		t.Fatalf("got actions %v, want GET, GET, PATCH", actions)
	}
	if patch := actions[2].(clienttesting.PatchAction); patch.GetPatchType() != types.ApplyPatchType || patch.GetName() != name || patch.GetNamespace() != namespace {
		t.Fatalf("unexpected Service patch: %v", patch)
	}
	obj, err := client.Tracker().Get(corev1.SchemeGroupVersion.WithResource("services"), namespace, name)
	if err != nil {
		t.Fatal(err)
	}
	if got := obj.(*corev1.Service).Spec.ExternalIPs; !slices.Equal(got, []string{"192.0.2.10"}) {
		t.Errorf("stored spec.externalIPs = %v, want [192.0.2.10]", got)
	}
}
