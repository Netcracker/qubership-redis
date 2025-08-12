package templates

import (
	"fmt"
	"regexp"

	constants "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	v2 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/core"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// labels
const (
	AppName              = "app.kubernetes.io/name"
	AppInstance          = "app.kubernetes.io/instance"
	AppVersion           = "app.kubernetes.io/version"
	AppComponent         = "app.kubernetes.io/component"
	AppManagedBy         = "app.kubernetes.io/managed-by"
	AppManagedByOperator = "app.kubernetes.io/managed-by-operator"
	AppProcByOperator    = "app.kubernetes.io/processed-by-operator"
	AppTechnology        = "app.kubernetes.io/technology"
	AppPartOf            = "app.kubernetes.io/part-of"
)

func GetRedisDeploymentTemplate(
	name string,
	namespace string,
	image string,
	args []string,
	env []v13.EnvVar,
	resources v13.ResourceRequirements,
	nodeSelector map[string]string,
	securityContext *v13.PodSecurityContext,
	serviceAccountName string,
	tolerations []v13.Toleration,
	label string,
	redisImagePullPolicy v13.PullPolicy,
	tls v2.TLS,
	priorityClassName string, partOf, managedBy string) *v1.Deployment {
	var r int32 = 1
	probe := &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			TCPSocket: &v13.TCPSocketAction{
				Port: intstr.IntOrString{Type: intstr.Int, IntVal: 6379},
			},
		},
		InitialDelaySeconds: 1,
		TimeoutSeconds:      10,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    10,
	}

	reg, err := regexp.Compile(`([0-9]+\.[0-9]+\.[0-9]+)`)
	if err != nil {
		panic(err)
	}
	labels := map[string]string{
		constants.Name: name,
		constants.App:  name,
		label:          label,
		AppName:        name,
		AppInstance:    fmt.Sprintf("redis-%s", namespace),
		AppVersion:     reg.FindString(image),
		AppComponent:   "operator",
		AppPartOf:      partOf,
		AppManagedBy:   managedBy,
		AppTechnology:  "",
	}

	volumes := []v13.Volume{
		{
			Name: "config",
			VolumeSource: v13.VolumeSource{
				ConfigMap: &v13.ConfigMapVolumeSource{
					LocalObjectReference: v13.LocalObjectReference{
						Name: name,
					},
					Items: []v13.KeyToPath{
						{
							Path: "redis.conf",
							Key:  "config",
						},
					},
				},
			},
		},
		{
			Name: name + "-data",
			VolumeSource: v13.VolumeSource{
				EmptyDir: &v13.EmptyDirVolumeSource{
					Medium: "",
				},
			},
		},
	}

	volumeMounts := []v13.VolumeMount{
		{
			MountPath: "/etc/redis",
			Name:      "config",
		},
		{
			Name:      name + "-data",
			MountPath: "/var/lib/redis/data",
		},
	}

	if tls.Enabled {
		var secretName string
		if tls.ClusterIssuerName == "" {
			secretName = "root-ca"
		} else {
			secretName = tls.CertificateSecretName
		}
		volProj := []v13.VolumeProjection{
			v13.VolumeProjection{
				Secret: &v13.SecretProjection{
					LocalObjectReference: v13.LocalObjectReference{
						Name: secretName,
					},
					Items: []v13.KeyToPath{
						{
							Path: tls.RootCAFileName,
							Key:  tls.RootCAFileName,
						},
						{
							Path: tls.PrivateKeyFileName,
							Key:  tls.PrivateKeyFileName,
						},
						{
							Path: tls.SignedCRTFileName,
							Key:  tls.SignedCRTFileName,
						},
					},
				},
			},
		}
		volumes = append(volumes, v13.Volume{
			Name: "tls",
			VolumeSource: v13.VolumeSource{
				Projected: &v13.ProjectedVolumeSource{
					Sources:     volProj,
					DefaultMode: nil,
				},
			},
		})

		volumeMounts = append(volumeMounts, v13.VolumeMount{
			Name:      "tls",
			MountPath: core.CertPath,
		})
		args = append(args, fmt.Sprintf("--tls-port %d", tls.TLSPort), fmt.Sprintf("--port %d", tls.NonTlsPort),
			fmt.Sprintf("--tls-cert-file %s", fmt.Sprintf("%s/%s", core.CertPath, tls.SignedCRTFileName)),
			fmt.Sprintf("--tls-key-file  %s", fmt.Sprintf("%s/%s", core.CertPath, tls.PrivateKeyFileName)),
			fmt.Sprintf("--tls-ca-cert-file %s", fmt.Sprintf("%s/%s", core.CertPath, tls.RootCAFileName)))
	}

	allowPrivilegeEscalation := false

	return &v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &r,
			Selector: &v12.LabelSelector{
				MatchLabels: map[string]string{
					constants.Name: name,
					constants.App:  name,
					label:          label,
				},
			},
			Template: v13.PodTemplateSpec{
				ObjectMeta: v12.ObjectMeta{
					Labels: labels,
				},
				Spec: v13.PodSpec{
					Containers: []v13.Container{
						v13.Container{
							Name:            name,
							Image:           image,
							ImagePullPolicy: redisImagePullPolicy,
							Args:            args,
							SecurityContext: &v13.SecurityContext{
								Capabilities: &v13.Capabilities{
									Drop: []v13.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							Ports: []v13.ContainerPort{
								v13.ContainerPort{
									Name:          "web",
									ContainerPort: 6379,
									Protocol:      "TCP",
								},
							},
							ReadinessProbe: probe,
							LivenessProbe:  probe,
							Env:            env,
							Resources:      resources,
							VolumeMounts:   volumeMounts,
						},
					},
					Volumes:            volumes,
					NodeSelector:       nodeSelector,
					PriorityClassName:  priorityClassName,
					SecurityContext:    securityContext,
					ServiceAccountName: serviceAccountName,
					Tolerations:        tolerations,
					RestartPolicy:      "Always",
				},
			},
		},
	}
}

func GetRedisServiceTemplate(
	name string,
	namespace string, partOf, managedBy string) *v13.Service {
	return &v13.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				constants.Name: name,
				constants.App:  name,
				AppName:        name,
				AppPartOf:      partOf,
				AppManagedBy:   managedBy,
			},
		},
		Spec: v13.ServiceSpec{
			Ports: []v13.ServicePort{
				{
					Name: "web",
					Port: 6379,
					TargetPort: intstr.IntOrString{
						IntVal: 6379,
					},
				},
			},
			// ClusterIP: v13.ClusterIPNone,
			Selector: map[string]string{
				constants.Name: name,
				constants.App:  name,
			},
		},
	}
}

func GetRedisConfigTemplate(name string, namespace string, config string) *v13.ConfigMap {
	return &v13.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				constants.Name: name,
			},
		},
		Data: map[string]string{
			"config": config,
		},
	}
}
