package impl

import (
	"context"
	"fmt"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v2 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/adapter"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/monitoring"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/redis"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/robotTests"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/utils"
	"github.com/Netcracker/qubership-redis/redis-operator/common"
	rc "github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/redis"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/templates"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RedisServicesCompound struct {
	core.DefaultCompound
}

type RedisServiceBuilder struct {
	core.ExecutableBuilder
}

func (r *RedisServiceBuilder) Build(ctx core.ExecutionContext) core.Executable {

	spec := ctx.Get(constants.ContextSpec).(*v2.DbaasRedisAdapter)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	kubeClient := ctx.Get(constants.ContextClient).(client.Client)
	runtimeScheme := ctx.Get(constants.ContextSchema).(*runtime.Scheme)

	log.Debug("Redis Executable build process is started")
	// It is needed for test proposes. Implementations is changed for module tests
	// TODO: Force key change based on deploy type?
	defaultUtilsHelper := &core.DefaultKubernetesHelperImpl{
		ForceKey: true,
		OwnerKey: false,
		Client:   kubeClient, //TODO we get client from ctx and set it to ctx
	}
	ctx.Set(constants.KubernetesHelperImpl, defaultUtilsHelper)

	var compound core.ExecutableCompound = &RedisServicesCompound{}

	if spec.Spec.Dbaas.Install {
		compound.AddStep((&adapter.AdapterBuilder{}).Build(ctx))
	} else {
		compound.AddStep((&redis.RedisBuilder{}).Build(ctx))
	}

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Update Existing DBs",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *v2.DbaasRedisAdapter, log *zap.Logger) error {

			log.Info("Updating Existing Redis Databases...")
			dcs := &v1.DeploymentList{}
			opts := []client.ListOption{
				client.InNamespace(request.Namespace),
				client.MatchingLabelsSelector{labels.SelectorFromSet(map[string]string{spec.Spec.Redis.Label: spec.Spec.Redis.Label})},
			}

			errList := kubeClient.List(context.TODO(), dcs, opts...)
			core.PanicError(errList, log.Error, "Could not retrieve list of DCs of redis")

			for _, dc := range dcs.Items {
				redisName := dc.ObjectMeta.Name
				envs := dc.Spec.Template.Spec.Containers[0].Env
				envs = common.MergeEnvs(envs, common.GetRedisEnvs(spec.Spec.Redis.TLS.TLS))
				var tolerations []corev1.Toleration
				if cr.Spec.Policies != nil {
					tolerations = cr.Spec.Policies.Tolerations
				}

				redisDC := templates.GetRedisDeploymentTemplate(redisName, request.Namespace, spec.Spec.Redis.DockerImage,
					spec.Spec.Redis.Args,
					envs,
					*spec.Spec.Redis.Resources,
					spec.Spec.Redis.NodeLabels,
					spec.Spec.PodSecurityContext,
					spec.Spec.ServiceAccountName,
					tolerations,
					spec.Spec.Redis.Label,
					spec.Spec.ImagePullPolicy,
					spec.Spec.Redis.TLS,
					spec.Spec.Redis.PriorityClassName,
					spec.Spec.PartOf, spec.Spec.ManagedBy,
				)

				if spec.Spec.Redis.TLS.ClusterIssuerName != "" {
					common.UpdateCertificate(spec.Spec.Redis.TLS.Enabled, spec.Spec.Redis.TLS.ClusterIssuerName, dc.ObjectMeta.Name, request.Namespace, kubeClient, runtimeScheme)
				}

				var updateErr error
				for i := 0; i < 3; i++ {
					updateErr = core.CreateOrUpdateRuntimeObject(kubeClient, runtimeScheme, nil, redisDC, redisDC.ObjectMeta, true)
					if updateErr == nil {
						break
					}
				}
				core.PanicError(updateErr, log.Error, "Failed to update existing DB")
			}

			return nil
		},
	})

	if spec.Spec.Monitoring.Install {
		compound.AddStep((&monitoring.MonitoringBuilder{}).Build(ctx))
	}

	if spec.Spec.RobotTests.Install {
		compound.AddStep((&robotTests.RobotBuilder{}).Build(ctx))
	}

	log.Debug("Redis Executable has been built")

	return compound
}

type RedisPreDeployBuilder struct {
	core.ExecutableBuilder
}

func (r *RedisPreDeployBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v2.DbaasRedisAdapter)

	var compound core.ExecutableCompound = &RedisServicesCompound{}
	redisClient := rc.RedisClient{}
	ctx.Set(utils.ContextRedis, redisClient)

	if spec.Spec.Dbaas.Install {
		// Running DBaaS Server only if spec has no changes
		// So every required resource is already deployed

		spec := ctx.Get(constants.ContextSpec).(*v2.DbaasRedisAdapter)
		kubeClient := ctx.Get(constants.ContextClient).(client.Client)
		request := ctx.Get(constants.ContextRequest).(reconcile.Request)
		runtimeScheme := ctx.Get(constants.ContextSchema).(*runtime.Scheme)
		if !ctx.Get(constants.ContextSpecHasChanges).(bool) {
			//we need this step for upgrade/reconcile scenarios, when spec is not change and Builder is not executed
			compound.AddStep(&utils.SimpleCtxExecutable{
				StepName: "Adapter Server",
				ExecuteFunc: func(ctx core.ExecutionContext, cr *v2.DbaasRedisAdapter, log *zap.Logger) error {
					return adapter.RunDBaaSServer(spec, redisClient, kubeClient, runtimeScheme, log, request.Namespace, true)
				},
			})

		}
	}

	return compound
}
