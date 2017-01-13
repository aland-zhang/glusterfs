package main

import (
	"flag"
	"os"
	"time"

	"github.com/appscode/glusterfs/pkg/controller"
	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/hold"
	"github.com/appscode/k8s-addons/pkg/election"
	"github.com/appscode/log"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/pflag"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

var (
	flags = pflag.NewFlagSet(`glusterc --election=<name>`, pflag.ExitOnError)

	master     = flags.String("master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	kubeConfig = flags.String("kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")

	name      = flags.String("election", "", "The name of the election")
	namespace = flags.String("namespace", "appscode", "The Kubernetes namespace for this election")
	ttl       = flags.Duration("ttl", 10*time.Second, "The TTL for this election")

	clusterDoman = flags.String("cluster-domain", "cluster.local", "Domain for this cluster.")
)

var (
	currentPodId = rand.Characters(13)
)

func main() {
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args[1:])
	flags.VisitAll(func(flag *pflag.Flag) {
		log.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
	validateFlags()

	logs.InitLogs()
	defer logs.FlushLogs()

	c, err := clientcmd.BuildConfigFromFlags(*master, *kubeConfig)
	if err != nil {
		log.Fatalf("error connecting to the client: %v", err)
	}

	e, err := election.NewElection(*name, currentPodId, *namespace, *ttl, Task, clientset.NewForConfigOrDie(c))
	if err != nil {
		log.Fatalf("failed to create election: %v", err)
	}
	go election.RunElection(e)
	hold.Hold()
}

func validateFlags() {
	if len(currentPodId) == 0 {
		log.Fatal("Pod id generates empty")
	}
	if len(*name) == 0 {
		log.Fatal("--election cannot be empty")
	}
}

func Task(leaderId string) {
	if leaderId == "" {
		log.Infoln("invalid leader id, skipping...")
		return
	}

	log.Infoln("leader found with id", leaderId)
	if leaderId == currentPodId {
		log.Infoln("OK, I became the leader")
		t := controller.NewGlusterFSController(*clusterDoman)
		t.Run(leaderId)
		log.Infoln("task completed")
	}
}
