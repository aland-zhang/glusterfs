package controller

import (
	"encoding/json"
	"net"
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
		g.resolvesPTRRecords()
		if ok, spec := firstController(endpoint.Annotations); !ok {
			g.count = spec.ReplicaCount
			added, removed := g.reConfig(peers, spec)
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

func (g *GlusterFSController) resolvesPTRRecords() {
	pod, err := g.KubeClient.Core().Pods(g.PodNamespace).Get(g.PodName)
	if err == nil {
		fqdn := g.podFQDN(*pod)
		answer, _ := net.LookupAddr(pod.Status.PodIP)
		if len(answer) > 0 {
			for _, domain := range answer {
				if domain == fqdn {
					g.supportPTRRecords = true
					return
				}
			}
		}
	}
}

func (g *GlusterFSController) loadPeers() []kapi.Pod {
	for {
		pods, err := g.KubeClient.Core().Pods(g.PodNamespace).List(kapi.ListOptions{
			LabelSelector: g.selector(),
		})

		if err != nil {
			log.Infoln("pod list encountered with error ", err, " not running, waiting to start")
			time.Sleep(time.Second * 20)
			continue
		}

		if len(pods.Items) <= 0 {
			log.Infoln("pod list has zero items, that is not possible, should wait for all pods to be created")
			time.Sleep(time.Second * 20)
			continue
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != kapi.PodRunning || pod.Status.PodIP == "" {
				// This pod do not contain any podIP or not running yet. So this can't add to
				// gluster peer. Wait for pod to get an IP
				// this could wait infinite time if any pod from this cluster goes to NotRunning States.
				log.Infoln("pod with name ", pod.Name, " not running, waiting to start")
				time.Sleep(time.Second * 20)
				continue
			}
		}
		log.Infoln("got all pods with its associated ip, totaling", len(pods.Items))
		return pods.Items
	}
}

func (g *GlusterFSController) updateSpecs(peers []kapi.Pod) error {
	specs := &GlusterControllerSpec{
		ReplicaCount:   len(peers),
		ControllerName: g.ID,
		Election:       g.ElectionId,
		Peer:           make([]IDSpecs, 0),
	}

	for _, peer := range peers {
		id := IDSpecs{
			Name:     peer.Name + "/" + peer.Namespace,
			HostName: peer.Spec.Hostname,
			IP:       peer.Status.PodIP,
			FQDN:     g.podAddress(peer),
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

func (g *GlusterFSController) reConfig(p []kapi.Pod, s *GlusterControllerSpec) ([]kapi.Pod, []kapi.Pod) {
	addedPeers := make([]kapi.Pod, 0)
	removedPeers := make([]kapi.Pod, 0)

	// Finding all the removed peers that needs to be removed from the peer list
	for _, oldPeers := range s.Peer {
		found := false
		for _, peer := range p {
			if g.podAddress(peer) == oldPeers.FQDN {
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
				Spec: kapi.PodSpec{
					Hostname: oldPeers.HostName,
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
			if peer.FQDN == g.podAddress(newPeer) {
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

func (g *GlusterFSController) podFQDN(pod kapi.Pod) string {
	return pod.Name + "." + pod.Spec.Hostname + "." + pod.Namespace + ".svc." + g.clusterDomain
}
