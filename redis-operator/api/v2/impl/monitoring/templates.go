package monitoring

import (
	"fmt"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	utils2 "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	netcrackerv1 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/utils"
	"github.com/Netcracker/qubership-redis/redis-operator/common"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v13 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	serviceName   = "redis-monitoring-agent"
	configMapName = serviceName + "-config"
)

var execCommand = []string{"/bin/sh", "-c", "/health.sh"}

type MonitoringCompound struct {
	core.MicroServiceCompound
}

type MonitoringBuilder struct {
	core.ExecutableBuilder
}

func (r *MonitoringBuilder) Build(ctx core.ExecutionContext) core.Executable {

	compound := MonitoringCompound{}
	compound.ServiceName = serviceName
	compound.CalcDeployType = func(ctx core.ExecutionContext) (deployType core.MicroServiceDeployType, err error) {
		return core.CleanDeploy, nil
	}
	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Monitoring Service",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			kubeClient := ctx.Get(constants.ContextClient).(client.Client)
			template := Service(cr)

			core.DeleteRuntimeObject(kubeClient, &corev1.Service{
				ObjectMeta: template.ObjectMeta,
			})

			err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)
			core.PanicError(err, log.Error, "Monitoring service creation failed")

			return nil
		},
	})

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Monitoring Deployment",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			request := ctx.Get(constants.ContextRequest).(reconcile.Request)
			helperImpl := ctx.Get(constants.KubernetesHelperImpl).(core.KubernetesHelper)

			deployment := MonitoringDeployment(cr)

			delErr := helperImpl.DeleteDeploymentAndPods(deployment.Name, request.Namespace, cr.Spec.WaitTimeout)
			core.PanicError(delErr, log.Error, "Deletion failed")

			err := utils.CreateRuntimeObjectContextWrapper(ctx, deployment, deployment.ObjectMeta)
			core.PanicError(err, log.Error, "Monitoring deployment creation failed")

			log.Debug("Waiting for monitoring is ready")
			err = helperImpl.WaitForPodsReady(
				deployment.Spec.Template.ObjectMeta.Labels,
				request.Namespace,
				1,
				cr.Spec.WaitTimeout)
			core.PanicError(err, log.Error, "Failed waiting for monitoring pod is ready")

			return nil
		},
	})

	return &compound
}

func (r *MonitoringCompound) Condition(ctx core.ExecutionContext) (bool, error) {
	return true, nil
}

func MonitoringDeployment(cr *netcrackerv1.DbaasRedisAdapter) *appsv1.Deployment {
	spec := cr.Spec.Monitoring
	image := spec.DockerImage
	resources := spec.Resources
	monitoringImagePullPolicy := cr.Spec.ImagePullPolicy

	var envs []corev1.EnvVar

	readinessProbe := &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			Exec: &v13.ExecAction{
				Command: execCommand,
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    10,
	}

	livenessProbe := &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			Exec: &v13.ExecAction{
				Command: execCommand,
			},
		},
		InitialDelaySeconds: 20,
		TimeoutSeconds:      10,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    10,
	}

	telegrafProcessCmd := "/sbin/tini -s -- telegraf 2>&1"
	if cr.Spec.VaultRegistration.Enabled {
		telegrafProcessCmd = fmt.Sprintf("%s %s", utils2.GetVaultEnvPath(), telegrafProcessCmd)
	}

	envs = append(envs,
		utils2.GetPlainTextEnvVar("NAMESPACE", cr.Namespace),
		utils2.GetPlainTextEnvVar("REDIS_INSTANCES_LABEL_SELECTOR", cr.Spec.Redis.Parameters.Label),
		utils2.GetPlainTextEnvVar("TELEGRAF_CONF_CONFIGMAP", "redis-monitoring-agent-config"),
		utils2.GetPlainTextEnvVar("CMD_MONITORING_COMMAND", telegrafProcessCmd),
	)

	var tolerations []corev1.Toleration
	if cr.Spec.Policies != nil {
		tolerations = cr.Spec.Policies.Tolerations
	}

	allowPrivilegeEscalation := false

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app":               serviceName,
				utils.AppName:       serviceName,
				utils.AppInstance:   cr.Spec.Instance,
				utils.AppVersion:    cr.Spec.ArtifactDescriptorVersion,
				utils.AppComponent:  "operator",
				utils.AppPartOf:     cr.Spec.PartOf,
				utils.AppManagedBy:  cr.Spec.ManagedBy,
				utils.AppTechnology: "",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": serviceName},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":               serviceName,
						utils.AppName:       serviceName,
						utils.AppInstance:   cr.Spec.Instance,
						utils.AppVersion:    cr.Spec.ArtifactDescriptorVersion,
						utils.AppComponent:  "operator",
						utils.AppPartOf:     cr.Spec.PartOf,
						utils.AppTechnology: "",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            serviceName,
							Image:           image,
							ImagePullPolicy: monitoringImagePullPolicy,
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8125,
									Protocol:      "TCP",
								},
								{
									ContainerPort: 8094,
									Protocol:      "TCP",
								},
								{
									ContainerPort: 8092,
									Protocol:      "UDP",
								},
								{
									ContainerPort: 9273,
									Protocol:      "TCP",
								},
							},
							ReadinessProbe: readinessProbe,
							LivenessProbe:  livenessProbe,
							Env:            envs,
							Resources:      *resources,
						},
					},
					NodeSelector:       spec.NodeLabels,
					SecurityContext:    cr.Spec.PodSecurityContext,
					ServiceAccountName: cr.Spec.ServiceAccountName,
					PriorityClassName:  spec.PriorityClassName,
					Tolerations:        tolerations,
					RestartPolicy:      "Always",
				},
			},
		},
	}

	utils2.VaultPodSpec(&deployment.Spec.Template.Spec, []string{}, cr.Spec.VaultRegistration)
	//TODO: The method above overrides command for monitoring, but it is not required because vault process is child in this case
	//Restore command for pod
	deployment.Spec.Template.Spec.Containers[0].Command = []string{}
	deployment.Spec.Template.Spec.Containers[0].Args = []string{}

	utils2.TLSSpecUpdate(&deployment.Spec.Template.Spec, common.RootCertPath, cr.Spec.Redis.TLS.TLS)

	return deployment
}

func Service(cr *netcrackerv1.DbaasRedisAdapter) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: cr.Namespace,
			Labels:    map[string]string{"app": serviceName, "name": serviceName},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "telegraf-statsd",
					Port:     8125,
					Protocol: "TCP",
					TargetPort: intstr.IntOrString{
						IntVal: 8125,
					},
				},
				{
					Name:     "telegraf-tcp",
					Port:     8094,
					Protocol: "TCP",
					TargetPort: intstr.IntOrString{
						IntVal: 8094,
					},
				},
				{
					Name:     "telegraf-udp",
					Port:     8092,
					Protocol: "UDP",
					TargetPort: intstr.IntOrString{
						IntVal: 8092,
					},
				},
				{
					Name:     "http",
					Port:     9273,
					Protocol: "TCP",
					TargetPort: intstr.IntOrString{
						IntVal: 9273,
					},
				},
			},
			Selector: map[string]string{"app": serviceName},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}
