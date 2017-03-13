package controller

import "k8s.io/kubernetes/pkg/client/restclient"

type Config struct {
	Master         string
	KubeConfig     string
	RESTConfig     *restclient.Config
	HeketiUrl      string
	GlusterFSImage string
}

func NewConfig() *Config {
	return &Config{
		HeketiUrl:      "http://127.0.0.1:8080",
		GlusterFSImage: "appscode/glusterd:3.10",
	}
}
