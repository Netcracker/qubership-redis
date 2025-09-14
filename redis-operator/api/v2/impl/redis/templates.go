package redis

import (
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	utils2 "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	netcrackerv1 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/utils"
	"github.com/Netcracker/qubership-redis/redis-operator/common"
	core2 "github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/core"
	service "github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/services"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/templates"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	serviceName   = "redis"
	configMapName = "redis-config"
)

type RedisCompound struct {
	core.MicroServiceCompound
}

type RedisBuilder struct {
	core.ExecutableBuilder
}

func (r *RedisBuilder) Build(ctx core.ExecutionContext) core.Executable {

	spec := ctx.Get(constants.ContextSpec).(*netcrackerv1.DbaasRedisAdapter)

	compound := RedisCompound{}
	compound.ServiceName = serviceName
	compound.CalcDeployType = func(ctx core.ExecutionContext) (deployType core.MicroServiceDeployType, err error) {
		return core.CleanDeploy, nil
	}

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Redis Service",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			kubeClient := ctx.Get(constants.ContextClient).(client.Client)
			request := ctx.Get(constants.ContextRequest).(reconcile.Request)
			template := templates.GetRedisServiceTemplate(
				core2.Redis,
				request.Namespace, spec.Spec.PartOf, spec.Spec.ManagedBy)

			core.DeleteRuntimeObject(kubeClient, &corev1.Service{
				ObjectMeta: template.ObjectMeta,
			})

			err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)
			core.PanicError(err, log.Error, "Redis service creation failed")

			return nil
		},
	})

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Redis ConfigMap",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			request := ctx.Get(constants.ContextRequest).(reconcile.Request)
			client := ctx.Get(constants.ContextClient).(client.Client)

			redisConfig := service.GetRedisDefaultConfigMap(client, request.Namespace, log)
			configString := service.RedisMapConfigToString(redisConfig)
			template := templates.GetRedisConfigTemplate(
				core2.Redis,
				request.Namespace,
				configString)

			core.DeleteRuntimeObjectWithCheck(client, &corev1.ConfigMap{
				ObjectMeta: template.ObjectMeta}, 60)

			err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)
			core.PanicError(err, log.Error, "Redis ConfigMap creation failed")

			return nil
		},
	})

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Redis Deployment",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			request := ctx.Get(constants.ContextRequest).(reconcile.Request)
			helperImpl := ctx.Get(constants.KubernetesHelperImpl).(core.KubernetesHelper)
			redisSpec := cr.Spec.Redis

			var tolerations []corev1.Toleration
			if cr.Spec.Policies != nil {
				tolerations = cr.Spec.Policies.Tolerations
			}

			envs := common.GetRedisEnvs(redisSpec.TLS.TLS)
			envs = append(envs, utils2.GetSecretEnvVar("REDIS_PASSWORD", redisSpec.SecretName, constants.Password))
			deployment := templates.GetRedisDeploymentTemplate(
				core2.Redis,
				request.Namespace,
				redisSpec.DockerImage,
				redisSpec.Args,
				envs,
				*redisSpec.Resources,
				redisSpec.NodeLabels,
				cr.Spec.PodSecurityContext,
				cr.Spec.ServiceAccountName,
				tolerations,
				redisSpec.Label,
				spec.Spec.ImagePullPolicy,
				spec.Spec.Redis.TLS,
				spec.Spec.Redis.PriorityClassName, spec.Spec.PartOf, spec.Spec.ManagedBy,
			)

			delErr := helperImpl.DeleteDeploymentAndPods(deployment.Name, request.Namespace, cr.Spec.WaitTimeout)
			core.PanicError(delErr, log.Error, "Deletion failed")

			err := utils.CreateRuntimeObjectContextWrapper(ctx, deployment, deployment.ObjectMeta)
			core.PanicError(err, log.Error, "Redis deployment creation failed")

			log.Debug("Waiting for Redis is ready")
			err = helperImpl.WaitForPodsReady(
				deployment.Spec.Template.ObjectMeta.Labels,
				request.Namespace,
				1,
				cr.Spec.WaitTimeout)
			core.PanicError(err, log.Error, "Failed waiting for Redis pod is ready")

			return nil
		},
	})

	return &compound
}
