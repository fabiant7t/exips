package node

import (
	"net/netip"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeName(t *testing.T) {
	name := "cp-1"
	n := New(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	if got, want := n.Name(), name; got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestNodeIsReady(t *testing.T) {
	for _, tc := range []struct {
		name string
		node *v1Node
		want bool
	}{
		{
			name: "kubelet is ready",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:               corev1.NodeReady,
							Status:             corev1.ConditionTrue,
							Reason:             "KubeletReady",
							Message:            "kubelet is posting ready status",
							LastHeartbeatTime:  metav1.Now(),
							LastTransitionTime: metav1.Now(),
						},
					},
				},
			}),
			want: true,
		},
		{
			name: "kubelet is not ready",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:               corev1.NodeReady,
							Status:             corev1.ConditionFalse,
							Reason:             "KubeletNotReady",
							Message:            "kubelet is not ready",
							LastHeartbeatTime:  metav1.Now(),
							LastTransitionTime: metav1.Now(),
						},
					},
				},
			}),
			want: false,
		},
		{
			name: "kubelet stopped posting node status",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:               corev1.NodeReady,
							Status:             corev1.ConditionUnknown,
							Reason:             "NodeStatusUnknown",
							Message:            "Kubelet stopped posting node status.",
							LastHeartbeatTime:  metav1.Now(),
							LastTransitionTime: metav1.Now(),
						},
					},
				},
			}),
			want: false,
		},
	} {
		if got, want := tc.node.IsReady(), tc.want; got != want {
			t.Errorf("%s: Got %t, want %t", tc.name, got, want)
		}
	}
}

func TestNodePublicInternalIP(t *testing.T) {
	for _, tc := range []struct {
		name             string
		node             *v1Node
		want             netip.Addr
		shouldRaiseError bool
	}{
		{
			name: "private internal IP",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Address: "192.168.0.2",
							Type:    corev1.NodeInternalIP,
						},
					},
				},
			}),
			shouldRaiseError: true,
		},
		{
			name: "public internal IP",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Address: "1.2.3.4",
							Type:    corev1.NodeInternalIP,
						},
					},
				},
			}),
			want: netip.MustParseAddr("1.2.3.4"),
		},
	} {
		publicInternalIP, err := tc.node.PublicInternalIP()
		if err != nil && !tc.shouldRaiseError {
			t.Errorf("%s raised error: %s", tc.name, err)
		}
		if got, want := publicInternalIP, tc.want; got != want {
			t.Errorf("%s: Got %v, want %v", tc.name, got, want)
		}
	}
}

func TestNodePublicIP(t *testing.T) {
	for _, tc := range []struct {
		name             string
		node             *v1Node
		want             netip.Addr
		shouldRaiseError bool
	}{
		{
			name: "private internal and private external IP",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Address: "192.168.0.2",
							Type:    corev1.NodeInternalIP,
						},
						{
							Address: "10.10.10.2",
							Type:    corev1.NodeExternalIP,
						},
					},
				},
			}),
			shouldRaiseError: true,
		},
		{
			name: "private internal and public external IP",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Address: "192.168.0.2",
							Type:    corev1.NodeInternalIP,
						},
						{
							Address: "1.2.3.4",
							Type:    corev1.NodeExternalIP,
						},
					},
				},
			}),
			want: netip.MustParseAddr("1.2.3.4"),
		},
		{
			name: "public internal and private external IP",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Address: "1.2.3.4",
							Type:    corev1.NodeInternalIP,
						},
						{
							Address: "192.168.0.2",
							Type:    corev1.NodeExternalIP,
						},
					},
				},
			}),
			want: netip.MustParseAddr("1.2.3.4"),
		},
		{
			name: "public internal and public external IP",
			node: New(&corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Address: "11.22.33.44",
							Type:    corev1.NodeInternalIP,
						},
						{
							Address: "1.2.3.4",
							Type:    corev1.NodeExternalIP,
						},
					},
				},
			}),
			want: netip.MustParseAddr("1.2.3.4"),
		},
	} {
		publicIP, err := tc.node.PublicIP()
		if err != nil && !tc.shouldRaiseError {
			t.Errorf("%s raised error: %s", tc.name, err)
		}
		if got, want := publicIP, tc.want; got != want {
			t.Errorf("%s: Got %v, want %v", tc.name, got, want)
		}
	}
}

func TestImplementsNode(t *testing.T) {
	var _ Node = &v1Node{}
}

func TestDummyNodeImplementsNode(t *testing.T) {
	var _ Node = &dummyNode{}
}

func TestDummyNode(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		isReady                   bool
		isSchedulable             bool
		isControlPlaneSchedulable bool
		publicIP                  *netip.Addr
		shouldRaiseError          bool
	}{
		{name: "cp-1", isReady: true, isSchedulable: true, isControlPlaneSchedulable: true, publicIP: func(a netip.Addr) *netip.Addr { return &a }(netip.MustParseAddr("1.2.3.1"))},
		{name: "cp-2", isReady: false, isSchedulable: true, isControlPlaneSchedulable: true, publicIP: func(a netip.Addr) *netip.Addr { return &a }(netip.MustParseAddr("1.2.3.2"))},
		{name: "cp-3", isReady: true, isSchedulable: true, isControlPlaneSchedulable: true, shouldRaiseError: true},
		{name: "cp-4", isReady: true, isSchedulable: true, isControlPlaneSchedulable: false, publicIP: func(a netip.Addr) *netip.Addr { return &a }(netip.MustParseAddr("1.2.3.4"))},
		{name: "cp-5", isReady: true, isSchedulable: false, isControlPlaneSchedulable: true, publicIP: func(a netip.Addr) *netip.Addr { return &a }(netip.MustParseAddr("2.3.4.5"))},
	} {
		n := NewDummyNode(tc.name, tc.isReady, tc.isSchedulable, tc.isControlPlaneSchedulable, tc.publicIP)
		if got, want := n.Name(), tc.name; got != want {
			t.Errorf("got %s, want %s", got, want)
		}
		if got, want := n.IsReady(), tc.isReady; got != want {
			t.Errorf("got %t, want %t", got, want)
		}
		if got, want := n.IsControlPlaneSchedulable(), tc.isControlPlaneSchedulable; got != want {
			t.Errorf("got %t, want %t", got, want)
		}
		publicIP, err := n.PublicIP()
		if err != nil && !tc.shouldRaiseError {
			t.Error(err)
		} else if err == nil {
			if got, want := publicIP.String(), tc.publicIP.String(); got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		}
	}
}

func TestNodeIsSchedulable(t *testing.T) {
	for _, tc := range []struct {
		name string
		node *v1Node
		want bool
	}{
		{
			name: "node is unschedulable",
			node: New(&corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "node.kubernetes.io/unschedulable",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			}),
			want: false,
		},
		{
			name: "node is schedulable",
			node: New(&corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{},
				},
			}),
			want: true,
		},
	} {
		if got, want := tc.node.IsSchedulable(), tc.want; got != want {
			t.Errorf("%s: Got %t, want %t", tc.name, got, want)
		}
	}
}

func TestNodeIsControlPlaneSchedulable(t *testing.T) {
	for _, tc := range []struct {
		name string
		node *v1Node
		want bool
	}{
		{
			name: "control-plane has taint to prevent scheduling workloads",
			node: New(&corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "node-role.kubernetes.io/control-plane",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			}),
			want: false,
		},
		{
			name: "control-plane allows scheduling workloads",
			node: New(&corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{},
				},
			}),
			want: true,
		},
		{
			name: "worker node has taint to prevent scheduling workloads",
			node: New(&corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "node.kubernetes.io/unschedulable",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			}),
			want: true,
		},
	} {
		if got, want := tc.node.IsControlPlaneSchedulable(), tc.want; got != want {
			t.Errorf("%s: Got %t, want %t", tc.name, got, want)
		}
	}
}
