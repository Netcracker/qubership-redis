package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"regexp"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"

	coreService "github.com/Netcracker/qubership-dbaas-adapter-core/pkg/service"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	mTypes "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	coreUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/vault"
	v2 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/common"
	customEntity "github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/entity"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/helper"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/redis"
	"github.com/Netcracker/qubership-redis/redis-operator/dbaas/pkg/templates"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AdministrationService struct {
	redisClient                       redis.RedisClientInterface
	supportedFeatures                 map[string]bool
	apiVersion                        string
	logger                            *zap.Logger
	kubeClient                        client.Client
	runtimeScheme                     *runtime.Scheme
	namespace                         string
	redisServicePort                  int
	redisResources                    v1.ResourceRequirements
	redisImage                        string
	redisArgs                         []string
	redisLabel                        string
	defaultRedisPassword              string
	defaultRedisDbStartWait           int
	vaulterHelper                     vault.VaultHelper
	nodeSelector                      map[string]string
	securityContext                   v1.PodSecurityContext
	serviceAccountName                string
	tolerations                       []v1.Toleration
	redisImagePullPolicy              v1.PullPolicy
	vaultRegistration                 mTypes.VaultRegistration
	tls                               v2.TLS
	priorityClassName                 string
	artDescVersion, partOf, managedBy string
}

var _ coreService.DbAdministration = &AdministrationService{}

var (
	credsSuffix        = "-credentials"
	regexpExpression   = "^[a-z][-a-z0-9]*[a-z0-9]?$"
	nameRegexp, _      = regexp.Compile(regexpExpression)
	redisPasswordConst = "REDIS_PASSWORD"
	passCharSet        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789"
	vaultPolicy                      = fmt.Sprintf("length = 16\nrule \"charset\" {\n  charset = \"%s\"\n}\n", passCharSet)
	RedisDefaultConfigMapName string = "redis-default-conf"
	// 55 characters because the max name length is 63 chars in k8s, but we use db name + various suffixes for creating other k8s units
	dbNameLenghtLimit = 55
)

func NewAdministrationService(
	redisClient redis.RedisClientInterface,
	supportedFeatures map[string]bool,
	apiVersion string,
	logger *zap.Logger,
	kubeClint client.Client,
	runtimeScheme *runtime.Scheme,
	namespace string,
	redisServicePort int,
	redisResources v1.ResourceRequirements,
	redisImage string,
	redisArgs []string,
	redisLabel string,
	defaultRedisPassword string,
	defaultRedisDbStartWait int,
	vaulterHelper vault.VaultHelper,
	vaultRegistration mTypes.VaultRegistration,
	nodeSelector map[string]string,
	securityContext v1.PodSecurityContext,
	serviceAccountName string,
	tolerations []v1.Toleration,
	redisImagePullPolicy v1.PullPolicy,
	redisTls v2.TLS,
	priorityClassName string, partOf, managedBy string) *AdministrationService {

	return &AdministrationService{
		redisClient:             redisClient,
		supportedFeatures:       supportedFeatures,
		apiVersion:              apiVersion,
		logger:                  logger,
		kubeClient:              kubeClint,
		runtimeScheme:           runtimeScheme,
		namespace:               namespace,
		redisServicePort:        redisServicePort,
		redisResources:          redisResources,
		redisImage:              redisImage,
		redisArgs:               redisArgs,
		redisLabel:              redisLabel,
		defaultRedisPassword:    defaultRedisPassword,
		defaultRedisDbStartWait: defaultRedisDbStartWait,
		vaulterHelper:           vaulterHelper,
		vaultRegistration:       vaultRegistration,
		nodeSelector:            nodeSelector,
		securityContext:         securityContext,
		serviceAccountName:      serviceAccountName,
		tolerations:             tolerations,
		redisImagePullPolicy:    redisImagePullPolicy,
		tls:                     redisTls,
		priorityClassName:       priorityClassName,
		partOf:                  partOf,
		managedBy:               managedBy,
	}
}

// Ugly way to have a list of resources to line API contract
// And do not include extra into service implementation clientsets, RESTMappers etc...
type DBResourceMapping struct {
	name   string
	object client.Object
}

func (adminService *AdministrationService) PreStart() {}

func (adminService *AdministrationService) GetDBPrefix() string {
	return "dbaas"
}

func (adminService *AdministrationService) GetDBPrefixDelimiter() string {
	return "-"
}

func (adminService *AdministrationService) GetFeatures() map[string]bool {
	return adminService.supportedFeatures
}

func (adminService *AdministrationService) GetSupportedRoles() []string {
	return []string{"admin"}
}

func (adminService *AdministrationService) GetVersion() dao.ApiVersion {
	return dao.ApiVersion(adminService.apiVersion)
}

func (adminService *AdministrationService) MigrateToVault(ctx context.Context, dbName, userName string) error {
	// No additional actions needed
	return nil
}

func (adminService *AdministrationService) getResourcesMapping(serviceName string) map[string]DBResourceMapping {
	om := metav1.ObjectMeta{
		Name:      serviceName,
		Namespace: adminService.namespace,
	}
	secretOm := om
	if !strings.HasSuffix(serviceName, credsSuffix) {
		secretOm.Name = credsName(secretOm.Name)
	}
	return map[string]DBResourceMapping{
		"Secret": {
			name: secretOm.Name,
			object: &v1.Secret{
				ObjectMeta: secretOm,
			},
		},
		"ConfigMap": {
			name: om.Name,
			object: &v1.ConfigMap{
				ObjectMeta: om,
			},
		},
		"Deployment": {
			name: om.Name,
			object: &v12.Deployment{
				ObjectMeta: om,
			},
		},
		"Service": {
			name: om.Name,
			object: &v1.Service{
				ObjectMeta: om,
			},
		},
	}
}

// Precompiled list of resources to reduce amount of get operations
func (adminService *AdministrationService) getDBResources(serviceName string) []dao.DbResource {
	resMapping := adminService.getResourcesMapping(serviceName)
	var result []dao.DbResource
	for kind, val := range resMapping {
		result = append(result, dao.DbResource{Kind: kind, Name: val.name})
	}
	return result
}

func (adminService *AdministrationService) listRedisDeployments(listOptions []client.ListOption) (v12.DeploymentList, error) {
	//check if deployment already exists
	redisDL := v12.DeploymentList{}
	dErr := adminService.kubeClient.List(context.TODO(), &redisDL, listOptions...)
	return redisDL, dErr
}

func (adminService *AdministrationService) CreateRoles(ctx context.Context, roles []dao.AdditionalRole) ([]dao.Success, *dao.Failure) {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	core.PanicError(fmt.Errorf("Not Implemeted"), logger.Error, "This function must never be called for Redis")
	return nil, nil
}

func (adminService *AdministrationService) CreateUser(ctx context.Context, userName string, requestOnCreateUser dao.UserCreateRequest) (*dao.CreatedUser, error) {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	core.PanicError(fmt.Errorf("Not Implemeted"), logger.Error, "This function must never be called for Redis")
	return nil, nil
}

func (adminService *AdministrationService) GetDefaultUserCreateRequest() dao.UserCreateRequest {
	core.PanicError(fmt.Errorf("Not Implemeted"), adminService.logger.Error, "This function must never be called for Redis")
	return dao.UserCreateRequest{}
}

func (adminService *AdministrationService) setMetadata(ctx context.Context, metadata map[string]interface{}, redisdb redis.RedisClientInterface) {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	metadataBytes, err := json.Marshal(metadata)
	core.PanicError(err, logger.Error, "Error during marshal metadata")
	err = redisdb.Set("dbaas.metadata", string(metadataBytes), 0)
	core.PanicError(err, logger.Error, fmt.Sprintf("Error during set metadata in %s", redisdb.Addr()))
}

func (adminService *AdministrationService) GetMetadata(ctx context.Context, serviceName string) map[string]interface{} {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	password := adminService.readRedisDBPassword(ctx, serviceName)
	redisdb := adminService.createRedisClient(ctx, fmt.Sprintf("%s.%s:%d", serviceName, adminService.namespace, adminService.redisServicePort), password, 0)
	defer redisdb.Close()
	meta, err := redisdb.Get("dbaas.metadata")
	core.PanicError(err, logger.Error, fmt.Sprintf("Failed to read metadata for DB %s", serviceName))

	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(meta), &metadata)
	core.PanicError(err, logger.Error, fmt.Sprintf("failed to unmarshal metadata for DB %s", serviceName))

	return metadata
}

func (adminService *AdministrationService) UpdateMetadata(ctx context.Context, newMetadata map[string]interface{}, serviceName string) {
	password := adminService.readRedisDBPassword(ctx, serviceName)
	redisdb := adminService.createRedisClient(ctx, fmt.Sprintf("%s.%s:%d", serviceName, adminService.namespace, adminService.redisServicePort), password, 0)
	defer redisdb.Close()
	adminService.setMetadata(ctx, newMetadata, redisdb)
}

func (adminService *AdministrationService) GetDefaultCreateRequest() dao.DbCreateRequest {
	selectorsCopy := make(map[string]interface{})
	for key, val := range adminService.nodeSelector {
		selectorsCopy[key] = val
	}
	redisSettings := map[string]interface{}{
		"redisDbResources":              *adminService.redisResources.DeepCopy(),
		"redisDbNodeSelector":           selectorsCopy,
		"redisDbWaitStartServiceSecond": adminService.defaultRedisDbStartWait,
	}
	return dao.DbCreateRequest{
		Settings: redisSettings,
	}
}

func (adminService *AdministrationService) CreateDatabase(ctx context.Context, requestOnCreateDb dao.DbCreateRequest) (string, *dao.LogicalDatabaseDescribed, error) {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	var logicalDatabaseName = requestOnCreateDb.DbName
	var err error

	if logicalDatabaseName == "" {
		logicalDatabaseName = helper.GenerateDbName()
	}

	settings, err := adminService.convertSettings(requestOnCreateDb.Settings)
	if err != nil {
		return "", nil, err
	}

	if requestOnCreateDb.NamePrefix != nil {
		if *requestOnCreateDb.NamePrefix != "" {
			// check if prefix + generated name more then the maximum length
			if len(*requestOnCreateDb.NamePrefix) > (dbNameLenghtLimit - len(logicalDatabaseName)) {
				return "", nil, customEntity.NewInvalidArgumentError(fmt.Sprintf("The combined length of the prefix [%s] and the database name [%s] can't be longer than 63 characters", *requestOnCreateDb.NamePrefix, logicalDatabaseName))
			}
			logicalDatabaseName = *requestOnCreateDb.NamePrefix + "-" + logicalDatabaseName
		}
	} else {
		if classifier, ok := requestOnCreateDb.Metadata["classifier"].(map[string]interface{}); ok {
			if namespace, ok := classifier["namespace"].(string); ok {
				if microserviceName, ok := classifier["microserviceName"].(string); ok {
					logicalDatabaseName, err = utils.PrepareDatabaseName(namespace, microserviceName, dbNameLenghtLimit)
					if err != nil {
						return "", nil, err
					}
					logicalDatabaseName = strings.ReplaceAll(logicalDatabaseName, "_", "-")
				}
			}
		} else {
			logicalDatabaseName = fmt.Sprintf("%s%s%s", adminService.GetDBPrefix(), adminService.GetDBPrefixDelimiter(), logicalDatabaseName)
		}
	}

	if len(logicalDatabaseName) > dbNameLenghtLimit {
		return "", nil, customEntity.NewInvalidArgumentError(fmt.Sprintf("The database name '%s' can't be longer than %d characters", logicalDatabaseName, dbNameLenghtLimit))
	}

	if !nameRegexp.MatchString(logicalDatabaseName) {
		return "", nil, customEntity.NewInvalidArgumentError(fmt.Sprintf("The database name '%s' must match on reqexp expression %s", logicalDatabaseName, regexpExpression))
	}

	//check if deployment already exists
	lo := []client.ListOption{
		client.InNamespace(adminService.namespace),
		client.MatchingLabelsSelector{labels.SelectorFromSet(map[string]string{constants.Name: logicalDatabaseName})},
	}

	redisDL, dErr := adminService.listRedisDeployments(lo)
	if dErr != nil && !errors.IsNotFound(dErr) {
		return "", nil, dErr
	}
	for _, db := range redisDL.Items {
		if logicalDatabaseName == db.Name {
			return "", nil, dao.NewResourceAlreadyExistsError(fmt.Sprintf("Database %s already exists", logicalDatabaseName))
		}
	}

	type objectToCreate struct {
		object client.Object
		meta   metav1.ObjectMeta
	}

	var objectsToCreate []objectToCreate

	certErr := common.UpdateCertificate(adminService.tls.Enabled, adminService.tls.ClusterIssuerName, logicalDatabaseName, adminService.namespace, adminService.kubeClient, adminService.runtimeScheme)
	core.PanicError(certErr, logger.Error, "Failed to update TLS certificate")

	// Making secret for pass
	credsSecretName := credsName(logicalDatabaseName)
	secret, plainTextPass, redisEnvCredErr := adminService.storeCredentialsAndGetEnvForRedisInstance(credsSecretName, requestOnCreateDb.Password)
	if redisEnvCredErr != nil {
		return "", nil, redisEnvCredErr
	}

	envVarForRedisInstance := coreUtils.GetSecretEnvVar(redisPasswordConst, credsSecretName, constants.Password)

	objectsToCreate = append(objectsToCreate,
    	objectToCreate{secret, secret.ObjectMeta})

	// Making ConfigMap
	redisConfig := GetRedisDefaultConfigMap(adminService.kubeClient, adminService.namespace, logger)
	adminService.setRedisDatabaseSettings(ctx, &redisConfig, &settings.RedisDbSettings)
	configString := RedisMapConfigToString(redisConfig)
	configMap := templates.GetRedisConfigTemplate(logicalDatabaseName, adminService.namespace, configString)
	objectsToCreate = append(objectsToCreate, objectToCreate{configMap, configMap.ObjectMeta})

	// The Redis Service
	redisService := templates.GetRedisServiceTemplate(
		logicalDatabaseName,
		adminService.namespace, adminService.partOf, adminService.managedBy)

	objectsToCreate = append(objectsToCreate, objectToCreate{redisService, redisService.ObjectMeta})

	envs := common.GetRedisEnvs(adminService.tls.TLS)
	envs = append(envs, envVarForRedisInstance)

	// The Redis Deployment
	redisDeployment := templates.GetRedisDeploymentTemplate(
		logicalDatabaseName,
		adminService.namespace,
		adminService.redisImage,
		adminService.redisArgs,
		envs,
		settings.RedisDbResources,
		settings.RedisDbNodeSelector,
		&adminService.securityContext,
		adminService.serviceAccountName,
		adminService.tolerations,
		adminService.redisLabel,
		adminService.redisImagePullPolicy,
		adminService.tls,
		adminService.priorityClassName,
		adminService.partOf,
		adminService.managedBy,
	)

	objectsToCreate = append(objectsToCreate, objectToCreate{redisDeployment, redisDeployment.ObjectMeta})

	if adminService.vaulterHelper != nil {
		coreUtils.VaultPodSpec(&redisDeployment.Spec.Template.Spec, nil, adminService.vaultRegistration)
		redisDeployment.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh"}
		redisDeployment.Spec.Template.Spec.Containers[0].Args = []string{"-c", strings.Join(redisDeployment.Spec.Template.Spec.Containers[0].Args, " ")}
	}

	var createAndCheckErr error

	//rollback - delete all if any object has failed to create
	defer func() {
		if createAndCheckErr != nil {
			for _, objectToCreate := range objectsToCreate {
				core.DeleteRuntimeObject(adminService.kubeClient, objectToCreate.object)
			}
		}

	}()

	for _, objectToCreate := range objectsToCreate {
		createAndCheckErr = core.CreateOrUpdateRuntimeObject(adminService.kubeClient, nil, nil, objectToCreate.object, objectToCreate.meta, true)
		if createAndCheckErr != nil {
			return "", nil, createAndCheckErr
		}
	}

	connectionProperties := createConnectionProperties(logicalDatabaseName, plainTextPass, adminService.namespace, adminService.redisServicePort)

	cp := &customEntity.ConnectionProperties{}
	mapstructure.Decode(connectionProperties[0], cp)
	createAndCheckErr = adminService.checkConnectAndSetMetadata(ctx, *cp, requestOnCreateDb)
	if createAndCheckErr != nil {
		return "", nil, createAndCheckErr
	}
	resources := adminService.getDBResources(logicalDatabaseName)

	logger.Info(fmt.Sprintf("Logical database with name %s has resources %+v", logicalDatabaseName, resources))

	return logicalDatabaseName, &dao.LogicalDatabaseDescribed{ConnectionProperties: connectionProperties, Resources: resources}, nil
}

func (adminService *AdministrationService) convertSettings(requestSettings map[string]interface{}) (*customEntity.DbCreateRequestSettings, error) {
	settings := customEntity.DbCreateRequestSettings{
		RedisDbResources: *adminService.redisResources.DeepCopy(),
	}

	jsonSettings, err := json.Marshal(requestSettings)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonSettings, &settings)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func (adminService *AdministrationService) setRedisDatabaseSettings(ctx context.Context, parameters *map[string]interface{}, settings *map[string]interface{}) {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	if *settings != nil {
		for parameter, value := range *settings {
			logger.Info(fmt.Sprintf("Redis database settings: parameter = %s, value = %+v", parameter, value))
			(*parameters)[parameter] = value
		}
	}
}

func (adminService *AdministrationService) checkConnectAndSetMetadata(ctx context.Context, connectionProperties customEntity.ConnectionProperties, requestOnCreateDb dao.DbCreateRequest) error {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	redisdb := adminService.createRedisClient(ctx, fmt.Sprintf("%s:%d", connectionProperties.Host, connectionProperties.Port), connectionProperties.Password, 0)
	defer redisdb.Close()
	timeWaitServiceSecond := adminService.defaultRedisDbStartWait
	settings, err := adminService.convertSettings(requestOnCreateDb.Settings)

	if err != nil {
		return fmt.Errorf("failed to decode request settings %v, err: %v", requestOnCreateDb.Settings, err)
	}

	if settings.RedisDbWaitStartServiceSecond > 0 {
		timeWaitServiceSecond = settings.RedisDbWaitStartServiceSecond
	}
	initialTime := timeWaitServiceSecond
	for {
		result, err := redisdb.Ping()
		if result == "PONG" {
			logger.Info(fmt.Sprintf("The service %s started", connectionProperties.Service))
			break
		}
		logger.Info(fmt.Sprintf("Wait until service %s will start and redis-client will be able to perform connect (host\"%s\"). Time left %d, result %s, error %s", connectionProperties.Service,
			redisdb.Addr(), timeWaitServiceSecond, result, err))
		if timeWaitServiceSecond != 0 {
			logger.Info(fmt.Sprint("sleep one second and then check connect again ..."))
			time.Sleep(time.Second)
			timeWaitServiceSecond--
		} else {
			return fmt.Errorf("the service %s could not start in %d second, err: %v", connectionProperties.Host, initialTime, err)
		}
	}
	adminService.setMetadata(ctx, requestOnCreateDb.Metadata, redisdb)
	logger.Info(fmt.Sprintf("Metadata %+v was set successfully", requestOnCreateDb.Metadata))

	return nil
}

func createConnectionProperties(logicalDatabaseName string, password string, namespace string, redisServicePort int) []dao.ConnectionProperties {
	cp := customEntity.ConnectionProperties{Host: fmt.Sprintf("%s.%s", logicalDatabaseName, namespace),
		Port: redisServicePort, Service: logicalDatabaseName, Password: password,
		Url: fmt.Sprintf("redis://%s.%s:%d", logicalDatabaseName, namespace, redisServicePort), Role: "admin"}
	var cpMap map[string]interface{}
	mapstructure.Decode(cp, &cpMap)
	return []dao.ConnectionProperties{cpMap}
}

func (adminService *AdministrationService) GetDatabases(ctx context.Context) []string {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	lo := []client.ListOption{
		client.InNamespace(adminService.namespace),
		client.MatchingLabelsSelector{labels.SelectorFromSet(map[string]string{adminService.redisLabel: adminService.redisLabel})},
	}

	var result []string
	redisDL, dErr := adminService.listRedisDeployments(lo)
	if dErr != nil {
		if !errors.IsNotFound(dErr) {
			core.PanicError(dErr, logger.Error, "Failed listing Redis Databases")
		} else {
			return result
		}
	}
	for _, deployment := range redisDL.Items {
		result = append(result, deployment.ObjectMeta.Name)
	}
	return result
}

func (adminService *AdministrationService) DropResources(ctx context.Context, resources []dao.DbResource) []dao.DbResource {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	var dropStatuses []dao.DbResource
	for _, resource := range resources {
		resourceKind := resource.Kind
		resourceName := resource.Name

		obj := adminService.getResourcesMapping(resourceName)[resourceKind]

		err := core.DeleteRuntimeObject(adminService.kubeClient, obj.object)

		if err != nil {
			logger.Warn(fmt.Sprintf("Error during deleting resource %s with name \"%s\", %+v", resource.Kind, resource.Name, err))
			resource.Status = dao.DELETE_FAILED
			resource.ErrorMessage = err.Error()
		} else {
			resource.Status = dao.DELETED
			logger.Info(fmt.Sprintf("The resource %s:%s was deleted successfully.", resourceKind, resourceName))
		}
		dropStatuses = append(dropStatuses, resource)
	}
	return dropStatuses
}

func (adminService *AdministrationService) readRedisDBPassword(ctx context.Context, serviceName string) string {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	passSecretName := credsName(serviceName)
	var password string

	// Read pass in secret
	secretObj := v1.Secret{}
	secretErr := adminService.kubeClient.Get(ctx, types.NamespacedName{Name: passSecretName, Namespace: adminService.namespace}, &secretObj)
	if secretErr != nil {
		core.PanicError(secretErr, logger.Error, fmt.Sprintf("Failed getting service %s password from secret %s", serviceName, passSecretName))
	}
	password = string(secretObj.Data[constants.Password])

	if adminService.vaulterHelper != nil && adminService.vaulterHelper.IsVaultURL(password) {
		isExist, secret, vaultErr := adminService.vaulterHelper.CheckSecretExists(passSecretName)
		if vaultErr != nil {
			core.PanicError(secretErr, logger.Error, fmt.Sprintf("Failed checking secret %s in vault", passSecretName))
		}
		if isExist {
			password = secret[constants.Password].(string)
		} else {
			core.PanicError(secretErr, logger.Error, fmt.Sprintf("Not found secret %s in vault for service %s", passSecretName, serviceName))
		}
	}

	if password == "" {
		core.PanicError(secretErr, logger.Error, fmt.Sprintf("The password for connect and update metadata was not found for %s", serviceName))
	}

	return password
}

func (adminService *AdministrationService) createRedisClient(ctx context.Context, address string, password string, db int) redis.RedisClientInterface {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	logger.Info(fmt.Sprintf("Create redis client with address %s", address))
	var caCert []byte
	if adminService.tls.Enabled {
		var err error
		caCert, err = ioutil.ReadFile(fmt.Sprintf("/usr/ssl/%s", adminService.tls.RootCAFileName))
		if err != nil {
			core.PanicError(err, logger.Error, "Failed to read root certificate")
		}
	}
	return adminService.redisClient.InitRedisClient(address, password, string(caCert), 0, adminService.tls.Enabled)
}

func generatePassword(length int) string {
	chars := []rune(passCharSet)
	var b bytes.Buffer
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func (adminService *AdministrationService) storeCredentialsAndGetEnvForRedisInstance(secretName string, passwordFromRequest string) (*v1.Secret, string, error) {
	var password string
	isVault := adminService.vaulterHelper != nil

	// If pass is not predefined, then generate it for vault or simple secret
	if passwordFromRequest != "" {
		password = passwordFromRequest
	} else {
		if isVault {
			generatedPass, err := adminService.vaulterHelper.GeneratePassword(vaultPolicy)
			if err != nil {
				return nil, "", err
			}
			password = generatedPass
		} else {
			password = generatePassword(16)
		}
	}

	// Password to return
	returnPass := password

	// Store predefined or generated pass if vault is used
	if isVault {
		err := adminService.vaulterHelper.StorePassword(secretName, password)
		if err != nil {
			return nil, "", err
		}

		// Replace password with vault url
		password = adminService.vaulterHelper.GetEnvTemplateForVault(redisPasswordConst, secretName).Value
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: adminService.namespace,
		},
		Data: map[string][]byte{
			constants.Password: []byte(password),
		},
	}

	return secret, returnPass, nil
}

func credsName(dbName string) string {
	return dbName + credsSuffix
}

func (adminService *AdministrationService) DescribeDatabases(ctx context.Context, logicalDatabases []string, showResources bool, showConnections bool) map[string]dao.LogicalDatabaseDescribed {
	logger := utils.AddLoggerContext(adminService.logger, ctx)
	describedLogicalDbs := make(map[string]dao.LogicalDatabaseDescribed)
	for _, service := range logicalDatabases {
		describedLogicalDb := dao.LogicalDatabaseDescribed{ConnectionProperties: nil, Resources: nil}
		if showResources {
			mapping := adminService.getResourcesMapping(service)
			var foundResources []dao.DbResource
			for kind, object := range mapping {
				err := adminService.kubeClient.Get(
					ctx,
					types.NamespacedName{
						Namespace: adminService.namespace,
						Name:      object.name,
					},
					object.object)
				if err != nil {
					if !errors.IsNotFound(err) {
						logger.Warn(fmt.Sprintf("Failed listing %s object for service %s: %+v", kind, service, err))
					}
				} else {
					foundResources = append(foundResources, dao.DbResource{
						Kind: kind,
						Name: object.name,
					})
				}
			}
			describedLogicalDb.Resources = foundResources
			logger.Info(fmt.Sprintf("Resources of database: %+v", describedLogicalDb.Resources))
		}
		if showConnections {
			password := adminService.readRedisDBPassword(ctx, service)
			conn := createConnectionProperties(service, password, adminService.namespace, adminService.redisServicePort)
			describedLogicalDb.ConnectionProperties = conn
		}
		describedLogicalDbs[service] = describedLogicalDb
	}
	return describedLogicalDbs
}

func RedisMapConfigToString(redisConfig map[string]interface{}) (configString string) {
	delimiter := ""
	for key, value := range redisConfig {
		configString += delimiter + key + " " + fmt.Sprintf("%v", value)
		delimiter = "\n"
	}
	return configString
}

func GetRedisDefaultConfigMap(kubeClient client.Client, namespace string, log *zap.Logger) map[string]interface{} {
	configMapFromCloud := &v1.ConfigMap{}
	err := kubeClient.Get(context.TODO(),
		types.NamespacedName{Name: RedisDefaultConfigMapName, Namespace: namespace}, configMapFromCloud)
	core.PanicError(err, log.Error, "Redis configuration listing failed")

	configBytesIn := configMapFromCloud.Data["config"]
	configFromCloud := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(configBytesIn), &configFromCloud)
	core.HandleError(err, log.Error, "Could not unmarshal passed yaml config!")

	return configFromCloud
}
