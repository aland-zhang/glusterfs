package cmd

import (
	"github.com/appscode/glusterfs/pkg/controller"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/util/runtime"
)

func NewCmdRun() *cobra.Command {
	config := controller.NewConfig()
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run Postgres in Kubernetes",
		Run: func(cmd *cobra.Command, args []string) {
			rest, err := clientcmd.BuildConfigFromFlags(config.Master, config.KubeConfig)
			if err != nil {
				log.Fatal("Failed to load KubeConfig", err)
			}
			defer runtime.HandleCrash()
			config.RESTConfig = rest

			cntrl := controller.NewController(config)
			cntrl.Run()
		},
	}
	runCmd.Flags().StringVar(&config.Master, "master", config.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	runCmd.Flags().StringVar(&config.KubeConfig, "kubeconfig", config.KubeConfig, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	runCmd.Flags().StringVar(&config.HeketiUrl, "heketi-url", config.HeketiUrl, "Heketi Server URL")
	runCmd.Flags().StringVar(&config.GlusterFSImage, "glusterfs-image", config.GlusterFSImage, "GlusterFS Image name to run")
	runCmd.Flags().StringVar(&config.ClusterDomain, "cluster-domain", config.ClusterDomain, "Cluster domain. Default cluster.local")
	runCmd.Flags().StringVar(&config.HeketiServiceName, "service-name", config.HeketiServiceName, "Heketi Service Name")

	return runCmd
}
