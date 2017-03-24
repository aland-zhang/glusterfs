package api

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"time"
)

type Glusterfs struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object's metadata.
	// More info: http://releases.k8s.io/release-1.2/docs/devel/api-conventions.md#metadata
	api.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlusterFSSpec   `json:"spec,omitempty"`
	Status GlusterFSStatus `json:"status,omitempty"`
}

type GlusterfsList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: http://releases.k8s.io/release-1.2/docs/devel/api-conventions.md#metadata
	unversioned.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Glusterfs.
	Items []Glusterfs `json:"items,omitempty"`
}

type GlusterFSSpec struct {
	// TODO(@sadlil) Fix this when replica size is variable in kubernetes
	// https://github.com/kubernetes/kubernetes/blob/release-1.5/pkg/volume/glusterfs/glusterfs.go#L503
	// This always tells heketi to create 3 replica, but heketi do not allow
	// creating two bricks in same node. So Replicas must >= 3.
	Replicas int32                `json:"replicas,omitempty"`
	Zone     int                  `json:"zone,omitempty"`
	Storage  GlusterfsStorageSpec `json:"storage,omitempty"`
}

type GlusterfsStorageSpec struct {
	// Name of the StorageClass to use when requesting storage provisioning.
	StorageClass string `json:"storageClass"`

	VolumeClaimTemplates api.PersistentVolumeClaimSpec `json:"volumeClaimTemplates,omitempty"`

	ClusterOptions ClusterOption `json:"clusterOptions"`
}

type ClusterOption struct {
	ClusterDeleteOptions *DeleteOptions `json:"deleteOptions"`
}

type DeleteOptions struct {
	DeleteVolumes bool `json:"deleteVolumes,omitempty"`
}

type GlusterFSStatus struct {
	CreatedAt          time.Time         `json:"createdAt,omitempty"`
	StorageClassName   string            `json:"storageClassname,omitempty"`
	StatefulSetName    string            `json:"statefulSetName,omitempty"`
	StatefulSetService string            `json:"statefulSetService,omitempty"`
	HeketiClusterID    string            `json:"heketiClusterId,omitempty"`
	PodMappings        map[string]string `json:"podMappings,omitempty"`
}
