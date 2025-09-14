package main

import (
	"context"
	"testing"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dbaas"
	coreTest "github.com/Netcracker/qubership-dbaas-adapter-core/testing"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v2 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	impl "github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl"
	adapter "github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/adapter"
	mUtils "github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl/utils"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/redis/mocks"

	// "github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type TestUtilsImpl struct {
	core.DefaultKubernetesHelperImpl
}

func (r *TestUtilsImpl) WaitForPVCBound(pvcName string, namespace string, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitForDeploymentReady(deployName string, namespace string, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitForPodsReady(labelSelectors map[string]string, namespace string, numberOfPods int, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitForPodsCompleted(labelSelectors map[string]string, namespace string, numberOfPods int, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitPodsCountByLabel(labelSelectors map[string]string, namespace string, numberOfPods int, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) GetPodLogs(kubeConfig *rest.Config, podName string, namespace string, containerName string, tailLines *int64, previous bool) (string, error) {

	return "", nil
}

func (r *TestUtilsImpl) WaitForTestsReady(deployName string, namespace string, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) ListPods(namespace string, labelSelectors map[string]string) (*v1core.PodList, error) {
	return &v1core.PodList{
		Items: []v1core.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: namespace,
					Labels:    labelSelectors,
				},
				Spec: v1core.PodSpec{
					Containers: []v1core.Container{
						{
							Name: "pod",
						},
					},
				},
			},
		},
	}, nil
}

func generateSecrets(namespace string, secretName string, user string, pass string) *v1core.Secret {
	return &v1core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{"username": []byte(user), "password": []byte(pass)},
	}
}

func generateDefaultConfigMap(namespace string) *v1core.ConfigMap {
	return &v1core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis-default-conf",
			Namespace: namespace,
		},
		Data: map[string]string{"config": `bind: "0.0.0.0"
protected-mode: "yes"
port: "6379"
tcp-backlog: "511"
timeout: "0"
tcp-keepalive: "300"
daemonize: "no"
supervised: "no"
loglevel: "notice"
logfile: "\"\""
databases: "16"
stop-writes-on-bgsave-error: "no"
rdbcompression: "yes"
rdbchecksum: "no"
dbfilename: "dump.rdb"
dir: "/var/lib/redis/data"
appendonly: "no"
appendfilename: "\"appendonly.aof\""
appendfsync: "no"
no-appendfsync-on-rewrite: "no"`},
	}
}

type CaseStruct struct {
	name                          string
	nameSpace                     string
	executor                      core.Executor
	builder                       core.ExecutableBuilder
	ctx                           core.ExecutionContext
	ctxToReplaceAfterServiceBuilt map[string]interface{}
	RunTestFunc                   func() error
	ReadResultFunc                func(t *testing.T, err error)
}

func GenerateDefaultRedisSpec() *v2.DbaasRedisAdapter {
	GiQuantity, _ := resource.ParseQuantity("5Gi")
	var fsGroup int64 = 999
	var tolerationSeconds int64 = 20

	rr := &v1core.ResourceRequirements{
		Limits: v1core.ResourceList{
			v1core.ResourceMemory: GiQuantity,
		},
		Requests: nil,
	}

	return &v2.DbaasRedisAdapter{
		Spec: v2.DbaasRedisAdapterSpec{
			ServiceAccountName: "redis-operator",
			Dbaas: v2.Dbaas{
				Install: true,
				Adapter: &v2.DbaasAdapter{
					Username:   "dbaas-adapter-username",
					SecretName: "dbaas-adapter-credentials",
					ApiVersion: "v2",
				},
				Aggregator: &v2.DbaasAggregator{
					Username:                           "dbaas-aggregator-username",
					SecretName:                         "dbaas-aggregator-credentials",
					Address:                            "dbaas-address",
					PhysicalDatabaseIdentifier:         "ind",
					DbaasAggregatorRegistrationAddress: "aggregator-registration-address",
				},
			},
			Redis: v2.Redis{
				DockerImage: "image",
				Args:        []string{"/etc/redis/redis.conf", "--requirepass $(REDIS_PASSWORD)"},
				Parameters:  v2.Parameters{},
				Resources:   rr,
				SecretName:  "redisdb-credentials",
			},
			Monitoring: v2.Monitoring{
				DockerImage: "image",
				Install:     true,
				Resources:   rr,
				InfluxDB: &v2.InfluxSettings{
					Host:            "influxhost",
					Database:        "influxdb",
					RetentionPolicy: "influxretentionpolicy",
					User:            "influxuser",
					SecretName:      "redis-monitoring-agent-client-secret",
				},
			},
			RobotTests: v2.RobotTests{
				DockerImage: "image",
				Install:     true,
				Resources:   rr,
			},
			WaitTimeout: 100,
			PodSecurityContext: &v1core.PodSecurityContext{
				FSGroup: &fsGroup,
			},
			Policies: &v2.Policies{
				Tolerations: []v1core.Toleration{
					{
						Key:               "key1",
						Value:             "value1",
						Operator:          v1core.TolerationOpEqual,
						Effect:            v1core.TaintEffectNoSchedule,
						TolerationSeconds: &tolerationSeconds,
					},
					{
						Key:               "key2",
						Value:             "value2",
						Operator:          v1core.TolerationOpEqual,
						Effect:            v1core.TaintEffectNoExecute,
						TolerationSeconds: &tolerationSeconds,
					},
				},
			},
		},
	}
}

func GenerateDefaultRedisTestCase(t *testing.T,
	testName string,
	redisServiceSpec *v2.DbaasRedisAdapter,
	nameSpace string,
	nameSpaceRequestName string,
	runtimeObjects ...runtime.Object,
) *CaseStruct {

	utilsHelp := &TestUtilsImpl{}
	utilsHelp.ForceKey = true
	// Because there is empty runtime Scheme
	utilsHelp.OwnerKey = false
	fakeClient := fake.NewFakeClient(runtimeObjects...)
	fakeRedis := mocks.NewRedisClientInterface(t)

	utilsHelp.Client = fakeClient

	caseStruct := &CaseStruct{
		name:      testName,
		nameSpace: nameSpace,
		executor:  core.DefaultExecutor(),
		builder:   &impl.RedisServiceBuilder{},
		ctx: core.GetExecutionContext(map[string]interface{}{
			constants.ContextSpec:   redisServiceSpec,
			constants.ContextSchema: &runtime.Scheme{},
			constants.ContextRequest: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: nameSpace,
					Name:      nameSpaceRequestName,
				},
			},
			constants.ContextClient:                fakeClient,
			constants.ContextKubeClient:            &rest.Config{},
			constants.ContextLogger:                core.GetLogger(true),
			"contextResourceOwner":                 redisServiceSpec, //todo hardcode replace
			constants.ContextServiceDeploymentInfo: map[string]string{},
			mUtils.ContextRedis:                    fakeRedis,
		}),
		ctxToReplaceAfterServiceBuilt: map[string]interface{}{
			constants.KubernetesHelperImpl: utilsHelp,
		},
		ReadResultFunc: func(t *testing.T, err error) {
			if err != nil {
				// Some error happened
				t.Error(err)
			}
		},
	}

	caseStruct.RunTestFunc = func() error {
		return caseStruct.executor.Execute(caseStruct.ctx)
	}

	return caseStruct
}

func GetRuntimeObjects(nameSpace string) []runtime.Object {
	adapterSecret := generateSecrets(nameSpace, "dbaas-adapter-credentials", "admin", "admin")
	aggregatorSecret := generateSecrets(nameSpace, "dbaas-aggregator-credentials", "admin", "admin")

	runtimeObj := []runtime.Object{}
	runtimeObj = append(runtimeObj, adapterSecret, aggregatorSecret, generateDefaultConfigMap(nameSpace))

	return runtimeObj
}

func GenerateDefaultRedisWrapper(t *testing.T,
	testName string,
) *CaseStruct {
	nameSpace := "redis-namespace"
	nameSpaceRequestName := "redis-name"

	redisSpec := GenerateDefaultRedisSpec()

	return GenerateDefaultRedisTestCase(t,
		testName,
		redisSpec,
		nameSpace,
		nameSpaceRequestName,
		GetRuntimeObjects(nameSpace)...,
	)
}

func TestExecutionCheck(t *testing.T) {
	aggregatorServer := coreTest.GetTestHttpAggregatorServer("dbaas-aggregator-username",
		"admin", "redis", "redis", false)
	defer aggregatorServer.Close()
	aggAddress := aggregatorServer.URL
	testFuncs := []func() CaseStruct{
		func() CaseStruct {
			cs := GenerateDefaultRedisWrapper(t,
				"Dbaas Adapter Schema")
			spec := cs.ctx.Get(constants.ContextSpec).(*v2.DbaasRedisAdapter)

			spec.Spec.Dbaas.Aggregator.DbaasAggregatorRegistrationAddress = aggAddress
			spec.Spec.Dbaas.Adapter.ApiVersion = "v1"
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			return *cs
		},
		func() CaseStruct {
			cs := GenerateDefaultRedisWrapper(t,
				"Redis Only Schema")
			spec := cs.ctx.Get(constants.ContextSpec).(*v2.DbaasRedisAdapter)
			spec.Spec.Dbaas.Install = false
			cs.ctx.Set(constants.ContextSpec, spec)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			return *cs
		},
	}

	tests := []CaseStruct{}
	for _, tf := range testFuncs {
		tests = append(tests, tf())
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, elem := range tt.ctxToReplaceAfterServiceBuilt {
				tt.ctx.Set(key, elem)
			}
			err := tt.RunTestFunc()
			tt.ReadResultFunc(t, err)
		})
	}
}

func Test_FullFeaturedConfigV2(t *testing.T) {
	logger := core.GetLogger(true)
	spec := GenerateDefaultRedisSpec()
	nameSpace := "redis-namespace"
	apiVersion := "v2"
	redisClient := mocks.NewRedisClientInterface(t)
	redisClient.On("InitRedisClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redisClient, nil)
	redisClient.On("Ping").Return("PONG", nil)
	redisClient.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	redisClient.On("Addr").Return("")
	redisClient.On("Close").Return(nil)
	dbAdmin := adapter.PrepareAdminService(spec, redisClient, fake.NewFakeClient(GetRuntimeObjects(nameSpace)...), &runtime.Scheme{}, logger, nil, nameSpace, apiVersion)

	appCredentials := coreTest.AppCredentials{
		AppName:           coreTest.Simplstr(),
		AdapterApiUser:    coreTest.Simplstr(),
		AdapterApiPass:    coreTest.Simplstr(),
		BackupApiUser:     coreTest.Simplstr(),
		BackupApiPass:     coreTest.Simplstr(),
		AggregatorApiUser: coreTest.Simplstr(),
		AggregatorApiPass: coreTest.Simplstr(),
	}

	aggregatorServer := coreTest.GetTestHttpAggregatorServer(appCredentials.AggregatorApiUser,
		appCredentials.AggregatorApiPass, appCredentials.AppName, appCredentials.AppName, false)
	defer aggregatorServer.Close()
	aggAddress := aggregatorServer.URL

	dbaasClient, err := dbaas.NewDbaasClient(aggAddress, &dao.BasicAuth{appCredentials.AggregatorApiUser, appCredentials.AggregatorApiPass}, nil)
	if err != nil {
		assert.Fail(t, "Failed to create Dbaas Client", err)
	}

	cancelFunc, testApp, appErr, appCredentials := coreTest.PrepateTestApp(dbaasClient, logger, dbAdmin, appCredentials, aggAddress)
	if appErr != nil {
		logger.Error("Error during app initialization")
		if cancelFunc != nil {
			cancelFunc()
		}
		assert.Fail(t, "Failed initializing app", appErr)
	}

	coreTest.UseFullFeaturedConfig(logger, t, testApp, string(dbAdmin.GetVersion()),
		appCredentials.AppName, appCredentials.AdapterApiUser, appCredentials.AdapterApiPass, false, false,
		appCredentials.BackupApiUser, appCredentials.BackupApiPass)
}
