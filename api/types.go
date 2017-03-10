package api

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type Glusterfs struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object's metadata.
	// More info: http://releases.k8s.io/release-1.2/docs/devel/api-conventions.md#metadata
	api.ObjectMeta `json:"metadata,omitempty"`

	Spec GlusterFSSpec `json:"spec,omitempty"`
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
	Replicas int32                `json:"replicas,omitempty"`
	Zone     int32                `json:"zone,omitempty"`
	Storage  GlusterfsStorageSpec `json:"storage,omitempty"`
}

type GlusterfsStorageSpec struct {
	// Name of the StorageClass to use when requesting storage provisioning.
	StorageClass string `json:"storageClass"`

	VolumeClaimTemplates api.PersistentVolumeClaimSpec `json:"volumeClaimTemplates,omitempty"`
}

type GlusterFSStatus struct {
	StatefulSetName   string `json:"statefulSetName,omitempty"`
	ServiceName       string `json:"serviceName,omitempty"`
	GlusterFSEndpoint string `json:"glusterfsEndpoint"`
}
