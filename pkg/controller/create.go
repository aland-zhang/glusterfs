package controller

import (
	"encoding/json"
	"github.com/appscode/glusterfs/api"
	"github.com/appscode/log"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/apps"
	"strconv"
)

const (
	GlusterFSResourcePrefix = "gfs-"
	GlusterFSSelectorKey    = "glusterfs.appscode.com"
)

var (
	PortsList = []int{
		1,                      // this port is needed to connect with glusterfs as volume for others
		111, 24007, 1110, 2049, // mandatory ports
		49152, 49153, 49154, 49155, 49156, 49157, 49158, 49159, 49160, // bricks ports
		38465, 38466, 38467, 38468, 38469, 38470, 38471, 38472, 38473, // nfs server ports
	}
)

func (c *Controller) create(obj interface{}) {
	if gfs, ok := obj.(*api.Glusterfs); ok {
		if c.validate(gfs) {
			// The Service Must Create Before The StatefulSet
			if err := c.createService(gfs); err != nil {
				log.Errorln("Failed to create Service, cause", err)
				return
			}

			if err := c.createStatefulSet(gfs); err != nil {
				log.Errorln("Failed to Create GlusterFS StatefulSets, cause", err)
				return
			}
		} else {
			log.Errorln("GlusterFS Resource is not valid, removing")
			c.ExtClient.Glusterfs(gfs.Namespace).Delete(gfs.Name)
		}
	} else {
		log.Errorln("Failed to assert Object to Glusterfs")
	}
}

func (c *Controller) createService(gfs *api.Glusterfs) error {
	svc := &kapi.Service{
		ObjectMeta: kapi.ObjectMeta{
			Name:        GlusterFSResourcePrefix + gfs.Name,
			Namespace:   gfs.Namespace,
			Labels:      gfs.Labels,
			Annotations: gfs.Annotations,
		},
		Spec: kapi.ServiceSpec{
			Type:      kapi.ServiceTypeClusterIP,
			Selector:  getSelectorLabels(gfs),
			Ports:     servicePorts(),
			ClusterIP: kapi.ClusterIPNone,
		},
	}

	_, err := c.Client.Core().Services(gfs.Namespace).Create(svc)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) createStatefulSet(gfs *api.Glusterfs) error {
	statefulSet := &apps.StatefulSet{
		ObjectMeta: kapi.ObjectMeta{
			Name:        GlusterFSResourcePrefix + gfs.Name,
			Namespace:   gfs.Namespace,
			Labels:      gfs.Labels,
			Annotations: gfs.Annotations,
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    gfs.Spec.Replicas,
			ServiceName: GlusterFSResourcePrefix + gfs.Name,
			Template: kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{
					Labels:      getSelectorLabels(gfs),
					Annotations: getAnnotations(gfs),
				},
				Spec: kapi.PodSpec{
					Containers: []kapi.Container{
						{
							Name:  "gluster",
							Image: c.config.GlusterFSImage,
							SecurityContext: &kapi.SecurityContext{
								Capabilities: &kapi.Capabilities{
									Add: []kapi.Capability{kapi.Capability("SYS_ADMIN")},
								},
							},
							ImagePullPolicy: kapi.PullAlways,
							Ports:           containerPorts(),
							VolumeMounts: []kapi.VolumeMount{
								{
									Name:      GlusterFSResourcePrefix + gfs.Name,
									MountPath: "/storage",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []kapi.PersistentVolumeClaim{
				{
					ObjectMeta: kapi.ObjectMeta{
						Name: GlusterFSResourcePrefix + gfs.Name,
						Annotations: map[string]string{
							"volume.beta.kubernetes.io/storage-class": gfs.Spec.Storage.StorageClass,
						},
					},
					Spec: gfs.Spec.Storage.VolumeClaimTemplates,
				},
			},
		},
	}

	_, err := c.Client.Apps().StatefulSets(gfs.Namespace).Create(statefulSet)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) validate(gfs *api.Glusterfs) bool {
	if gfs == nil {
		log.Errorln("GlusterFS resource can not be nil")
		return false
	}

	if gfs.Spec.Replicas <= 0 {
		log.Errorln("GlusterFS resource replica can not be 0")
		return false
	}

	if len(gfs.Spec.Storage.StorageClass) == 0 {
		log.Errorln("GlusterFS StorageClass not set")
		return false
	}

	if sc, err := c.Client.Storage().StorageClasses().Get(gfs.Spec.Storage.StorageClass); err != nil {
		log.Errorln("Error getting StorageClass, cause", err)
		return false
	} else if sc == nil {
		log.Errorf("StorageClass %s is nil", gfs.Spec.Storage.StorageClass)
		return false
	}
	return true
}

func getSelectorLabels(gfs *api.Glusterfs) map[string]string {
	// Forward GlusterFS Object Labels to Selector Template
	selectors := gfs.Labels
	if selectors == nil {
		selectors = make(map[string]string)
	}

	// Add Additional Selector Labels
	selectors[GlusterFSSelectorKey+"/resource"] = gfs.Name
	return selectors
}

func getAnnotations(gfs *api.Glusterfs) map[string]string {
	// Forward GlusterFS Object Labels to Selector Template
	annotations := gfs.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	affinity := &kapi.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []kapi.WeightedPodAffinityTerm{
			{
				Weight: 5,
				PodAffinityTerm: kapi.PodAffinityTerm{
					LabelSelector: &unversioned.LabelSelector{
						MatchLabels: getSelectorLabels(gfs),
					},
				},
			},
		},
	}
	affinityBytes, err := json.Marshal(affinity)
	if err == nil {
		annotations[kapi.AffinityAnnotationKey] = string(affinityBytes)
	}
	annotations[GlusterFSSelectorKey+"/resource"] = gfs.Name
	return annotations
}

func servicePorts() []kapi.ServicePort {
	return []kapi.ServicePort{
		{
			Name: "port-ep",
			Port: 1,
		},
	}
}

func containerPorts() []kapi.ContainerPort {
	ports := make([]kapi.ContainerPort, 0)
	for i, p := range PortsList {
		port := kapi.ContainerPort{
			Name:          "port-" + strconv.Itoa(i),
			ContainerPort: int32(p),
		}
		ports = append(ports, port)
	}
	return ports
}
