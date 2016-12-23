package controller

import (
	"time"

	"github.com/appscode/k8s-addons/pkg/stash"
	"github.com/appscode/log"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

func (g *GlusterFSController) Watch() {
	log.Info("started watching for pods resource")
	lw := &cache.ListWatch{
		ListFunc:  g.listFunc(g.KubeClient),
		WatchFunc: g.watchFunc(g.KubeClient),
	}
	// kCachePopulated(k, events.Pod, &api.Pod{}, nil)
	_, controller := stash.NewInformerPopulated(lw,
		&kapi.Pod{},
		time.Second*20,
		eventHandlerFuncs(g),
	)
	go controller.Run(wait.NeverStop)
}

func (g *GlusterFSController) listFunc(c clientset.Interface) func(kapi.ListOptions) (runtime.Object, error) {
	return func(opts kapi.ListOptions) (runtime.Object, error) {
		return c.Core().Pods(g.PodNamespace).List(kapi.ListOptions{
			LabelSelector: g.selector(),
		})
	}
}

func (g *GlusterFSController) watchFunc(c clientset.Interface) func(options kapi.ListOptions) (watch.Interface, error) {
	return func(options kapi.ListOptions) (watch.Interface, error) {
		return c.Core().Pods(g.PodNamespace).Watch(kapi.ListOptions{
			LabelSelector: g.selector(),
		})
	}
}

func eventHandlerFuncs(g *GlusterFSController) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Infoln("got one added event")
			g.reconfig()
		},
		DeleteFunc: func(obj interface{}) {
			log.Infoln("got one deleted event")
			g.reconfig()
		},
	}
}
