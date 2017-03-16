package controller

import (
	"github.com/appscode/log"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

const (
	BetaStorageClassKey = "volume.beta.kubernetes.io/storage-class"
	dynamicEpSvcPrefix  = "glusterfs-dynamic-"
)

type pvcController struct {
	c *Controller
}

func NewPVCController(c *Controller) *pvcController {
	return &pvcController{
		c: c,
	}
}

// Blocks caller.
func (p *pvcController) Run() {
	lw := &cache.ListWatch{
		ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
			return p.c.Client.Core().PersistentVolumeClaims(kapi.NamespaceAll).List(kapi.ListOptions{})
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return p.c.Client.Core().PersistentVolumeClaims(kapi.NamespaceAll).Watch(kapi.ListOptions{})
		},
	}
	_, controller := cache.NewInformer(lw,
		&kapi.PersistentVolumeClaim{},
		p.c.SyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: p.create,
		},
	)
	log.Infoln("Running PVC Watcher")
	controller.Run(wait.NeverStop)
}

func (p *pvcController) create(obj interface{}) {
	if pvc, ok := obj.(*kapi.PersistentVolumeClaim); ok {
		if pvc.Annotations != nil {
			if sc, ok := pvc.Annotations[BetaStorageClassKey]; ok {
				storageClass, err := p.c.Client.Storage().StorageClasses().Get(sc)
				if err != nil {
					if storageClass.Annotations != nil {
						val, ok := storageClass.Annotations["glusterfs.appscode.com/provisioner"]
						if ok && val == "knight" {
							log.Infoln("PVC Created. Need To create Service and Endpoints")
							service := &kapi.Service{
								ObjectMeta: kapi.ObjectMeta{
									Name:      dynamicEpSvcPrefix + pvc.Name,
									Namespace: pvc.Namespace,
									Labels: map[string]string{
										"gluster.kubernetes.io/provisioned-for-pvc": pvc.Name,
									},
								},
								Spec: kapi.ServiceSpec{
									Ports: []kapi.ServicePort{
										{Protocol: "TCP", Port: 1},
									},
									Selector: storageClass.Labels,
								},
							}

							_, err := p.c.Client.Core().Services(pvc.Namespace).Create(service)
							if err != nil && errors.IsAlreadyExists(err) {
								err := p.c.Client.Core().Services(pvc.Namespace).Delete(service.Name, &kapi.DeleteOptions{})
								if err != nil {
									log.Errorln("Failed To delete existing service for pvc")
									return
								}
								p.c.Client.Core().Services(pvc.Namespace).Create(service)
							}
						}
					}
				}
			}
		}
	}
}
