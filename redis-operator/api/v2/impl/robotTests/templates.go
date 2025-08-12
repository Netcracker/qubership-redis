package robotTests

import (
	"fmt"
	"strconv"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	utils2 "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	netcrackerv1 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/adapter"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/utils"
	"github.com/Netcracker/qubership-redis/redis-operator/common"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServiceName = "robot-tests"
	Name        = "name"
)

type RobotTestsCompound struct {
	core.MicroServiceCompound
}

type RobotBuilder struct {
	core.ExecutableBuilder
}

func (r *RobotBuilder) Build(ctx core.ExecutionContext) core.Executable {

	compound := RobotTestsCompound{}
	compound.ServiceName = ServiceName
	compound.CalcDeployType = func(ctx core.ExecutionContext) (deployType core.MicroServiceDeployType, err error) {
		return core.CleanDeploy, nil
	}

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Robot Deployment",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			// request := ctx.Get(constants.ContextRequest).(reconcile.Request)
			helperImpl := ctx.Get(constants.KubernetesHelperImpl).(core.KubernetesHelper)

			dc := RobotDeployment(cr)

			err := helperImpl.DeleteDeploymentAndPods(dc.Name, cr.Namespace, cr.Spec.WaitTimeout)
			core.PanicError(err, log.Error, "RobotTests deployment config processing failed")

			err = utils.CreateRuntimeObjectContextWrapper(ctx, dc, dc.ObjectMeta)
			core.PanicError(err, log.Error, "RobotTests deployment config processing failed")

			log.Debug("Waiting for robot is ready")
			err = helperImpl.WaitForTestsReady(
				dc.Name,
				dc.Namespace,
				cr.Spec.WaitTimeout)
			core.PanicError(err, log.Error, "RobotTests failed")

			return nil
		},
	})

	return &compound
}

func RobotDeployment(cr *netcrackerv1.DbaasRedisAdapter) *v1.Deployment {
	spec := cr.Spec.RobotTests
	image := spec.DockerImage
	resources := spec.Resources
	robotTestImagePullPolicy := cr.Spec.ImagePullPolicy
	tlsEnabled := utils.IsTLSEnableForDBAAS(cr.Spec.Dbaas.Aggregator.DbaasAggregatorRegistrationAddress, cr.Spec.Redis.TLS.TLS.Enabled)

	envs := []corev1.EnvVar{
		{
			Name:  "OPENSHIFT_WORKSPACE_WA",
			Value: cr.Namespace,
		},
		{
			Name:  "DBAAS_AGGREGATOR_REGISTRATION_ADDRESS",
			Value: cr.Spec.Dbaas.Aggregator.DbaasAggregatorRegistrationAddress,
		},
		{
			Name:  "CLOUD_URL_WA",
			Value: "https://kubernetes.default.svc.cluster.local",
		},
		{
			Name:  "REDIS_PORT",
			Value: "6379",
		},
		{
			Name:  "REDIS_PASSWORD",
			Value: "redis",
		},
		{
			Name:  "REDIS_DBAAS_ADAPTER_HOST",
			Value: fmt.Sprintf("%s.%s.svc", adapter.ServiceName, cr.Namespace),
		},
		{
			Name:  "REDIS_DBAAS_ADAPTER_PORT",
			Value: strconv.Itoa(int(utils.GetHTTPPort(tlsEnabled))),
		},
		{
			Name:  "DBAAS_ADAPTER_API_VERSION",
			Value: cr.Spec.Dbaas.Adapter.ApiVersion,
		},
		{
			Name:  "TAGS",
			Value: spec.Tags,
		},
		{
			Name:  "WAIT_TIMEOUT",
			Value: strconv.Itoa(cr.Spec.WaitTimeout),
		},
		{
			Name:  "DBAAS_ENABLED",
			Value: strconv.FormatBool(cr.Spec.Dbaas.Install),
		},
		utils2.GetPlainTextEnvVar("STATUS_CUSTOM_RESOURCE_PATH", fmt.Sprintf("apps/v1/%s/deployments/robot-tests", cr.Namespace)),
		utils2.GetPlainTextEnvVar("STATUS_WRITING_ENABLED", "true"),
	}

	if cr.Spec.Dbaas.Install {
		envs = append(envs,
			utils2.GetSecretEnvVar("REDIS_DBAAS_USER", cr.Spec.Dbaas.Adapter.SecretName, common.Username),
			utils2.GetSecretEnvVar("REDIS_DBAAS_PASSWORD", cr.Spec.Dbaas.Adapter.SecretName, common.Password),
		)
	}

	var tolerations []corev1.Toleration
	if cr.Spec.Policies != nil {
		tolerations = cr.Spec.Policies.Tolerations
	}

	allowPrivilegeEscalation := false
	var replicas int32 = 1
	dc := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app":               ServiceName,
				utils.AppName:       ServiceName,
				utils.AppInstance:   cr.Spec.Instance,
				utils.AppVersion:    cr.Spec.ArtifactDescriptorVersion,
				utils.AppComponent:  "operator",
				utils.AppPartOf:     cr.Spec.PartOf,
				utils.AppManagedBy:  cr.Spec.ManagedBy,
				utils.AppTechnology: "",
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					Name: ServiceName,
				},
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cr.Namespace,
					Labels: map[string]string{
						Name:                ServiceName,
						utils.AppName:       ServiceName,
						utils.AppInstance:   cr.Spec.Instance,
						utils.AppVersion:    cr.Spec.ArtifactDescriptorVersion,
						utils.AppComponent:  "operator",
						utils.AppPartOf:     cr.Spec.PartOf,
						utils.AppTechnology: "python",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cr.Spec.ServiceAccountName,
					PriorityClassName:  spec.PriorityClassName,
					NodeSelector:       spec.NodeLabels,
					SecurityContext:    cr.Spec.PodSecurityContext,
					Tolerations:        tolerations,
					RestartPolicy:      corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:            ServiceName,
							Image:           image,
							ImagePullPolicy: robotTestImagePullPolicy,
							Env:             envs,
							Resources:       *resources,
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
						},
					},
				},
			},
		},
	}

	utils2.VaultPodSpec(&dc.Spec.Template.Spec, []string{"/docker-entrypoint.sh", "run-robot"}, cr.Spec.VaultRegistration)

	utils2.TLSSpecUpdate(&dc.Spec.Template.Spec, common.RootCertPath, cr.Spec.Redis.TLS.TLS)

	return dc
}
