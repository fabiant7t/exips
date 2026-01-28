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

type registry struct {
	sync.RWMutex
	repo map[string]node.Node
}

// New creates and returns a node registry.
// Call the `SyncForever` method in a goroutine to start syncing cluster state.
func New() *registry {
	return &registry{
		repo: make(map[string]node.Node),
	}
}

func (r *registry) add(n node.Node) {
	r.Lock()
	defer r.Unlock()
	r.repo[n.Name()] = n
}

func (r *registry) delete(n node.Node) {
	r.Lock()
	defer r.Unlock()
	delete(r.repo, n.Name())
}

// update assumes name is a stable identifier and never changes.
func (r *registry) update(_ node.Node, new node.Node) {
	r.Lock()
	defer r.Unlock()
	r.repo[new.Name()] = new
}

func (r *registry) Get(name string) (node.Node, bool) {
	r.RLock()
	defer r.RUnlock()
	n, ok := r.repo[name]
	return n, ok
}

// List returns all nodes, ordererd by node name ascending.
func (r *registry) List() []node.Node {
	r.RLock()
	defer r.RUnlock()

	keys := make([]string, 0, len(r.repo))
	for k := range r.repo {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	nodes := make([]node.Node, 0, len(r.repo))
	for _, k := range keys {
		nodes = append(nodes, r.repo[k])
	}
	return nodes
}

// SyncForever
func (r *registry) SyncForever(ctx context.Context, client kubernetes.Interface, defaultResync time.Duration) error {
	factory := informers.NewSharedInformerFactory(client, defaultResync)
	nodeInformer := factory.Core().V1().Nodes().Informer()
	_, err := nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			r.add(node.New(obj.(*corev1.Node)))
		},
		DeleteFunc: func(obj any) {
			r.delete(node.New(obj.(*corev1.Node)))
		},
		UpdateFunc: func(oldObj, newObj any) {
			r.update(node.New(oldObj.(*corev1.Node)), node.New(newObj.(*corev1.Node)))
		},
	})
	if err != nil {
		return err
	}
	nodeInformer.RunWithContext(ctx)
	return nil
}
