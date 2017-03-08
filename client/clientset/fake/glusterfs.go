package fake

import (
	glusterapi "github.com/appscode/glusterfs/api"
	"k8s.io/kubernetes/pkg/api"
	schema "k8s.io/kubernetes/pkg/api/unversioned"
	testing "k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type FakeGlusterfs struct {
	Fake *testing.Fake
	ns   string
}

var glusterResource = schema.GroupVersionResource{Group: "appscode.com", Version: "v1beta1", Resource: "glusterfses"}

// Get returns the Ingress by name.
func (mock *FakeGlusterfs) Get(name string) (*glusterapi.Glusterfs, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewGetAction(glusterResource, mock.ns, name), &glusterapi.Glusterfs{})

	if obj == nil {
		return nil, err
	}
	return obj.(*glusterapi.Glusterfs), err
}

// List returns the a of Ingresss.
func (mock *FakeGlusterfs) List(opts api.ListOptions) (*glusterapi.GlusterfsList, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewListAction(glusterResource, mock.ns, opts), &glusterapi.Glusterfs{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &glusterapi.GlusterfsList{}
	for _, item := range obj.(*glusterapi.GlusterfsList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Create creates a new Ingress.
func (mock *FakeGlusterfs) Create(svc *glusterapi.Glusterfs) (*glusterapi.Glusterfs, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewCreateAction(glusterResource, mock.ns, svc), &glusterapi.Glusterfs{})

	if obj == nil {
		return nil, err
	}
	return obj.(*glusterapi.Glusterfs), err
}

// Update updates a Ingress.
func (mock *FakeGlusterfs) Update(svc *glusterapi.Glusterfs) (*glusterapi.Glusterfs, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateAction(glusterResource, mock.ns, svc), &glusterapi.Glusterfs{})

	if obj == nil {
		return nil, err
	}
	return obj.(*glusterapi.Glusterfs), err
}

// Delete deletes a Ingress by name.
func (mock *FakeGlusterfs) Delete(name string) error {
	_, err := mock.Fake.
		Invokes(testing.NewDeleteAction(glusterResource, mock.ns, name), &glusterapi.Glusterfs{})

	return err
}

func (mock *FakeGlusterfs) UpdateStatus(srv *glusterapi.Glusterfs) (*glusterapi.Glusterfs, error) {
	obj, err := mock.Fake.
		Invokes(testing.NewUpdateSubresourceAction(glusterResource, "status", mock.ns, srv), &glusterapi.Glusterfs{})

	if obj == nil {
		return nil, err
	}
	return obj.(*glusterapi.Glusterfs), err
}

func (mock *FakeGlusterfs) Watch(opts api.ListOptions) (watch.Interface, error) {
	return mock.Fake.
		InvokesWatch(testing.NewWatchAction(glusterResource, mock.ns, opts))
}
