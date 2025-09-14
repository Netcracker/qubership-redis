package adapter

import (
	"context"
	"fmt"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dbaas"
	fiber2 "github.com/Netcracker/qubership-dbaas-adapter-core/pkg/impl/fiber"
	coreService "github.com/Netcracker/qubership-dbaas-adapter-core/pkg/service"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	nosqlFiber "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/fiber"
	v2 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/utils"
	mCore "github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/core"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/redis"
	service "github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RunDBaaSServer(spec *v2.DbaasRedisAdapter, redisClient redis.RedisClientInterface, kubeClient client.Client, runtimeScheme *runtime.Scheme,
	log *zap.Logger, namespace string, forceShutdown bool) error {

	tlsEnabled := utils.IsTLSEnableForDBAAS(spec.Spec.Dbaas.Aggregator.DbaasAggregatorRegistrationAddress, spec.Spec.Redis.TLS.TLS.Enabled)
	coreInstance := *nosqlFiber.GetFiberService()

	appName := "redis"
	apiVersion := "v1"

	appPath := "/" + appName

	adapterCredentialsSecret, secretErr := core.ReadSecret(kubeClient, spec.Spec.Dbaas.Adapter.SecretName, namespace)
	core.PanicError(secretErr, log.Error, fmt.Sprintf("Failed reading dbaas adapter secret %s", spec.Spec.Dbaas.Adapter.SecretName))
	apiPass := string(adapterCredentialsSecret.Data[constants.Password])

	aggregatorCredentialsSecret, secretAgErr := core.ReadSecret(kubeClient, spec.Spec.Dbaas.Aggregator.SecretName, namespace)
	core.PanicError(secretAgErr, log.Error, fmt.Sprintf("Failed reading dbaas aggregator secret %s", spec.Spec.Dbaas.Aggregator.SecretName))
	regPass := string(aggregatorCredentialsSecret.Data[constants.Password])

	supports := dao.SupportsBase{
		Users:             false,
		Settings:          false,
		DescribeDatabases: true,
		AdditionalKeys: dao.Supports{
			"backupRestore": false,
		},
	}

	basicRegistrationAuth := dao.BasicAuth{
		Username: spec.Spec.Dbaas.Aggregator.Username,
		Password: regPass,
	}

	dbaasClient, err := dbaas.NewDbaasClient(spec.Spec.Dbaas.Aggregator.DbaasAggregatorRegistrationAddress, &basicRegistrationAuth, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get DBaaS aggregator version, err %v. Setting default API version", err))
	}

	version, _ := dbaasClient.GetVersion() // if err != nil it will fail in the condition above
	if version == "v3" {
		apiVersion = "v2"
	} else {
		apiVersion = "v1"
	}

	adminService := PrepareAdminService(spec, redisClient, kubeClient, runtimeScheme, log, namespace, apiVersion)

	port := utils.GetHTTPPort(spec.Spec.Redis.TLS.TLS.Enabled)
	admService := coreService.NewCoreAdministrationService(
		namespace,
		int(port),
		adminService,
		log,
		false, //we don't need static roles in vault
		nil,
		"",
	)

	app := func(app *fiber.App, ctx context.Context) error {
		fiber2.BuildFiberDBaaSAdapterHandlers(
			app,
			spec.Spec.Dbaas.Adapter.Username,
			apiPass,
			appPath,
			admService,
			coreService.NewPhysicalRegistrationService(
				appName,
				log,
				spec.Spec.Dbaas.Aggregator.PhysicalDatabaseIdentifier,
				fmt.Sprintf("%s://dbaas-redis-adapter.%v:%d", utils.GetHTTPProtocol(tlsEnabled), namespace, utils.GetHTTPPort(tlsEnabled)),
				dao.BasicAuth{
					Username: spec.Spec.Dbaas.Adapter.Username,
					Password: apiPass,
				},
				spec.Spec.Aggregator.PhysicalDatabaseLabels,
				dbaasClient,
				150000,
				60000,
				5000,
				admService,
				ctx,
			),
			nil,
			supports.ToMap(),
			log,
			false, "")
		return nil
	}

	if tlsEnabled {
		return coreInstance.CreateTLS(int(utils.GetHTTPPort(tlsEnabled)),
			fmt.Sprintf("%s/%s", mCore.CPath, spec.Spec.Redis.TLS.TLS.SignedCRTFileName),
			fmt.Sprintf("%s/%s", mCore.CPath, spec.Spec.Redis.TLS.TLS.PrivateKeyFileName),
			spec.Spec.Redis.TLS.TLS.Enabled, app, forceShutdown)
	} else {
		return coreInstance.Create(int(utils.GetHTTPPort(tlsEnabled)), app, forceShutdown)

	}

}

func PrepareAdminService(spec *v2.DbaasRedisAdapter, redisClient redis.RedisClientInterface, kubeClient client.Client, runtimeScheme *runtime.Scheme,
	log *zap.Logger, namespace string, apiVersion string) *service.AdministrationService {
	redisSpec := spec.Spec.Redis
	redisPort := 6379

	var tolerations []v12.Toleration
	if spec.Spec.Policies != nil {
		tolerations = spec.Spec.Policies.Tolerations
	}

	return service.NewAdministrationService(
		redisClient,
		spec.Spec.Adapter.SupportedFeatures,
		apiVersion,
		log.Named("DBaaS Adapter"),
		kubeClient,
		runtimeScheme,
		namespace,
		redisPort,
		*redisSpec.Resources,
		redisSpec.DockerImage,
		redisSpec.Args,
		redisSpec.Label,
		"redis",
		spec.Spec.Adapter.CreateDBTimeout,
		redisSpec.NodeLabels,
		*spec.Spec.PodSecurityContext,
		spec.Spec.ServiceAccountName,
		tolerations,
		spec.Spec.ImagePullPolicy,
		spec.Spec.Redis.TLS,
		spec.Spec.Redis.PriorityClassName,
		spec.Spec.PartOf, spec.Spec.ManagedBy,
	)
}
