package main

import (
	"flag"
	"os"
	"time"

	_ "github.com/appscode/glusterfs/pkg/controller"
	rand "github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/hold"
	v "github.com/appscode/go/version"
	"github.com/appscode/k8s-addons/pkg/election"
	"github.com/appscode/log"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/pflag"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
	kubectl_util "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var (
	Version         string
	VersionStrategy string
	Os              string
	Arch            string
	CommitHash      string
	GitBranch       string
	GitTag          string
	CommitTimestamp string
	BuildTimestamp  string
	BuildHost       string
	BuildHostOs     string
	BuildHostArch   string
)

func init() {
	v.Version.Version = Version
	v.Version.VersionStrategy = VersionStrategy
	v.Version.Os = Os
	v.Version.Arch = Arch
	v.Version.CommitHash = CommitHash
	v.Version.GitBranch = GitBranch
	v.Version.GitTag = GitTag
	v.Version.CommitTimestamp = CommitTimestamp
	v.Version.BuildTimestamp = BuildTimestamp
	v.Version.BuildHost = BuildHost
	v.Version.BuildHostOs = BuildHostOs
	v.Version.BuildHostArch = BuildHostArch
}

var (
	flags = pflag.NewFlagSet(`glusterc --election=<name>`, pflag.ExitOnError)

	name      = flags.String("election", "", "The name of the election")
	namespace = flags.String("namespace", "appscode", "The Kubernetes namespace for this election")
	ttl       = flags.Duration("ttl", 10*time.Second, "The TTL for this election")
	inCluster = flags.Bool("incluster", true, "Should this request use cluster credentials")
	task      = flags.String("task", "", "Leader tasks to run")
)

var (
	currentPodId = rand.Characters(13)
)

func makeClient() (client.Interface, error) {
	var cfg *restclient.Config
	var err error

	if *inCluster {
		if cfg, err = restclient.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		clientConfig := kubectl_util.DefaultClientConfig(flags)
		if cfg, err = clientConfig.ClientConfig(); err != nil {
			return nil, err
		}
	}
	return client.NewForConfig(cfg)
}

func validateFlags() {
	if len(currentPodId) == 0 {
		log.Fatal("Pod id generates empty")
	}
	if len(*name) == 0 {
		log.Fatal("--election cannot be empty")
	}
}

func main() {
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)
	validateFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	kubeClient, err := makeClient()
	if err != nil {
		log.Fatalf("error connecting to the client: %v", err)
	}

	e, err := election.NewElection(*name, currentPodId, *namespace, *ttl, Task, kubeClient)
	if err != nil {
		log.Fatalf("failed to create election: %v", err)
	}
	go election.RunElection(e)
	hold.Hold()
}

func Task(leaderId string) {
	if leaderId == "" {
		log.Infoln("invalid leader id, skipping...")
		return
	}

	log.Infoln("leader found with id", leaderId)
	if leaderId == currentPodId {
		log.Infoln("OK, I became the leader, task mood is", *task)
		t, err := election.GetTask(*task)
		if err != nil {
			log.Fatal(err)
		}
		t.Run(leaderId)
		log.Infoln("task completed")
	}
}
