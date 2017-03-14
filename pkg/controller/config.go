package controller

import (
	"os"

	"k8s.io/kubernetes/pkg/client/restclient"
)

type Config struct {
	Master            string
	KubeConfig        string
	RESTConfig        *restclient.Config
	HeketiUrl         string
	GlusterFSImage    string
	ClusterDomain     string
	HeketiServiceName string

	K8sPodName      string
	K8sPodNamespace string

	heketiServiceIP string
}

const (
	PodNameEnvVar      = "K8S_POD_NAME"
	PodNamespaceEnvVar = "K8S_POD_NAMESPACE"
)

func NewConfig() *Config {
	return &Config{
		HeketiUrl:         "http://127.0.0.1:8080",
		GlusterFSImage:    "appscode/glusterd:3.10",
		ClusterDomain:     "cluster.local",
		HeketiServiceName: "appscode-gluster",
		K8sPodName:        os.Getenv(PodNameEnvVar),
		K8sPodNamespace:   os.Getenv(PodNamespaceEnvVar),
	}
}
