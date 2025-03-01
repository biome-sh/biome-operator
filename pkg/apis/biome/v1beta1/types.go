// Copyright (c) 2017 Chef Software Inc. and/or applicable contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	BiomeResourcePlural = "biomes"
	BiomeShortName      = "bio"

	// BiomeLabel labels the resources that belong to Biome.
	// Example: 'biome: true'
	BiomeLabel = "biome"
	// BiomeNameLabel contains the user defined Biome Service name.
	// Example: 'biome-name: db'
	BiomeNameLabel = "biome-name"

	TopologyLabel        = "topology"
	BiomeTopologyLabel = "operator.biome.sh/topology"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Biome struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BiomeSpec   `json:"spec"`
	Status            BiomeStatus `json:"status,omitempty"`
	// CustomVersion is a field that works around the lack of support for running
	// multiple versions of a CRD.  It encodes the actual version of the type, so
	// that controllers can decide whether to discard an object if the version
	// doesn't match.
	CustomVersion *string `json:"customVersion,omitempty"`
}

type BiomeSpec struct {
	// V1beta2 are fields for the v1beta2 type.
	// +optional
	V1beta2 *V1beta2 `json:"v1beta2"`
}

// V1beta2 are fields for the v1beta2 type.
type V1beta2 struct {
	// Count is the amount of Services to start in this Biome.
	Count int `json:"count"`
	// Image is the Docker image of the Biome Service.
	Image   string         `json:"image"`
	Service ServiceV1beta2 `json:"service"`
	// Env is a list of environment variables.
	// The EnvVar type is documented at https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.9/#envvar-v1-core.
	// Optional.
	Env []corev1.EnvVar `json:"env,omitempty"`
	// +optional
	PersistentStorage *PersistentStorage `json:"persistentStorage,omitempty"`
}

// PersistentStorage contains the details of the persistent storage that the
// cluster should provision.
type PersistentStorage struct {
	// Size is the volume's size.
	// It uses the same format as Kubernetes' size fields, e.g. 10Gi
	Size string `json:"size"`
	// MountPath is the path at which the PersistentVolume will be mounted.
	MountPath string `json:"mountPath"`
	// StorageClassName is the name of the StorageClass that the StatefulSet will request.
	StorageClassName string `json:"storageClassName"`
}

type BiomeStatus struct {
	State   BiomeState `json:"state,omitempty"`
	Message string       `json:"message,omitempty"`
}

type BiomeState string

type ServiceV1beta2 struct {
	// Group is the value of the --group flag for the bio client.
	// Defaults to `default`.
	// +optional
	Group *string `json:"group,omitempty"`
	// Topology is the value of the --topology flag for the bio client.
	Topology `json:"topology"`
	// ConfigSecretName is the name of a Secret containing a Biome service's config in TOML format.
	// It will be mounted inside the pod as a file, and it will be used by Biome to configure the service.
	// +optional
	ConfigSecretName *string `json:"configSecretName,omitempty"`
	// The name of the secret that contains the ring key.
	// +optional
	RingSecretName *string `json:"ringSecretName,omitempty"`
	// The name of a secret containing the files directory.  It will be mounted inside the pod
	// as a directory.
	// +optional
	FilesSecretName *string `json:"filesSecretName,omitempty"`
	// Bind is when one service connects to another forming a producer/consumer relationship.
	// +optional
	Bind []Bind `json:"bind,omitempty"`
	// Name is the name of the Biome service that this Biome object represents.
	// This field is used to mount the user.toml file in the correct directory under /bio/user/ in the Pod.
	Name string `json:"name"`
	// Channel is the value of the --channel flag for the bio client.
	// It can be used to track upstream packages in builder channels but will never be used directly by the supervisor.
	// The should only be used in conjunction with the biome updater https://github.com/biome-sh/biome-updater
	// Defaults to `stable`.
	// +optional
	Channel *string `json:"channel,omitempty"`
}

type Bind struct {
	// Name is the name of the bind specified in the Biome configuration files.
	Name string `json:"name"`
	// Service is the name of the service this bind refers to.
	Service string `json:"service"`
	// Group is the group of the service this bind refers to.
	Group string `json:"group"`
}

type Topology string

func (t Topology) String() string {
	return string(t)
}

const (
	BiomeStateCreated   BiomeState = "Created"
	BiomeStateProcessed BiomeState = "Processed"

	TopologyStandalone Topology = "standalone"
	TopologyLeader     Topology = "leader"

	BiomeKind = "Biome"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BiomeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Biome `json:"items"`
}
