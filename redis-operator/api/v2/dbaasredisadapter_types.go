/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DbaasRedisAdapterStatus defines the observed state of DbaasRedisAdapter
type DbaasRedisAdapterStatus struct {
	Conditions []types.ServiceStatusCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// DbaasRedisAdapter is the Schema for the dbaasredisadapters API
type DbaasRedisAdapter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DbaasRedisAdapterSpec   `json:"spec,omitempty"`
	Status DbaasRedisAdapterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DbaasRedisAdapterList contains a list of DbaasRedisAdapter
type DbaasRedisAdapterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DbaasRedisAdapter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DbaasRedisAdapter{}, &DbaasRedisAdapterList{})
}

type DbaasRedisAdapterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Dbaas                     `json:"dbaas,omitempty"`
	Redis                     `json:"redis,omitempty"`
	Monitoring                `json:"monitoringAgent,omitempty"`
	RobotTests                `json:"robotTests"`
	WaitTimeout               int                     `json:"waitTimeout,omitempty"`
	PodSecurityContext        *v1.PodSecurityContext  `json:"securityContext,omitempty"`
	Policies                  *Policies               `json:"policies,omitempty"`
	DeploymentVersion         string                  `json:"deploymentVersion,omitempty"`
	VaultRegistration         types.VaultRegistration `json:"vaultRegistration,omitempty"`
	ServiceAccountName        string                  `json:"serviceAccountName"`
	ImagePullPolicy           v1.PullPolicy           `json:"imagePullPolicy,omitempty" common:"true"`
	TLS                       `json:"tls,omitempty" common:"true"`
	DeploymentSessionId       string `json:"deploymentSessionId,omitempty"`
	ArtifactDescriptorVersion string `json:"artifactDescriptorVersion,omitempty"`
	PartOf                    string `json:"partOf,omitempty"`
	ManagedBy                 string `json:"managedBy,omitempty"`
	Instance                  string `json:"instance,omitempty"`
}

type DbaasAdapter struct {
	Username          string          `json:"username,omitempty"`
	SecretName        string          `json:"secretName,omitempty"`
	SupportedFeatures map[string]bool `json:"supportedFeatures,omitempty"`
	ApiVersion        string          `json:"apiVersion,omitempty"`
	CreateDBTimeout   int             `json:"createDBTimeout,omitempty"`
}

type DbaasAggregator struct {
	Username                           string            `json:"username,omitempty"`
	SecretName                         string            `json:"secretName,omitempty"`
	Address                            string            `json:"address,omitempty"`
	PhysicalDatabaseIdentifier         string            `json:"physicalDatabaseIdentifier,omitempty"`
	PhysicalDatabaseLabels             map[string]string `json:"physicalDatabaseLabels,omitempty"`
	DbaasAggregatorRegistrationAddress string            `json:"dbaasAggregatorRegistrationAddress,omitempty"`
}

type Dbaas struct {
	Install    bool             `json:"install"`
	Adapter    *DbaasAdapter    `json:"adapter,omitempty"`
	Aggregator *DbaasAggregator `json:"aggregator,omitempty"`
	TLS        TLS              `json:"tls,omitempty" common:"true"`
}

type Redis struct {
	DockerImage       string   `json:"dockerImage"`
	Args              []string `json:"args"`
	Parameters        `json:"parameters,omitempty"`
	Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
	Maxmem            string                   `json:"maxmem,omitempty"`
	SecretName        string                   `json:"secretName,omitempty"`
	NodeLabels        map[string]string        `json:"nodeLabels,omitempty"`
	TLS               TLS                      `json:"tls,omitempty" common:"true"`
	PriorityClassName string                   `json:"priorityClassName,omitempty"`
}

type InfluxSettings struct {
	Host            string `json:"host,omitempty"`
	Database        string `json:"database,omitempty"`
	RetentionPolicy string `json:"retentionPolicy,omitempty"`
	User            string `json:"user,omitempty"`
	SecretName      string `json:"secretName,omitempty"`
}

type Monitoring struct {
	Install           bool                     `json:"install,omitempty"`
	DockerImage       string                   `json:"dockerImage,omitempty"`
	NodeLabels        map[string]string        `json:"nodeLabels,omitempty"`
	Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
	InfluxDB          *InfluxSettings          `json:"influxDB,omitempty"`
	MetricCollector   string                   `json:"metricCollector,omitempty"`
	PriorityClassName string                   `json:"priorityClassName,omitempty"`
}

type RobotTests struct {
	Install           bool                     `json:"install,omitempty"`
	DockerImage       string                   `json:"dockerImage,omitempty"`
	Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
	Tags              string                   `json:"tags,omitempty"`
	NodeLabels        map[string]string        `json:"nodeLabels,omitempty"`
	PriorityClassName string                   `json:"priorityClassName,omitempty"`
}

type TLS struct {
	types.TLS `json:",omitempty"`

	//Port to accept tls connections.
	TLSPort int `json:"tlsPort,omitempty"`
	//Port to accept non-tls connections. 0 to disable the non-TLS port completely
	NonTlsPort        int    `json:"nonTlsPort,omitempty"`
	ClusterIssuerName string `json:"clusterIssuerName,omitempty"`
}

type Parameters struct {
	Label string `json:"label"`
}

type Policies struct {
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}
