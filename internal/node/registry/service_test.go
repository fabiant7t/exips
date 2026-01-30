package registry

import (
	"net/netip"
	"reflect"
	"testing"

	"github.com/fabiant7t/exips/internal/node"
)

func TestParseExternalIPs(t *testing.T) {
	reg := New()
	// cp-1 is not ready
	reg.add(node.NewDummyNode("cp-1", false, true, ptr(netip.MustParseAddr("1.2.3.4"))))
	// cp-2 is ready and has no public ip
	reg.add(node.NewDummyNode("cp-2", true, true, nil))
	// cp-3 is ready and has a public IP
	reg.add(node.NewDummyNode("cp-3", true, true, ptr(netip.MustParseAddr("3.2.1.0"))))
	// cp-4 is ready and has a public IP
	reg.add(node.NewDummyNode("cp-4", true, true, ptr(netip.MustParseAddr("4.3.2.1"))))
	// cp-5 is ready and has a public IP, but has taint to prevent scheduling workloads
	reg.add(node.NewDummyNode("cp-5", true, false, ptr(netip.MustParseAddr("2.3.4.5"))))

	got := reg.ParseExternalIPs()
	want := []netip.Addr{
		netip.MustParseAddr("3.2.1.0"),
		netip.MustParseAddr("4.3.2.1"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %+v, want %+v", got, want)
	}
}

func ptr(a netip.Addr) *netip.Addr {
	return &a
}
