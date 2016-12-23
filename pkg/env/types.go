package env

const (
	EnvVarGlusterFSCluster = "GLUSTERFS_CLUSTER"
	EnvVarElectionID       = "ELECTION_ID"
	EnvVarPodName          = "KUBE_POD_NAME"
	EnvVarPodNamespace     = "KUBE_POD_NAMESPACE"

	GlusterFSClusterNameKey = "glusterfs.beta.appscode.com/name"
	GlusterFSElectionIDKey  = "glusterfs.beta.appscode.com/electionID"
)
