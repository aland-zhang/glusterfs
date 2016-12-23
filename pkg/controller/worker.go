package controller

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/appscode/errors"
	"github.com/appscode/log"
	kapi "k8s.io/kubernetes/pkg/api"
)

const (
	glusterFSClusterControllerData = "glusterfs.appscode.com/controller"
)

func (g *GlusterFSController) Run(controllerId string) {
	g.ID = controllerId
	// Query the election for is this cluster runs for the first time
	// or there was another leader before who died.
	endpoint, err := g.KubeClient.Core().Endpoints(g.PodNamespace).Get(g.ElectionId)
	if err != nil {
		// this endpoint should create before it comes to this point.
		// if it is not found we should exit this leader
		log.Fatalln(err)
	}

	peers := g.loadPeers()
	g.Size = len(peers)
	if ok, _ := firstController(endpoint.Annotations); ok {
		g.count = len(peers)
		err := g.updateSpecs(peers)
		if err == nil {
			log.Infoln("update success, running peer probe")
			g.execPeerConnect(peers)
			g.execVolumeCreateCmd(peers)
			log.Infoln("completed first leader tasks")
		}
	} else {
		g.Do(func() {
			time.Sleep(time.Second * 30)
		})
		log.Infoln("not leader for first time, running reconfigure")
		g.reconfig()
	}

	g.Watch()
}

func (g *GlusterFSController) reconfig() {
	g.Lock()
	log.Infoln("called reconfig() to autofix")
	defer g.Unlock()

	endpoint, err := g.KubeClient.Core().Endpoints(g.PodNamespace).Get(g.ElectionId)
	if err == nil {
		peers := g.loadPeers()
		if ok, spec := firstController(endpoint.Annotations); !ok {
			g.count = spec.ReplicaCount
			added, removed := reConfig(peers, spec)
			log.Infoln("reconfig returned added", len(added), "removed", len(removed))
			if len(added) > 0 || len(removed) > 0 {
				log.Infoln("someone changed in cluster, needs to update and reconnect")
				g.execPeerConnect(added)
				g.execBrickReConfigure(added, removed)
				g.updateSpecs(peers)
			}
			log.Infoln("reconfiguration completes")
		}
	}
	log.Infoln("reconfig done returnung to caller, releasing lock")
}

func (g *GlusterFSController) loadPeers() []kapi.Pod {
	pods, err := g.KubeClient.Core().Pods(g.PodNamespace).List(kapi.ListOptions{
		LabelSelector: g.selector(),
	})

	if err != nil {
		log.Infoln("pod list encountered with error ", err, " not running, waiting to start")
		time.Sleep(time.Second * 20)
		return g.loadPeers()
	}

	if len(pods.Items) <= 0 {
		log.Infoln("pod list has zero items, that is not possible, should wait for all pods to be created")
		time.Sleep(time.Second * 20)
		return g.loadPeers()
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != kapi.PodRunning || pod.Status.PodIP == "" {
			// This pod do not contain any podIP or not running yet. So this can't add to
			// gluster peer. Wait for pod to get an IP
			// this could wait infinite time if any pod from this cluster goes to NotRunning States.
			log.Infoln("pod with name ", pod.Name, " not running, waiting to start")
			time.Sleep(time.Second * 20)
			return g.loadPeers()
		}
	}
	log.Infoln("got all pods with its associated ip, totaling", len(pods.Items))
	return pods.Items
}

func (g *GlusterFSController) updateSpecs(peers []kapi.Pod) error {
	specs := &ControllerSpecs{
		ReplicaCount:   len(peers),
		ControllerName: g.ID,
		Election:       g.ElectionId,
		Peer:           make([]IDSpecs, 0),
	}

	for _, peer := range peers {
		id := IDSpecs{
			Name: peer.Name + "/" + peer.Namespace,
			IP:   peer.Status.PodIP,
		}
		specs.Peer = append(specs.Peer, id)
	}
	log.Infoln("updating specs with ", specs)
	data, err := json.Marshal(specs)
	if err != nil {
		return errors.New().WithCause(err).Internal()
	}
	return g.updateSpecEndpoint(string(data))
}

func (g *GlusterFSController) updateSpecEndpoint(data string) error {
	ep, err := g.KubeClient.Core().Endpoints(g.PodNamespace).Get(g.ElectionId)
	if err == nil {
		ep.Annotations[glusterFSClusterControllerData] = data
		_, err = g.KubeClient.Core().Endpoints(g.PodNamespace).Update(ep)
		if err != nil {
			return errors.New().WithCause(err).Internal()
		}
		return nil
	}
	return errors.New().WithCause(err).Internal()
}

func reConfig(p []kapi.Pod, s *ControllerSpecs) ([]kapi.Pod, []kapi.Pod) {
	addedPeers := make([]kapi.Pod, 0)
	removedPeers := make([]kapi.Pod, 0)

	// Finding all the removed peers that needs to be removed from the peer list
	for _, oldPeers := range s.Peer {
		found := false
		for _, peer := range p {
			if peer.Status.PodIP == oldPeers.IP {
				found = true
				break
			}
		}
		if !found {
			log.Infoln("found removed peers", oldPeers)
			name, namespace := oldPeers.Name[:strings.LastIndex(oldPeers.Name, "/")], oldPeers.Name[strings.LastIndex(oldPeers.Name, "/")+1:]
			pod := kapi.Pod{
				ObjectMeta: kapi.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Status: kapi.PodStatus{
					PodIP: oldPeers.IP,
				},
			}
			removedPeers = append(removedPeers, pod)
		}
	}

	for _, newPeer := range p {
		found := false
		for _, peer := range s.Peer {
			if peer.IP == newPeer.Status.PodIP {
				found = true
				break
			}
		}
		if !found {
			log.Infoln("found added peers", newPeer.Name, newPeer.Status.PodIP)
			addedPeers = append(addedPeers, newPeer)
		}
	}
	return addedPeers, removedPeers
}
