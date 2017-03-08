package gluster

import "k8s.io/kubernetes/pkg/client/restclient"

type Config struct {
	Master     string
	KubeConfig string
	RESTConfig *restclient.Config
}
