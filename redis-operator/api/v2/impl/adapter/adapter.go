package adapter

import (
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	netcrackerv1 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/utils"
	rc "github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/redis"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ServiceName  = "dbaas-redis-adapter"
	OperatorName = "dbaas-redis-operator"
)

type AdapterCompound struct {
	core.MicroServiceCompound
}

type AdapterBuilder struct {
	core.ExecutableBuilder
}

func (r *AdapterBuilder) Build(ctx core.ExecutionContext) core.Executable {

	spec := ctx.Get(constants.ContextSpec).(*netcrackerv1.DbaasRedisAdapter)
	kubeClient := ctx.Get(constants.ContextClient).(client.Client)
	logger := ctx.Get(constants.ContextLogger).(*zap.Logger)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	runtimeScheme := ctx.Get(constants.ContextSchema).(*runtime.Scheme)
	redisClient := ctx.Get(utils.ContextRedis).(rc.RedisClientInterface)
	compound := AdapterCompound{}
	compound.ServiceName = ServiceName
	compound.CalcDeployType = func(ctx core.ExecutionContext) (deployType core.MicroServiceDeployType, err error) {
		return core.CleanDeploy, nil
	}

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Adapter Service",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			kubeClient := ctx.Get(constants.ContextClient).(client.Client)
			template := Service(cr)

			core.DeleteRuntimeObject(kubeClient, &corev1.Service{
				ObjectMeta: template.ObjectMeta,
			})

			err := utils.CreateRuntimeObjectContextWrapper(ctx, template, template.ObjectMeta)
			core.PanicError(err, log.Error, "Adapter service creation failed")

			return nil
		},
	})

	compound.AddStep(&utils.SimpleCtxExecutable{
		StepName: "Adapter Server",
		ExecuteFunc: func(ctx core.ExecutionContext, cr *netcrackerv1.DbaasRedisAdapter, log *zap.Logger) error {
			return RunDBaaSServer(spec, redisClient, kubeClient, runtimeScheme, logger, request.Namespace, true)
		},
	})

	return &compound
}

func Service(cr *netcrackerv1.DbaasRedisAdapter) *corev1.Service {
	tlsEnabled := utils.IsTLSEnableForDBAAS(cr.Spec.Dbaas.Aggregator.DbaasAggregatorRegistrationAddress, cr.Spec.Redis.TLS.TLS.Enabled)
	port := utils.GetHTTPPort(tlsEnabled)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: cr.Namespace,
			Labels:    map[string]string{"app": ServiceName},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "web",
					Port: port,
					TargetPort: intstr.IntOrString{
						IntVal: port,
					},
				},
			},
			Selector: map[string]string{"name": OperatorName},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}
