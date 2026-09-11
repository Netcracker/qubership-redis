package templates

import (
	"testing"

	types "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	v2 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	corev1 "k8s.io/api/core/v1"
)

func tlsVolumeSecretName(t *testing.T, spec corev1.PodSpec) string {
	t.Helper()
	for _, vol := range spec.Volumes {
		if vol.Name != "tls" {
			continue
		}
		if vol.Projected == nil || len(vol.Projected.Sources) == 0 || vol.Projected.Sources[0].Secret == nil {
			t.Fatalf("tls volume has no projected secret source: %+v", vol)
		}
		return vol.Projected.Sources[0].Secret.Name
	}
	t.Fatalf("tls volume not found among: %+v", spec.Volumes)
	return ""
}

func TestGetRedisDeploymentTemplate_TLSSecretIsAlwaysPerInstance(t *testing.T) {
	tests := []struct {
		name              string
		instanceName      string
		clusterIssuerName string
		wantSecretName    string
	}{
		{
			name:           "self-signed default issuer - logical db",
			instanceName:   "mydb",
			wantSecretName: "mydb-tls",
		},
		{
			name:              "external cluster issuer - logical db",
			instanceName:      "mydb",
			clusterIssuerName: "some-cluster-issuer",
			wantSecretName:    "mydb-tls",
		},
		{
			name:           "self-signed default issuer - direct single instance",
			instanceName:   "redis",
			wantSecretName: "redis-tls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tls := v2.TLS{
				TLS: types.TLS{
					Enabled:            true,
					RootCAFileName:     "ca.crt",
					PrivateKeyFileName: "tls.key",
					SignedCRTFileName:  "tls.crt",
				},
				TLSPort:           6380,
				NonTlsPort:        6379,
				ClusterIssuerName: tt.clusterIssuerName,
			}

			deployment := GetRedisDeploymentTemplate(
				tt.instanceName, "my-ns", "redis:8", []string{}, nil,
				corev1.ResourceRequirements{}, nil, nil, "sa",
				nil, "app", corev1.PullIfNotPresent, tls,
				"", "part-of", "managed-by",
				tt.instanceName+"-credentials",
			)

			got := tlsVolumeSecretName(t, deployment.Spec.Template.Spec)
			if got != tt.wantSecretName {
				t.Errorf("tls secret name = %q, want %q", got, tt.wantSecretName)
			}
		})
	}
}

func TestGetRedisDeploymentTemplate_NoTLSVolumeWhenDisabled(t *testing.T) {
	tls := v2.TLS{TLS: types.TLS{Enabled: false}}

	deployment := GetRedisDeploymentTemplate(
		"mydb", "my-ns", "redis:8", []string{}, nil,
		corev1.ResourceRequirements{}, nil, nil, "sa",
		nil, "app", corev1.PullIfNotPresent, tls,
		"", "part-of", "managed-by",
		"mydb-credentials",
	)

	for _, vol := range deployment.Spec.Template.Spec.Volumes {
		if vol.Name == "tls" {
			t.Fatalf("expected no tls volume when TLS disabled, found one: %+v", vol)
		}
	}
}
