package common

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	cm "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RootCertPath        = "/usr/ssl/"
	RedisPort    int    = 6379
	Username     string = "username"
	Password     string = "password"
	Charset             = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var RedisContainerEntryPoint = []string{"/run_entry.sh"}

var TLSSecretNamePattern = "%s-tls"

// Must match the Issuer name defined in redis-tls-issuer.yaml.
var RedisTLSIssuerName = "redis-tls-issuer"

// main.go never registers the cert-manager scheme at startup, so it's done
// here on first use instead - once per process, not once per call, since
// UpdateCertificate runs on every reconcile for every existing database.
var registerCertManagerScheme sync.Once
var certManagerSchemeErr error

func UpdateCertificate(tlsEnabled bool, clusterIssuerName, logicalDatabaseName, namespace string, kubeClient client.Client, runtimeScheme *runtime.Scheme) error {
	if !tlsEnabled {
		return nil
	}

	certificateTemplate := GetCertificateTemplate(logicalDatabaseName, namespace, clusterIssuerName)

	registerCertManagerScheme.Do(func() {
		certManagerSchemeErr = cm.AddToScheme(runtimeScheme)
	})
	if certManagerSchemeErr != nil {
		return certManagerSchemeErr
	}

	certifErr := core.CreateOrUpdateRuntimeObject(kubeClient, runtimeScheme, nil, certificateTemplate,
		v1.ObjectMeta{Name: certificateTemplate.GetName(), Namespace: certificateTemplate.GetNamespace()}, true)
	if certifErr != nil {
		return certifErr
	}
	return nil
}

func GetCertificateTemplate(dbName, namespace, clusterIssuerName string) client.Object {

	var ref cmeta.ObjectReference
	if clusterIssuerName != "" {
		ref = cmeta.ObjectReference{
			Name:  clusterIssuerName,
			Kind:  "ClusterIssuer",
			Group: "cert-manager.io",
		}
	} else {
		ref = cmeta.ObjectReference{
			Name:  RedisTLSIssuerName,
			Kind:  "Issuer",
			Group: "cert-manager.io",
		}
	}

	return &cm.Certificate{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-certificate", dbName),
			Namespace: namespace,
		},
		Spec: cm.CertificateSpec{
			SecretName: fmt.Sprintf(TLSSecretNamePattern, dbName),
			Duration:   &v1.Duration{Duration: time.Duration(365*24) * time.Hour},
			CommonName: "redis-cn",
			// Both forms are listed so the cert is valid no matter which one the
			// connecting client actually resolves.
			DNSNames: []string{
				fmt.Sprintf("%s.%s", dbName, namespace),
				fmt.Sprintf("%s.%s.svc", dbName, namespace),
			},
			// Leaf cert chained to IssuerRef - must not be a CA itself.
			IsCA: false,
			Usages: []cm.KeyUsage{
				cm.UsageDigitalSignature,
				cm.UsageKeyEncipherment,
				cm.UsageServerAuth,
				cm.UsageClientAuth,
			},
			PrivateKey: &cm.CertificatePrivateKey{
				Algorithm: cm.RSAKeyAlgorithm,
				Encoding:  cm.PKCS1,
				Size:      2048,
			},
			IssuerRef: ref,
		},
	}
}

func GetRedisEnvs(tls types.TLS) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "REDIS_PORT",
			Value: fmt.Sprint(RedisPort),
		},
		utils.GetPlainTextEnvVar("TLS_ENABLED", strconv.FormatBool(tls.Enabled)),
		utils.GetPlainTextEnvVar("TLS_ROOTCERT", "/usr/ssl/ca.crt"),
	}
}

func MergeEnvs(from, to []corev1.EnvVar) []corev1.EnvVar {
	if len(to) == 0 {
		return from
	}
	result := to
	for _, envFrom := range from {
		toPut := envFrom
		contains := false
		for _, envTo := range to {
			if envFrom.Name == envTo.Name {
				contains = true
			}
		}
		if !contains {
			result = append(result, toPut)
		}
	}

	return result
}
