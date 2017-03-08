package controller

import (
	"fmt"
	"time"

	"github.com/appscode/glusterfs/api"
	gluserclientset "github.com/appscode/glusterfs/client/clientset"
	"github.com/appscode/log"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

type Controller struct {
	// Kubernetes client to apiserver
	Client clientset.Interface
	// ThirdPartyExtension client to apiserver
	ExtClient gluserclientset.GlusterfsExtensionInterface
	// sync time to sync the list.
	SyncPeriod time.Duration

	config *Config
}

func NewController(c *Config) *Controller {
	return &Controller{
		Client:     clientset.NewForConfigOrDie(c.RESTConfig),
		ExtClient:  gluserclientset.NewGlusterfsExtensionsForConfigOrDie(c.RESTConfig),
		SyncPeriod: time.Minute * 2,
		config:     c,
	}
}

// Blocks caller.
func (c *Controller) Run() {
	c.ensureResource()

	lw := &cache.ListWatch{
		ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Glusterfs(kapi.NamespaceAll).List(kapi.ListOptions{})
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Glusterfs(kapi.NamespaceAll).Watch(kapi.ListOptions{})
		},
	}
	_, controller := cache.NewInformer(lw,
		&api.Glusterfs{},
		c.SyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println("Do Things When Added")
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Println("Do Things When Deleted")
			},
			UpdateFunc: func(old, new interface{}) {
				fmt.Println("Do Things When Updated")
			},
		},
	)
	controller.Run(wait.NeverStop)
}

var resourceList = []string{
	"glusterfs",
}

func (c *Controller) ensureResource() {
	for _, resource := range resourceList {
		// This is version dependent
		_, err := c.Client.Extensions().ThirdPartyResources().Get(resource + "." + api.V1Beta1SchemeGroupVersion.Group)
		if k8serrors.IsNotFound(err) {
			tpr := &extensions.ThirdPartyResource{
				TypeMeta: unversioned.TypeMeta{
					APIVersion: "extensions/v1beta1",
					Kind:       "ThirdPartyResource",
				},
				ObjectMeta: kapi.ObjectMeta{
					Name: resource + "." + api.V1Beta1SchemeGroupVersion.Group,
				},
				Versions: []extensions.APIVersion{
					{
						Name: api.V1Beta1SchemeGroupVersion.Version,
					},
				},
			}
			_, err := c.Client.Extensions().ThirdPartyResources().Create(tpr)
			if err != nil {
				// This should fail if there is one third party resource data missing.
				log.Fatalln(tpr.Name, "failed to create, causes", err.Error())
			}
			time.Sleep(time.Second * 5)
		}
	}
}
