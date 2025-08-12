package main

import (
	"context"
	"strconv"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v13 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func main() {
	logger := core.GetLogger(getEnvAsBool("DEBUG_LOG", true))
	logger.Info("Preparing initial configuration...")

	namespace := getEnv("NAMESPACE", "")
	redisInstancesLabelSelector := getEnv("REDIS_INSTANCES_LABEL_SELECTOR", "redis-dbaas-adapter")
	telegrafConfConfigMapName := getEnv("TELEGRAF_CONF_CONFIGMAP", "")
	//cmdMonitoringCommand := getEnv("CMD_MONITORING_COMMAND", "/vault/vault-env /sbin/tini -- telegraf")
	//cmdMonitoringCommand := getEnv("CMD_MONITORING_COMMAND", "/sbin/tini -- telegraf")
	cmdMonitoringCommand := getEnv("CMD_MONITORING_COMMAND", "env && ls && echo")
	redisPort := getEnv("REDIS_PORT", "6379")
	if namespace == "" || telegrafConfConfigMapName == "" {
		panic("NAMESPACE or TELEGRAF_CONF_CONFIGMAP variable not set")
	}

	clientSet, client := GetClient(logger)

	telegrafConfConfigMap := v12.ConfigMap{}
	getTelegrafConfigMapErr := client.Get(context.TODO(), types.NamespacedName{Name: telegrafConfConfigMapName, Namespace: namespace}, &telegrafConfConfigMap)
	core.PanicError(getTelegrafConfigMapErr, logger.Error, "Telegraf configuration map not found: "+telegrafConfConfigMapName)

	logger.Info("Start watching...")

	monitoringService := NewTelegrafProcessService(cmdMonitoringCommand, client, logger, namespace, &telegrafConfConfigMap)

	//Watch redis DBs list events and update telegraf configuration
	stopper := make(chan struct{})
	defer close(stopper)
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet,
		0,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(ls *v1.ListOptions) { ls.LabelSelector = redisInstancesLabelSelector }))
	informer := factory.Apps().V1().Deployments().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logger.Info("Redis DB is added. Starting monitoring...")
			instance := obj.(*v13.Deployment)
			redisEnv := instance.Spec.Template.Spec.Containers[0].Env
			var caFile string
			var tlsEnabled bool
			for _, env := range redisEnv {
				if env.Name == "REDIS_PORT" {
					redisPort = env.Value
				}
				if env.Name == "TLS_ENABLED" {
					tlsEnabled, _ = strconv.ParseBool(env.Value)
				}
				if env.Name == "TLS_ROOTCERT" {
					caFile = env.Value
				}
			}
			startErr := monitoringService.Start(instance.Name, redisPort, tlsEnabled, caFile)
			HandleError(startErr, logger.Error, "Failed to start monitoring process", true)
		},
		//UpdateFunc: func(oldObj interface{}, newObj interface{}) {
		//	logger.Info("Redis DB is updated. Refreshing monitoring...")
		//	instance := newObj.(*v13.Deployment)
		//
		//	startErr := monitoringService.Start(instance.Name)
		//	HandleError(startErr, logger.Error, "Failed to start monitoring process", true)
		//},
		DeleteFunc: func(obj interface{}) {
			logger.Info("Redis DB is deleted. Stopping monitoring...")
			instance := obj.(*v13.Deployment)

			stopErr := monitoringService.Stop(instance.Name)
			HandleError(stopErr, logger.Error, "Failed to stop monitoring process", true)
		},
	})
	informer.Run(stopper)
}
