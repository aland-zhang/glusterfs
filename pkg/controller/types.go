package controller

import (
	"os"
	"sync"

	"github.com/appscode/glusterfs/pkg/env"
	"github.com/appscode/log"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	rest "k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/selection"
	"k8s.io/kubernetes/pkg/util/sets"
)

type GlusterFSController struct {
	ID         string
	ElectionId string

	PodName           string
	PodNamespace      string
	GlusterFS         string
	Selector          labels.Selector
	Size              int
	count             int
	supportPTRRecords bool
	clusterDomain     string

	Spec *GlusterControllerSpec

	KubeClient clientset.Interface
	KubeConfig *rest.Config

	sync.Once
	sync.Mutex
}

type IDSpecs struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	HostName string `json:"host_name"`
	FQDN     string `json:"fqdn"`
}

type GlusterControllerSpec struct {
	ReplicaCount   int       `json:"replicaCount"`
	ControllerName string    `json:"controllerName"`
	Election       string    `json:"election"`
	Peer           []IDSpecs `json:"peer"`
}

func NewGlusterFSController(clusterDomain string) *GlusterFSController {
	conf, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	c, err := clientset.NewForConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

	return &GlusterFSController{
		GlusterFS:     os.Getenv(env.EnvVarGlusterFSCluster),
		ElectionId:    os.Getenv(env.EnvVarElectionID),
		PodName:       os.Getenv(env.EnvVarPodName),
		PodNamespace:  os.Getenv(env.EnvVarPodNamespace),
		KubeClient:    c,
		KubeConfig:    conf,
		clusterDomain: clusterDomain,
	}
}

func (g *GlusterFSController) selector() labels.Selector {
	if g.Selector == nil {
		l := labels.NewSelector()
		ls, _ := labels.NewRequirement(
			env.GlusterFSClusterNameKey,
			selection.Equals,
			sets.NewString(g.GlusterFS).List())
		l = l.Add(*ls)
		ls, _ = labels.NewRequirement(
			env.GlusterFSElectionIDKey,
			selection.Equals,
			sets.NewString(g.ElectionId).List())
		l = l.Add(*ls)
		g.Selector = l
	}
	return g.Selector
}
