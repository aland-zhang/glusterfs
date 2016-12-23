package controller

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
)

func (g *GlusterFSController) execPeerConnect(peers []api.Pod) {
	cmds := g.glusterPeerConnectCmd(peers)
	exec := &ExecOptions{
		Namespace:     g.PodNamespace,
		PodName:       g.PodName,
		ContainerName: "glusterfs",
		Command:       cmds,

		Executor: &RemoteBashExecutor{},

		Client: g.KubeClient,
		Config: g.KubeConfig,
	}
	// ignoring the error in case of current execs.
	exec.Run(5)
}

func (g *GlusterFSController) glusterPeerConnectCmd(pods []api.Pod) []string {
	cmds := make([]string, 0)
	for _, pod := range pods {
		if pod.Name == g.PodName {
			continue
		}
		cmds = append(cmds, "/usr/sbin/gluster peer probe "+pod.Status.PodIP)
	}
	return cmds
}

func (g *GlusterFSController) execVolumeCreateCmd(peers []api.Pod) {
	cmds := g.glusterVolumeCreateCmd(peers)
	exec := &ExecOptions{
		Namespace:     g.PodNamespace,
		PodName:       g.PodName,
		ContainerName: "glusterfs",
		Command:       cmds,

		Executor: &RemoteBashExecutor{},

		Client: g.KubeClient,
		Config: g.KubeConfig,
	}
	exec.Run(5)
}

func (g *GlusterFSController) glusterVolumeCreateCmd(pods []api.Pod) []string {
	bricks := ""
	for _, p := range pods {
		bricks = bricks + p.Status.PodIP + ":/storage/volumes/" + g.GlusterFS + " "
	}

	createCmd := ""
	if len(pods) > 1 {
		createCmd = fmt.Sprintf(`/usr/sbin/gluster volume create %s replica %d transport tcp %s force`, g.GlusterFS, len(pods), bricks)
	} else {
		createCmd = fmt.Sprintf(`/usr/sbin/gluster volume create %s transport tcp %s force`, g.GlusterFS, bricks)
	}
	cmds := []string{
		createCmd,
		fmt.Sprintf(`/usr/sbin/gluster volume start %s`, g.GlusterFS),
	}
	return cmds
}

func (g *GlusterFSController) execBrickReConfigure(added []api.Pod, removed []api.Pod) {
	cmds := g.glusterBrickReConfigureCmd(added, removed)
	exec := &ExecOptions{
		Namespace:     g.PodNamespace,
		PodName:       g.PodName,
		ContainerName: "glusterfs",
		Command:       cmds,

		Executor: &RemoteBashExecutor{},

		Client: g.KubeClient,
		Config: g.KubeConfig,
	}
	exec.Run(5)

	cmds = g.glusterBrickHealCmd()
	execHeal := &ExecOptions{
		Namespace:     g.PodNamespace,
		PodName:       g.PodName,
		ContainerName: "glusterfs",
		Command:       cmds,

		Executor: &RemoteBashExecutor{},

		Client: g.KubeClient,
		Config: g.KubeConfig,
	}
	execHeal.Run(5)
}

func (g *GlusterFSController) glusterBrickReConfigureCmd(added []api.Pod, removed []api.Pod) []string {
	cmds := make([]string, 0)
	cmds = append(cmds, g.glusterBrickAddCmd(added)...)
	cmds = append(cmds, g.glusterBrickRemoveCmd(removed)...)
	cmds = append(cmds, g.glusterPeerDetachCmd(removed)...)
	return cmds
}

func (g *GlusterFSController) glusterBrickRemoveCmd(removed []api.Pod) []string {
	cmds := make([]string, 0)
	for i := range removed {
		g.count--
		cmds = append(cmds, fmt.Sprintf(`yes y | /usr/sbin/gluster volume remove-brick %s replica %v %s force`,
			g.GlusterFS,
			g.count,
			removed[i].Status.PodIP+":/storage/volumes/"+g.GlusterFS,
		))
	}
	return cmds
}

func (g *GlusterFSController) glusterBrickAddCmd(added []api.Pod) []string {
	cmds := make([]string, 0)
	for i := range added {
		g.count++
		cmds = append(cmds, fmt.Sprintf(`/usr/sbin/gluster volume add-brick %s replica %v %s force`,
			g.GlusterFS,
			g.count,
			added[i].Status.PodIP+":/storage/volumes/"+g.GlusterFS,
		))
	}
	return cmds
}

func (g *GlusterFSController) glusterBrickHealCmd() []string {
	cmds := []string{
		fmt.Sprintf("/usr/sbin/gluster volume heal %s full", g.GlusterFS),
	}
	return cmds
}

func (g *GlusterFSController) glusterPeerDetachCmd(removed []api.Pod) []string {
	cmds := make([]string, 0)
	for i := range removed {
		cmds = append(cmds, fmt.Sprintf(`/usr/sbin/gluster peer detach %s`,
			removed[i].Status.PodIP,
		))
	}
	return cmds
}
