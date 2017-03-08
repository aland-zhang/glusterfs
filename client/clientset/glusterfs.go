package clientset

import (
	glusterapi "github.com/appscode/glusterfs/api"
	"k8s.io/kubernetes/pkg/api"
	rest "k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/watch"
)

type GlusterfsNamespacer interface {
	Glusterfs(namespace string) GlusterfsInterface
}

// ExtendedIngressInterface exposes methods to work on ExtendedIngress resources.
type GlusterfsInterface interface {
	List(opts api.ListOptions) (*glusterapi.GlusterfsList, error)
	Get(name string) (*glusterapi.Glusterfs, error)
	Create(ExtendedIngress *glusterapi.Glusterfs) (*glusterapi.Glusterfs, error)
	Update(ExtendedIngress *glusterapi.Glusterfs) (*glusterapi.Glusterfs, error)
	Delete(name string) error
	Watch(opts api.ListOptions) (watch.Interface, error)
}

// ExtendedIngress implements ExtendedIngressNamespacer interface
type GlusterfsImpl struct {
	r  rest.Interface
	ns string
}

// newGlusterfs returns a ExtendedIngress
func newGlusterfs(c *GlusterfsExtensionsClient, namespace string) *GlusterfsImpl {
	return &GlusterfsImpl{c.restClient, namespace}
}

// List returns a list of ExtendedIngress that match the label and field selectors.
func (c *GlusterfsImpl) List(opts api.ListOptions) (result *glusterapi.GlusterfsList, err error) {
	result = &glusterapi.GlusterfsList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("glusterfses").
		VersionedParams(&opts, ExtendedCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular ExtendedIngress.
func (c *GlusterfsImpl) Get(name string) (result *glusterapi.Glusterfs, err error) {
	result = &glusterapi.Glusterfs{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("glusterfses").
		Name(name).
		Do().
		Into(result)
	return
}

// Create creates a new ExtendedIngress.
func (c *GlusterfsImpl) Create(gfs *glusterapi.Glusterfs) (result *glusterapi.Glusterfs, err error) {
	result = &glusterapi.Glusterfs{}
	err = c.r.Post().
		Namespace(c.ns).
		Resource("glusterfses").
		Body(gfs).
		Do().
		Into(result)
	return
}

// Update updates an existing ExtendedIngress.
func (c *GlusterfsImpl) Update(gfs *glusterapi.Glusterfs) (result *glusterapi.Glusterfs, err error) {
	result = &glusterapi.Glusterfs{}
	err = c.r.Put().
		Namespace(c.ns).
		Resource("glusterfses").
		Name(gfs.Name).
		Body(gfs).
		Do().
		Into(result)
	return
}

// Delete deletes a ExtendedIngress, returns error if one occurs.
func (c *GlusterfsImpl) Delete(name string) (err error) {
	return c.r.Delete().
		Namespace(c.ns).
		Resource("glusterfses").
		Name(name).
		Do().
		Error()
}

// Watch returns a watch.Interface that watches the requested ExtendedIngress.
func (c *GlusterfsImpl) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("glusterfses").
		VersionedParams(&opts, ExtendedCodec).
		Watch()
}
