package registry

import (
	"slices"
	"testing"

	"github.com/fabiant7t/exips/internal/node"
)

func TestAdd(t *testing.T) {
	reg := New()
	if got, want := len(reg.List()), 0; got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	n := node.NewDummyNode("cp-1", true, true, true, nil)
	reg.add(n)
	if got, want := len(reg.List()), 1; got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}

func TestDelete(t *testing.T) {
	reg := New()
	n := node.NewDummyNode("cp-1", true, true, true, nil)
	reg.add(n)
	if got, want := len(reg.List()), 1; got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	reg.delete(n)
	if got, want := len(reg.List()), 0; got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}

func TestGetMissing(t *testing.T) {
	reg := New()
	_, ok := reg.Get("missing")
	if got, want := ok, false; got != want {
		t.Errorf("Got %t, want %t", got, want)
	}
}

func TestGet(t *testing.T) {
	reg := New()
	reg.add(node.NewDummyNode("cp-1", true, true, true, nil))
	n, ok := reg.Get("cp-1")
	if got, want := ok, true; got != want {
		t.Errorf("Got %t, want %t", got, want)
	}
	if got, want := n.Name(), "cp-1"; got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestUpdate(t *testing.T) {
	reg := New()
	n := node.NewDummyNode("cp-1", true, true, true, nil)
	reg.add(node.NewDummyNode("cp-1", true, true, true, nil))
	nNotReady := node.NewDummyNode("cp-1", false, false, true, nil)
	reg.update(n, nNotReady)
	nFromReg, ok := reg.Get("cp-1")
	if got, want := ok, true; got != want {
		t.Errorf("Got %t, want %t", got, want)
	}
	if got, want := nFromReg.IsReady(), false; got != want {
		t.Errorf("Got %t, want %t", got, want)
	}
}

func TestList(t *testing.T) {
	reg := New()
	n1 := node.NewDummyNode("1", true, true, true, nil)
	n2 := node.NewDummyNode("2", true, true, true, nil)
	n3 := node.NewDummyNode("3", true, true, true, nil)
	reg.add(n1)
	reg.add(n2)
	reg.add(n3)
	for range 9 { // validate order is deterministic and not just lucky
		if got, want := reg.List(), []node.Node{n1, n2, n3}; !slices.Equal(got, want) {
			t.Errorf("Got %v, want %v", got, want)
		}
	}
}
