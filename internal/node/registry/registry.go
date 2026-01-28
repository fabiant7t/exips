package registry

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/fabiant7t/exips/internal/node"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Registry struct {
	mu   sync.RWMutex
	repo map[string]node.Node
}

// New creates and returns a node registry.
// Call the `Run` method in a goroutine to start syncing cluster state.
func New() *Registry {
	r := &Registry{
		repo: make(map[string]node.Node),
	}
	return r
}

func (r *Registry) add(n node.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.repo[n.Name()] = n
}

func (r *Registry) delete(n node.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.repo, n.Name())
}

// update assumes name is a stable identifier and never changes.
func (r *Registry) update(_, new node.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.repo[new.Name()] = new
}

// Get looks up a node by name and returns the node and a boolean indicating
// whether or not there was a match.
func (r *Registry) Get(name string) (node.Node, bool) {
	r.mu.RLock()
	n, ok := r.repo[name]
	r.mu.RUnlock()
	return n, ok
}

// List returns all nodes, ordered by node name ascending.
func (r *Registry) List() []node.Node {
	r.mu.RLock()

	keys := make([]string, 0, len(r.repo))
	for k := range r.repo {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	nodes := make([]node.Node, 0, len(r.repo))
	for _, k := range keys {
		nodes = append(nodes, r.repo[k])
	}

	r.mu.RUnlock()

	return nodes
}

// Run
func (r *Registry) Run(ctx context.Context, client kubernetes.Interface, defaultResync time.Duration) error {
	factory := informers.NewSharedInformerFactory(client, defaultResync)
	nodeInformer := factory.Core().V1().Nodes().Informer()

	_, err := nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			n, ok := obj.(*corev1.Node)
			if !ok || n == nil {
				return
			}
			r.add(node.New(n))
		},
		DeleteFunc: func(obj any) {
			var n *corev1.Node
			switch t := obj.(type) {
			case *corev1.Node:
				n = t
			case cache.DeletedFinalStateUnknown:
				if nCandidate, ok := t.Obj.(*corev1.Node); ok {
					n = nCandidate
				} else { // tombstone
					return
				}
			default:
				return
			}
			if n == nil {
				return
			}
			r.delete(node.New(n))
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldN, ok := oldObj.(*corev1.Node)
			if !ok || oldN == nil {
				return
			}
			newN, ok := newObj.(*corev1.Node)
			if !ok || newN == nil {
				return
			}
			r.update(node.New(oldN), node.New(newN))
		},
	})
	if err != nil {
		return err
	}

	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), nodeInformer.HasSynced) {
		return ctx.Err()
	}

	<-ctx.Done()
	return ctx.Err()
}
