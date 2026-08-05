package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/adwpc/pagent"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	redisTelegrafConfigMapInitKey    = "telegraf-init"
	redisTelegrafConfigMapOutputsKey = "telegraf-outputs"
	redisTelegrafConfigMapInputsKey  = "telegraf-inputs"

	envCredentialsPrefix  = "TELEGRAF_REDIS_PREFIX_"
	envDBName             = envCredentialsPrefix + "DBNAME"
	envDBService          = envCredentialsPrefix + "DBSERVICE"
	envDBPort             = envCredentialsPrefix + "DBPORT"
	envDBPass             = envCredentialsPrefix + "DBPASS"

	Password = "password"

	duration = time.Duration(5)
)

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func HandleError(err error, log func(msg string, fields ...zap.Field), message string, print ...bool) {
	if err != nil {
		if len(print) > 0 && print[0] {
			log(fmt.Sprintf("%s\n%s", message, err.Error()))
		} else {
			log(message)
		}
	}
}

func PanicError(err error, log func(msg string, fields ...zap.Field), message string) {
	HandleError(err, log, message, false)
	if err != nil {
		panic(&core.ExecutionError{Msg: fmt.Sprintf("%s\n%s", message, err.Error())})
	}
}

func GetClient(logger *zap.Logger) (kubernetes.Interface, client.Client) {
	restConfig := config.GetConfigOrDie()

	cl, clientErr := client.New(restConfig, client.Options{})
	core.PanicError(clientErr, logger.Error, "Failed creating kubernetes client")

	clientset, clientSetErr := kubernetes.NewForConfig(restConfig)
	core.PanicError(clientSetErr, logger.Error, "Failed creating clientset")

	return clientset, cl
}

type ProcessMaster struct {
	pagent.Master
	logger *zap.Logger
}

func (a *ProcessMaster) running(id, str string) error {
	a.logger.Info("[Process Master][" + id + "] " + str)
	return nil
}

func (a *ProcessMaster) finish(id string, err error) error {
	a.logger.Info("[Process Master][" + id + "][FINISHED]")
	HandleError(err, a.logger.Error, "Process finished with error", true)
	return err
}

func NewProcessMaster(logger *zap.Logger) *ProcessMaster {
	return &ProcessMaster{
		logger: logger,
	}
}

func NewTelegrafProcessService(cmd string, client client.Client, logger *zap.Logger, namespace string, rtcm *v12.ConfigMap) *TelegrafMonitoringService {
	ms := &TelegrafMonitoringService{
		TelegrafProcessArgs:    cmd,
		client:                 client,
		logger:                 logger,
		processMaster:          NewProcessMaster(logger),
		workerName:             "monitoring",
		confFullPath:           "/tmp/telegraf.conf",
		namespace:              namespace,
		redisConf:              map[string]RedisConf{},
		redisTelegrafConfigMap: rtcm,
		mutex:                  &sync.Mutex{},
	}
	//Refreshable timer that runs once and sleeps. Restarts on timer.Reset event and runs function with delay
	//Delay is required to prevent multiple restarts if several redis instances added at once (like it is happening on start)
	ms.timer = time.AfterFunc(time.Second*duration, func() {
		core.HandleError(ms.Refresh(), logger.Error, "Failed refreshing monitoring")
	})

	return ms
}

type RedisConf struct {
	port       string
	tlsEnabled bool
	caFile     string
}

type TelegrafMonitoringService struct {
	TelegrafProcessArgs    string
	client                 client.Client
	logger                 *zap.Logger
	processMaster          *ProcessMaster
	workerName             string
	confFullPath           string
	namespace              string
	timer                  *time.Timer
	redisConf              map[string]RedisConf
	redisTelegrafConfigMap *v12.ConfigMap
	mutex                  *sync.Mutex
}

func (r *TelegrafMonitoringService) startProcess() error {
	cmd := "sh"
	resultStr := r.TelegrafProcessArgs + " -config " + r.confFullPath
	r.logger.Info(fmt.Sprintf("Starting process: %s -c %s", cmd, resultStr))
	return r.processMaster.
		GetWorker(r.workerName).
		Start(
			cmd,
			r.processMaster.running,
			r.processMaster.finish,
			"-c", resultStr)
}

func (r *TelegrafMonitoringService) stopProcess() error {
	return r.processMaster.
		DelWorker(r.workerName)
}

func (r *TelegrafMonitoringService) updateTelegrafConfAndEnvironmentVariables(redisTelegrafConfConfigMap *v12.ConfigMap, redisConfMap map[string]RedisConf) error {
	//telegraf-init + telegraf-outputs
	resultRaw := []string{
		redisTelegrafConfConfigMap.Data[redisTelegrafConfigMapInitKey],
		redisTelegrafConfConfigMap.Data[redisTelegrafConfigMapOutputsKey],
	}

	//generate environment variable pass with prefix
	//replace variables with prefix in inputs with the variables from environment
	redisInstancesCount := len(redisConfMap)

	if redisInstancesCount < 1 {
		return &core.ExecutionError{Msg: "Redis instances not found"}
	}

	r.logger.Debug(fmt.Sprintf("Found %v redis instances", len(redisConfMap)))

	for redisInstance, redisConf := range redisConfMap {

		passSecret := v12.Secret{}
		secretGetErr := r.client.Get(context.TODO(), types.NamespacedName{Name: redisInstance + "-credentials", Namespace: r.namespace}, &passSecret)
		if secretGetErr != nil {
			r.logger.Error("Failed reading " + redisInstance + "-credentials" + " secret")
			return secretGetErr
		}
		dbPassEnvName := envCredentialsPrefix + strings.ReplaceAll(strings.ToUpper(redisInstance), "-", "_")
		setEnvErr := os.Setenv(dbPassEnvName, string(passSecret.Data[Password]))
		if setEnvErr != nil {
			r.logger.Error("Failed setting " + dbPassEnvName + " environment variable")
			return setEnvErr
		}

		r.logger.Debug("Variable " + dbPassEnvName + " is added to envrionment")

		variablesReplaceMap := map[string]string{
			envDBName:             redisInstance,
			envDBService:          fmt.Sprintf("%s.%s.svc", redisInstance, r.namespace),
			envDBPort:             redisConf.port,
			envDBPass:             "$" + dbPassEnvName + "",
		}

		inputs := redisTelegrafConfConfigMap.Data[redisTelegrafConfigMapInputsKey]
		for key, value := range variablesReplaceMap {
			inputs = strings.ReplaceAll(inputs, key, value)
		}

		resultRaw = append(resultRaw, inputs)
	}

	result := strings.Join(resultRaw, "\n")

	r.logger.Debug(fmt.Sprintln("Telegraf configuration result: \n" + result))

	//save config to file
	f, saveCfgErr := os.Create(r.confFullPath)
	if saveCfgErr != nil {
		r.logger.Error("Unable to create a file " + r.confFullPath)
		return saveCfgErr
	}
	defer f.Close()

	truncateErr := f.Truncate(0)
	if truncateErr != nil {
		r.logger.Error("Unable to flush a file " + r.confFullPath)
		return truncateErr
	}

	_, writeErr := f.WriteString(result)
	if writeErr != nil {
		r.logger.Error("Unable to write to a file " + r.confFullPath)
		return writeErr
	}

	return nil
}

func (r *TelegrafMonitoringService) Start(redisInstance, redisPort string, tlsEnabled bool, caFile string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.timer.Reset(time.Second * duration)

	r.redisConf[redisInstance] = RedisConf{
		port:       redisPort,
		caFile:     caFile,
		tlsEnabled: tlsEnabled,
	}

	return nil
}

func (r *TelegrafMonitoringService) Stop(redisInstance string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.timer.Reset(time.Second * duration)
	delete(r.redisConf, redisInstance)

	return nil
}

//Starts new or restarts existed process
func (r *TelegrafMonitoringService) Refresh() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.logger.Info("Stopping monitoring process...")
	core.HandleError(r.stopProcess(), r.logger.Warn, "Failed stopping monitoring process")
	if len(r.redisConf) > 0 {
		r.logger.Debug("Building new telegraf.conf....")
		updConfErr := r.updateTelegrafConfAndEnvironmentVariables(r.redisTelegrafConfigMap, r.redisConf)
		if updConfErr != nil {
			return updConfErr
		}
	}
	r.logger.Info("Starting monitoring process....")
	return r.startProcess()
}
