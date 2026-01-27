package node_test

import (
	"net/netip"
	"testing"

	"github.com/fabiant7t/exips/internal/node"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeName(t *testing.T) {
	name := "cp-1"
	n := node.New(&corev1.Node{
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
		node node.Node
		want bool
	}{
		{
			name: "kubelet is ready",
			node: node.New(&corev1.Node{
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
			node: node.New(&corev1.Node{
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
			node: node.New(&corev1.Node{
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
		node             node.Node
		want             netip.Addr
		shouldRaiseError bool
	}{
		{
			name: "private internal IP",
			node: node.New(&corev1.Node{
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
			node: node.New(&corev1.Node{
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
		node             node.Node
		want             netip.Addr
		shouldRaiseError bool
	}{
		{
			name: "private internal and private external IP",
			node: node.New(&corev1.Node{
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
			node: node.New(&corev1.Node{
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
			node: node.New(&corev1.Node{
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
			node: node.New(&corev1.Node{
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
