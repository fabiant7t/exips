package registry

import (
	"net/netip"
)

// ParseExternalIPs returns the public IPs of all ready and schedulable nodes,
// excluding control-plane nodes tainted with
// node-role.kubernetes.io/control-plane:NoSchedule.
func (r *Registry) ParseExternalIPs() []netip.Addr {
	nodes := r.List() // already ordered
	ips := make([]netip.Addr, 0, len(nodes))
	for _, n := range nodes {
		if !n.IsReady() {
			continue
		}
		if !n.IsSchedulable() {
			continue
		}
		if !n.IsControlPlaneSchedulable() {
			continue
		}
		pubIP, err := n.PublicIP()
		if err == nil {
			ips = append(ips, pubIP)
		}
	}
	return ips
}
