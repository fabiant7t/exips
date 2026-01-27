package node

import (
	"errors"
	"log/slog"
	"net/netip"

	corev1 "k8s.io/api/core/v1"
)

var ErrNoPublicIP = errors.New("error: no public IP")

// INTERFACE

// Node interface
type Node interface {
	Name() string
	IsReady() bool
	PublicIP() (netip.Addr, error)
}

// CONSTRUCTORS

// New constructs a v1Node instance
func New(n *corev1.Node) *v1Node {
	return &v1Node{node: n}
}

// New constructs a dummyNode instance
func NewDummyNode(name string, isReady bool, publicIP *netip.Addr) *dummyNode {
	return &dummyNode{
		name:     name,
		isReady:  isReady,
		publicIP: publicIP,
	}
}

// IMPLEMENTATIONS

// v1Node is the anti-corruption layer for corev1 Kubernetes Node objects.
type v1Node struct {
	node *corev1.Node
}

// Name of the node
func (n *v1Node) Name() string {
	return n.node.Name
}

// IsReady returns true when NodeReady condition is true (not false and not unknown).
func (n *v1Node) IsReady() bool {
	for _, cond := range n.node.Status.Conditions {
		if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// PublicIP returns the first non private external or internal IP
func (n *v1Node) PublicIP() (netip.Addr, error) {
	pubExtIP, err := n.PublicExternalIP()
	if err == nil {
		return pubExtIP, nil
	}
	pubIntIP, err := n.PublicInternalIP()
	if err == nil {
		return pubIntIP, nil
	}
	return netip.Addr{}, ErrNoPublicIP
}

// PublicInternalIP returns the non private internal IP
func (n *v1Node) PublicInternalIP() (netip.Addr, error) {
	for _, addr := range n.node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			ip, err := netip.ParseAddr(addr.Address)
			if err != nil {
				slog.Debug("ignoring error when parsing internal IP", "node", n.node.Name, "addr", addr.Address)
				continue
			}
			if !ip.IsPrivate() {
				return ip, nil
			}
		}
	}
	return netip.Addr{}, ErrNoPublicIP
}

// PublicExternalIP returns the first non private IP
func (n *v1Node) PublicExternalIP() (netip.Addr, error) {
	for _, addr := range n.node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			ip, err := netip.ParseAddr(addr.Address)
			if err != nil {
				slog.Debug("ignoring error when parsing external IP", "node", n.node.Name, "addr", addr.Address)
				continue
			}
			if !ip.IsPrivate() {
				return ip, nil
			}
		}
	}
	return netip.Addr{}, ErrNoPublicIP
}

// dummyNode satisfies the Node interface and can be used in tests.
type dummyNode struct {
	name     string
	isReady  bool
	publicIP *netip.Addr
}

func (n *dummyNode) Name() string {
	return n.name
}

func (n *dummyNode) IsReady() bool {
	return n.isReady
}

func (n *dummyNode) PublicIP() (netip.Addr, error) {
	if n.publicIP == nil {
		return netip.Addr{}, ErrNoPublicIP
	}
	return *n.publicIP, nil
}
