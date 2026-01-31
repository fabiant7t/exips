package registry

import (
	"net/netip"
	"reflect"
	"testing"

	"github.com/fabiant7t/exips/internal/node"
)

func TestParseExternalIPs(t *testing.T) {
	reg := New()
	// cp-1 is not ready (and schedulable for testing, which is not the case in real life)
	reg.add(node.NewDummyNode("cp-1", false, true, true, ptr(netip.MustParseAddr("1.2.3.4"))))
	// cp-2 is ready and schedulable, but has no public ip
	reg.add(node.NewDummyNode("cp-2", true, true, true, nil))
	// cp-3 is ready, schedulable and has a public IP
	reg.add(node.NewDummyNode("cp-3", true, true, true, ptr(netip.MustParseAddr("3.2.1.0"))))
	// cp-4 is ready, unschedulable and has a public IP
	reg.add(node.NewDummyNode("cp-4", true, false, true, ptr(netip.MustParseAddr("4.3.2.1"))))
	// cp-5 is ready and has a public IP, but has taint to prevent scheduling workloads
	reg.add(node.NewDummyNode("cp-5", true, true, false, ptr(netip.MustParseAddr("2.3.4.5"))))
	// w-1 is ready and has a public IP, but is unschedulable (cordoned)
	reg.add(node.NewDummyNode("w-1", true, false, true, ptr(netip.MustParseAddr("3.4.5.6"))))
	// w-2 is ready, schedulable and has a public IP
	reg.add(node.NewDummyNode("w-2", true, true, true, ptr(netip.MustParseAddr("5.6.7.8"))))

	got := reg.ParseExternalIPs()
	want := []netip.Addr{
		netip.MustParseAddr("3.2.1.0"),
		netip.MustParseAddr("5.6.7.8"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %+v, want %+v", got, want)
	}
}

func ptr(a netip.Addr) *netip.Addr {
	return &a
}
