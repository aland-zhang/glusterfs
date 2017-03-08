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
	Items []Glusterfs `json:"items"`
}

type GlusterFSSpec struct{}

type GlusterFSStatus struct{}
