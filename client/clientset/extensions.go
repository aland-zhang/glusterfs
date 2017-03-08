package clientset

import (
	"fmt"

	schema "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	rest "k8s.io/kubernetes/pkg/client/restclient"
)

const (
	defaultAPIPath = "/apis"
)

type GlusterfsExtensionInterface interface {
	RESTClient() rest.Interface
	GlusterfsNamespacer
}

// GlusterfsExtensionsClient is used to interact with experimental Kubernetes features.
// Features of Extensions group are not supported and may be changed or removed in
// incompatible ways at any time.
type GlusterfsExtensionsClient struct {
	restClient rest.Interface
}

func (a *GlusterfsExtensionsClient) Glusterfs(namespace string) GlusterfsInterface {
	return newGlusterfs(a, namespace)
}

// NewAppsCodeExtensions creates a new GlusterfsExtensionsClient for the given config. This client
// provides access to experimental Kubernetes features.
// Features of Extensions group are not supported and may be changed or removed in
// incompatible ways at any time.
func NewGlusterfsExtensionsForConfig(c *rest.Config) (*GlusterfsExtensionsClient, error) {
	config := *c
	if err := setExtensionsDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &GlusterfsExtensionsClient{client}, nil
}

func NewGlusterfsExtensionsForConfigOrDie(c *rest.Config) *GlusterfsExtensionsClient {
	client, err := NewGlusterfsExtensionsForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

func NewGlusterfsExtensions(c rest.Interface) *GlusterfsExtensionsClient {
	return &GlusterfsExtensionsClient{c}
}

func setExtensionsDefaults(config *rest.Config) error {
	gv, err := schema.ParseGroupVersion("appscode.com/v1beta1")
	if err != nil {
		return err
	}
	// if appscode.com/v1beta1 is not enabled, return an error
	if !registered.IsEnabledVersion(gv) {
		return fmt.Errorf("appscode.com/v1beta1 is not enabled")
	}
	config.APIPath = defaultAPIPath
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	if config.GroupVersion == nil || config.GroupVersion.Group != "appscode.com" {
		g, err := registered.Group("appscode.com")
		if err != nil {
			return err
		}
		copyGroupVersion := g.GroupVersion
		config.GroupVersion = &copyGroupVersion
	}

	config.NegotiatedSerializer = DirectCodecFactory{extendedCodec: ExtendedCodec}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *GlusterfsExtensionsClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
