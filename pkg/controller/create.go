package controller

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/appscode/glusterfs/api"
	"github.com/appscode/log"
	heketiapi "github.com/heketi/heketi/pkg/glusterfs/api"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/pkg/labels"
)

const (
	GlusterFSResourcePrefix = "glusterfs-"
	GlusterDomain           = "gluster"
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
	context := &options{}
	if gfs, ok := obj.(*api.Glusterfs); ok {
		if c.validate(context, gfs) {
			// The Service Must Create Before The StatefulSet
			if err := c.ensureNamespaceService(context, gfs); err != nil {
				log.Errorln("Failed to create Service, cause", err)
				return
			}

			if err := c.createStatefulSet(context, gfs); err != nil {
				log.Errorln("Failed to Create GlusterFS StatefulSets, cause", err)
				return
			}
			c.waitForPodsToRun(context, gfs)
			if err := c.addNewHeketiCluster(context, gfs); err != nil {
				log.Errorln("Failed to Create GlusterFS StatefulSets, cause", err)
				return
			}

			if err := c.addStorageClass(context, gfs); err != nil {
				log.Errorln("Failed to Create StorageClass, cause", err)
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

func (c *Controller) ensureNamespaceService(ctx *options, gfs *api.Glusterfs) error {
	_, err := c.Client.Core().Services(gfs.Namespace).Get(GlusterDomain)
	if err != nil {
		svc := &kapi.Service{
			ObjectMeta: kapi.ObjectMeta{
				Name:        GlusterDomain,
				Namespace:   gfs.Namespace,
				Labels:      gfs.Labels,
				Annotations: gfs.Annotations,
			},
			Spec: kapi.ServiceSpec{
				Type:      kapi.ServiceTypeClusterIP,
				ClusterIP: kapi.ClusterIPNone,
			},
		}

		_, err := c.Client.Core().Services(gfs.Namespace).Create(svc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) createStatefulSet(ctx *options, gfs *api.Glusterfs) error {
	replicaCount := int32(3)
	if gfs.Spec.Replicas > replicaCount {
		replicaCount = gfs.Spec.Replicas
	}
	statefulSet := &apps.StatefulSet{
		ObjectMeta: kapi.ObjectMeta{
			Name:        GlusterFSResourcePrefix + gfs.Name,
			Namespace:   gfs.Namespace,
			Labels:      gfs.Labels,
			Annotations: gfs.Annotations,
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    replicaCount,
			ServiceName: GlusterDomain,
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
									// SubPath: "gfs",
								},
								//{
								//	Name:      GlusterFSResourcePrefix + gfs.Name,
								//	MountPath: "/dev",
								//	SubPath: "glusterdev",
								//},
								//{
								//	Name:      GlusterFSResourcePrefix + gfs.Name,
								//	MountPath: "/var/lib/misc/glusterfsd",
								//	SubPath: "glusterfmisc",
								//},
								//{
								//	Name:      GlusterFSResourcePrefix + gfs.Name,
								//	MountPath: "/sys/fs/cgroup",
								//	SubPath: "cgroup",
								//},
								//{
								//	Name:      GlusterFSResourcePrefix + gfs.Name,
								//	MountPath: "/etc/glusterfs",
								//	SubPath: "glusteretc",
								//},
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

func (c *Controller) addNewHeketiCluster(ctx *options, gfs *api.Glusterfs) error {
	cluster, err := c.HeketiClient.ClusterCreate()
	if err != nil {
		return err
	}

	// Create a cleanup function in case no
	// nodes or devices are created
	defer func() {
		// Get cluster information
		info, err := c.HeketiClient.ClusterInfo(cluster.Id)
		// Delete empty cluster
		if err == nil && len(info.Nodes) == 0 && len(info.Volumes) == 0 {
			c.HeketiClient.ClusterDelete(cluster.Id)
		}
	}()
	log.Infoln("Cluster Created, Cluster ID", cluster.Id)
	ctx.heketiOptions.ClusterID = cluster.Id
	pods, err := c.Client.Core().Pods(gfs.Namespace).List(kapi.ListOptions{
		LabelSelector: labels.SelectorFromSet(getSelectorLabels(gfs)),
	})

	ctx.heketiOptions.NodeIDMap = make(map[string]string)
	for {
		for _, pod := range pods.Items {
			if _, ok := ctx.heketiOptions.NodeIDMap[pod.Name]; !ok {
				fqdn := strings.Join([]string{
					pod.Name,
					GlusterDomain,
					pod.Namespace,
					"svc",
					c.config.ClusterDomain,
				}, ".")
				log.Infoln("Adding Node with host name", fqdn)
				req := &heketiapi.NodeAddRequest{
					Zone:      gfs.Spec.Zone,
					ClusterId: cluster.Id,
					Hostnames: heketiapi.HostAddresses{
						Manage:  sort.StringSlice([]string{fqdn}),
						Storage: sort.StringSlice([]string{fqdn}),
					},
				}
				if req.Zone <= 0 {
					req.Zone = 1
				}
				node, err := c.HeketiClient.NodeAdd(req)
				if err != nil {
					log.Infoln("Add Node Failed with reason", err)
					continue
				}
				ctx.heketiOptions.NodeIDMap[pod.Name] = node.Id
			}
		}
		if len(ctx.heketiOptions.NodeIDMap) == len(pods.Items) {
			break
		}
		log.Infoln("All Node not added, retring...")
		time.Sleep(time.Second * 20)
	}

	for _, v := range ctx.heketiOptions.NodeIDMap {
		deviceAddReq := &heketiapi.DeviceAddRequest{
			Device: heketiapi.Device{Name: "/storage"},
			NodeId: v,
		}
		err = c.HeketiClient.DeviceAdd(deviceAddReq)
		if err != nil {
			log.Errorln("Device add faild with error", err)
		}
	}
	log.Infoln("All node Added in the cluster")
	return nil
}

func (c *Controller) addStorageClass(ctx *options, gfs *api.Glusterfs) error {
	sc := &storage.StorageClass{
		ObjectMeta: kapi.ObjectMeta{
			Name:   GlusterFSResourcePrefix + gfs.Name,
			Labels: getSelectorLabels(gfs),
			Annotations: map[string]string{
				GlusterFSSelectorKey + "/provisioner": "knight",
			},
		},
		Provisioner: "kubernetes.io/glusterfs",
		Parameters: map[string]string{
			"resturl": fmt.Sprintf("http://%s:8080", c.config.heketiServiceIP),
			// TODO (@sadlil) Fix those when we use 1.6
			// "clusterid": ctx.heketiOptions.ClusterID,
			// "volumetype": fmt.Sprintf("replicate:%v", gfs.Spec.Replicas),
		},
	}

	_, err := c.Client.Storage().StorageClasses().Create(sc)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) waitForPodsToRun(ctx *options, gfs *api.Glusterfs) {
	// TODO (@sadlil): Should we add a max retry limit?
	for {
		pods, err := c.Client.Core().Pods(gfs.Namespace).List(kapi.ListOptions{
			LabelSelector: labels.SelectorFromSet(getSelectorLabels(gfs)),
		})
		if err != nil {
			log.Infoln("Pod list encountered with error ", err, "waiting...")
			time.Sleep(time.Second * 20)
			continue
		}

		if int32(len(pods.Items)) < gfs.Spec.Replicas {
			log.Infoln("Pod count mismatched, waiting...")
			time.Sleep(time.Second * 20)
			continue
		}

		if int32(len(pods.Items)) == gfs.Spec.Replicas {
			for _, pod := range pods.Items {
				if pod.Status.Phase != kapi.PodRunning || len(pod.Status.PodIP) == 0 {
					log.Infoln("Pod", pod.Name, "not running, waiting...")
					time.Sleep(time.Second * 20)
					continue
				}
			}
		}
		break
	}
	time.Sleep(time.Second * 20)
}

func (c *Controller) validate(ctx *options, gfs *api.Glusterfs) bool {
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
	selectors[GlusterFSSelectorKey+"/provisioner"] = "knight"
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
