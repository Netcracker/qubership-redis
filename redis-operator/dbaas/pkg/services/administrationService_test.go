package service

import (
	"testing"

	cm "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	v1 "k8s.io/api/core/v1"
)

func TestGetResourcesMapping_TLSSecret(t *testing.T) {
	adminService := &AdministrationService{namespace: "test-ns"}

	mapping := adminService.getResourcesMapping("mydb")

	tlsResource, ok := mapping["TLSSecret"]
	if !ok {
		t.Fatalf("getResourcesMapping() has no \"TLSSecret\" entry, got kinds: %v", mappingKinds(mapping))
	}
	if tlsResource.name != "mydb-tls" {
		t.Errorf("TLSSecret name = %q, want %q", tlsResource.name, "mydb-tls")
	}
	secret, ok := tlsResource.object.(*v1.Secret)
	if !ok {
		t.Fatalf("TLSSecret object is %T, want *v1.Secret", tlsResource.object)
	}
	if secret.Namespace != "test-ns" {
		t.Errorf("TLSSecret namespace = %q, want %q", secret.Namespace, "test-ns")
	}

	// credentials Secret entry must be unaffected by adding TLSSecret
	credsResource, ok := mapping["Secret"]
	if !ok {
		t.Fatalf("getResourcesMapping() has no \"Secret\" entry")
	}
	if credsResource.name != "mydb-credentials" {
		t.Errorf("Secret (credentials) name = %q, want %q", credsResource.name, "mydb-credentials")
	}

	certResource, ok := mapping["Certificate"]
	if !ok {
		t.Fatalf("getResourcesMapping() has no \"Certificate\" entry")
	}
	if certResource.name != "mydb-certificate" {
		t.Errorf("Certificate name = %q, want %q", certResource.name, "mydb-certificate")
	}
	if _, ok := certResource.object.(*cm.Certificate); !ok {
		t.Fatalf("Certificate object is %T, want *cm.Certificate", certResource.object)
	}
}

func TestGetResourcesMapping_TLSSecret_IdempotentOnDeletePath(t *testing.T) {
	adminService := &AdministrationService{namespace: "test-ns"}

	// DropResources re-derives the mapping using the already-suffixed name it
	// got back from the initial getDBResources() call - must not double-suffix.
	mapping := adminService.getResourcesMapping("mydb-tls")

	tlsResource := mapping["TLSSecret"]
	if tlsResource.name != "mydb-tls" {
		t.Errorf("TLSSecret name on delete path = %q, want %q (no double suffixing)", tlsResource.name, "mydb-tls")
	}
}

func mappingKinds(mapping map[string]DBResourceMapping) []string {
	kinds := make([]string, 0, len(mapping))
	for k := range mapping {
		kinds = append(kinds, k)
	}
	return kinds
}
