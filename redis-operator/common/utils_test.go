package common

import (
	"fmt"
	"reflect"
	"testing"

	cm "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestGetCertificateTemplate_SAN(t *testing.T) {
	tests := []struct {
		name              string
		dbName            string
		namespace         string
		clusterIssuerName string
		wantDNSNames      []string
		wantIssuerKind    string
		wantIssuerName    string
	}{
		{
			name:           "self-signed default issuer",
			dbName:         "mydb",
			namespace:      "myns",
			wantDNSNames:   []string{"mydb.myns", "mydb.myns.svc"},
			wantIssuerKind: "Issuer",
			wantIssuerName: RedisTLSIssuerName,
		},
		{
			name:              "external cluster issuer",
			dbName:            "mydb",
			namespace:         "myns",
			clusterIssuerName: "my-cluster-issuer",
			wantDNSNames:      []string{"mydb.myns", "mydb.myns.svc"},
			wantIssuerKind:    "ClusterIssuer",
			wantIssuerName:    "my-cluster-issuer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := GetCertificateTemplate(tt.dbName, tt.namespace, tt.clusterIssuerName)
			cert, ok := obj.(*cm.Certificate)
			if !ok {
				t.Fatalf("GetCertificateTemplate() returned %T, want *cm.Certificate", obj)
			}

			if !reflect.DeepEqual(cert.Spec.DNSNames, tt.wantDNSNames) {
				t.Errorf("DNSNames = %v, want %v", cert.Spec.DNSNames, tt.wantDNSNames)
			}
			if cert.Spec.SecretName != fmt.Sprintf(TLSSecretNamePattern, tt.dbName) {
				t.Errorf("SecretName = %q, want %q", cert.Spec.SecretName, fmt.Sprintf(TLSSecretNamePattern, tt.dbName))
			}
			if cert.Spec.IssuerRef.Kind != tt.wantIssuerKind {
				t.Errorf("IssuerRef.Kind = %q, want %q", cert.Spec.IssuerRef.Kind, tt.wantIssuerKind)
			}
			if cert.Spec.IssuerRef.Name != tt.wantIssuerName {
				t.Errorf("IssuerRef.Name = %q, want %q", cert.Spec.IssuerRef.Name, tt.wantIssuerName)
			}
			if !cert.Spec.IsCA {
				t.Errorf("IsCA = false, want true")
			}
		})
	}
}

func TestMergeEnvs(t *testing.T) {
	type args struct {
		from []corev1.EnvVar
		to   []corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want []corev1.EnvVar
	}{
		{
			name: "Merge with intersections, len(from) < len(to) ",
			args: args{
				from: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}},
				to:   []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
			},
			want: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
		},
		{
			name: "Merge with intersections, len(from) > len(to) ",
			args: args{
				from: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
				to:   []corev1.EnvVar{{Name: "foo1", Value: "bar1"}},
			},
			want: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
		},
		{
			name: "Merge with no intersections",
			args: args{
				from: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
				to:   []corev1.EnvVar{{Name: "foo3", Value: "bar3"}},
			},
			want: []corev1.EnvVar{{Name: "foo3", Value: "bar3"}, {Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
		},
		{
			name: "Merge empty first",
			args: args{
				from: []corev1.EnvVar{},
				to:   []corev1.EnvVar{{Name: "foo", Value: "bar"}},
			},
			want: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
		},
		{
			name: "Merge empty second",
			args: args{
				from: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
				to:   []corev1.EnvVar{},
			},
			want: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeEnvs(tt.args.from, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeEnvs() = %v, want %v", got, tt.want)
			}
		})
	}
}
