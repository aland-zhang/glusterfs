package fake

import (
	"github.com/appscode/glusterfs/client/clientset"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	rest "k8s.io/kubernetes/pkg/client/restclient"
	testing "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
)

type FakeGlusterfsExtensionClient struct {
	*testing.Fake
}

var _ clientset.GlusterfsExtensionInterface = &FakeGlusterfsExtensionClient{}

func NewFakeGlusterfsExtensionClient(objects ...runtime.Object) *FakeGlusterfsExtensionClient {
	o := testing.NewObjectTracker(api.Scheme, api.Codecs.UniversalDecoder())
	for _, obj := range objects {
		if obj.GetObjectKind().GroupVersionKind().Group == "appscode.com" {
			if err := o.Add(obj); err != nil {
				panic(err)
			}
		}
	}

	fakePtr := testing.Fake{}
	fakePtr.AddReactor("*", "*", testing.ObjectReaction(o, registered.RESTMapper()))
	fakePtr.AddWatchReactor("*", testing.DefaultWatchReactor(watch.NewFake(), nil))
	return &FakeGlusterfsExtensionClient{&fakePtr}
}

func (a *FakeGlusterfsExtensionClient) Glusterfs(namespace string) clientset.GlusterfsInterface {
	return &FakeGlusterfs{a.Fake, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeGlusterfsExtensionClient) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
